package main

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Bridge struct {
	// TODO handle duplicates and validate data
	db     *sqlx.DB
	sender *TwillioSender
}

var schemaPersons = `
CREATE TABLE IF NOT EXISTS persons (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	phone TEXT NOT NULL UNIQUE ON CONFLICT FAIL,
	center_id INTEGER NOT NULL,
	group_num INTEGER NOT NULL,
	last_call INTEGER,
	last_call_accepted INTEGER,
	status INTEGER NOT NULL
);
`
var schemaCalls = `
CREATE TABLE IF NOT EXISTS calls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	center_id INTEGER NOT NULL,
	capacity INTEGER NOT NULL,
	time_start DATETIME NOT NULL,
	time_end DATETIME NOT NULL,
	location TEXT NOT NULL,
	sent INTEGER NOT NULL
);
`

var schemaUsers = `
CREATE TABLE IF NOT EXISTS users (
  username text primary key,
  password text
);
`

func NewBridge() *Bridge {

	log.Info("Creating new bridge")
	log.Info("Using database:", dbPath)

	// Open connection to database file. Will be created if it does not already
	// exist. Exit application on errors, we can't continue without database
	db, err := sqlx.Connect("sqlite3", dbPath)

	// Only required because of a bug with sqlx and sqlite.
	// TODO remove when migrating to postgresql if performance is too bad
	db.SetMaxOpenConns(1)
	if err != nil {
		log.Fatal(err)
	}

	// Exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers

	log.Debug("Verifying DB schema for calls")
	db.MustExec(schemaCalls)

	log.Debug("Verifying DB schema for persons")
	db.MustExec(schemaPersons)

	log.Debug("Verifying DB schema for users")
	db.MustExec(schemaUsers)

	sender := NewTwillioSender(
		os.Getenv("IMPF_TWILIO_API_ENDPOINT"),
		os.Getenv("IMPF_TWILIO_API_USER"),
		os.Getenv("IMPF_TWILIO_API_PASS"),
		os.Getenv("IMPF_TWILIO_API_FROM"),
	)

	bridge := Bridge{db: db, sender: sender}

	// 15-minute timer/ticker tor periodically do stuff
	ticker := time.NewTicker(15 * time.Minute)
	quit := make(chan struct{})

	go func() {
		// Initial run when the ticker starts so we don't have to wait until
		// the ticker on first start
		bridge.SendNotifications()
		bridge.DeleteOldCalls()
		for {
			select {
			case <-ticker.C:
				bridge.SendNotifications()
				bridge.DeleteOldCalls()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return &bridge
}

// DeleteOldCalls finds calls for which the end_time has passed and deletes
// them from the db
func (b Bridge) DeleteOldCalls() {

	m := map[string]interface{}{"now": time.Now()}
	result, err := b.db.NamedExec("DELETE FROM calls WHERE time_end < :now", m)

	if err != nil {
		log.Error("Error deleting calls: ", err)
		return
	}

	numrows, err := result.RowsAffected()
	if err != nil {
		log.Error(err)
	}

	log.Info("Old calls deleted: ", numrows)
}

// SendNotifications checks for all active calls, if any notifications have to
// be send. If so it finds the next persons to send notifications to and
// triggers the notificaion mechanism
func (b Bridge) SendNotifications() {
	log.Debug("Timer reached, sending notifications to calls")

	// Get all calls where end-time has not been reached yet
	calls, err := b.GetActiveCalls()

	if err != nil {
		log.Error(err)
	}

	// For each call
	for _, v := range calls {

		// Check how many people have accepted
		acceptedPersons, err := b.GetAcceptedPersons(v.ID)

		if err != nil {
			log.Error(err)
			return
		}

		// if less then capacity, send notifications
		if v.Capacity > len(acceptedPersons) {
			err := b.NotifyCall(v.ID, v.Capacity-len(acceptedPersons))
			if err != nil {
				log.Error(err)
			}
		}
	}
}

// NotifyCall is given a call ID and a maximum number of persons to notify. It
// sends notificaions for that call to the amount of persons specified or less
// if there are no persons to notify left
func (b *Bridge) NotifyCall(id, numPersons int) error {
	// TODO
	persons, err := b.GetNextPersonsForCall(numPersons, id)

	call, err := b.GetCallStatus(strconv.Itoa(id))

	if err != nil {
		return err
	}

	for k := range persons {
		if err := b.sender.SendMessageNotify(
			persons[k].Phone,
			call.Call.TimeStart.Format("14:12"),
			call.Call.TimeEnd.Format("14:12"),
			call.Call.Location,
		); err != nil {
			log.Error(err)
		}
	}

	return nil
}

// AddCall adds a call to the database
func (b *Bridge) AddCall(call Call) error {

	log.Debugf("Adding call %+v\n", call)

	tx := b.db.MustBegin()
	res, err := tx.NamedExec(
		"INSERT INTO calls ( title, center_id, capacity, time_start, time_end, location, sent) VALUES"+
			"( :title, :center_id, :capacity, :time_start, :time_end, :location, :sent)", &call)

	rows, _ := res.RowsAffected()
	log.Debugf("%v rows affected\n", rows)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// AddPerson adds a person to the databse
func (b *Bridge) AddPerson(person Person) error {

	log.Debugf("Adding person %+v\n", person)

	var res sql.Result
	var err error

	tx := b.db.MustBegin()
	if res, err = tx.NamedExec(
		"INSERT INTO persons (center_id, group_num, phone, last_call, last_call_accepted, status) VALUES "+
			"(:center_id, :group_num, :phone, :last_call, :last_call_accepted, :status)", &person); err != nil {
		return err
	}

	numrows, err := res.RowsAffected()
	log.Debugf("Persons added: %v\n", numrows)

	return tx.Commit()
}

// AddPersons adds multiple persons at once. We don't just reuse the
// AddPerson() function here, to optimize performance. Named transactions are
// created for each person, but the commit is only done once
func (b *Bridge) AddPersons(persons []Person) error {

	log.Debugf("Adding persons: %+v\n ", persons)

	tx := b.db.MustBegin()
	for k := range persons {
		if _, err := tx.NamedExec(
			"INSERT INTO persons (center_id, group_num, phone, last_call, last_call_accepted, status) VALUES "+
				"(:center_id, :group_num, :phone, :last_call, :last_call_accepted, :status)", &persons[k]); err != nil {
			return err
		}

	}
	return tx.Commit()
}

type callstatus struct {
	Call    Call
	Persons []Person
}

// GetCallStatus returns the status of a call. This is used for the status
// template to bundle information about the call and the persons that have
// accepted it
func (b *Bridge) GetCallStatus(id string) (callstatus, error) {

	var err error
	status := callstatus{}

	//Retrieve call information
	call := Call{}
	if err = b.db.Get(&call, "SELECT * FROM calls WHERE id=$1", id); err != nil {
		log.Warn("Failed to find call with callID:", id)
		log.Warn(err)
	}
	status.Call = call

	// Retrieve persons notified for that call
	persons := []Person{}
	if err = b.db.Select(&persons, "SELECT * FROM persons WHERE last_call_accepted=$1", id); err != nil {
		log.Warn("Failed to find persons for callID:", id)
		log.Warn(err)
	}
	status.Persons = persons

	return status, err
}

// GetActiveCalls returns a list of active calls (time_end > now)
func (b *Bridge) GetActiveCalls() ([]Call, error) {

	log.Debugf("Retrieving active calls with end time before %s", time.Now())

	// Query the database, storing results in a []User (wrapped in []interface{})
	calls := []Call{}
	// b.db.Select(&calls, "SELECT * FROM calls ORDER BY time_start ASC")j
	err := b.db.Select(&calls, "SELECT * FROM calls where time_end > $1", time.Now())

	if err != nil {
		log.Error(err)
		return calls, err
	}

	log.Debugf("Found calls: %+v\n", calls)
	return calls, err
}

// GetNextPersonsForCall finds the next `num` persons that should be notified
// for a callID. Selection is based on group_num
func (b *Bridge) GetNextPersonsForCall(num, callID int) ([]Person, error) {

	log.Debugf("Retrieving next persons %v for call ID: %v\n", num, callID)

	persons := []Person{}
	err := b.db.Select(&persons, "SELECT * FROM persons WHERE last_call!=$1 OR last_call IS NULL ORDER BY group_num ASC LIMIT $2", callID, num)
	if err != nil {
		log.Error(err)
		return persons, err
	}

	log.Debugf("Found persons: %+v\n", persons)
	return persons, err
}

// GetAcceptedPersons returns all persons that have accepted a call
func (b *Bridge) GetAcceptedPersons(id int) ([]Person, error) {

	log.Debugf("Retrieving accepted persons for call ID: %v\n", id)

	persons := []Person{}
	err := b.db.Select(&persons, "SELECT * FROM persons where last_call_accepted=$1", id)
	if err != nil {
		log.Error(err)
		return persons, err
	}

	log.Debugf("Found persons: %+v\n", persons)
	return persons, err
}

// GetPersons gets all persons currently in the database.
func (b *Bridge) GetPersons() ([]Person, error) {

	log.Debug("Retrieving persons")

	// Query the database, storing results in a []User (wrapped in []interface{})
	persons := []Person{}
	// b.db.Select(&calls, "SELECT * FROM calls ORDER BY time_start ASC")
	err := b.db.Select(&persons, "SELECT * FROM persons")
	if err != nil {
		log.Error(err)
		return persons, err
	}

	log.Debugf("Found persons: %+v\n", persons)
	return persons, err
}

func (b *Bridge) PersonAcceptCall(phoneNumber string) error {

	log.Debugf("number %s trying to accept call\n", phoneNumber)
	var err error

	// Get the last call the person was called to
	lastCallOfPerson := Call{}
	err = b.db.Get(&lastCallOfPerson, "SELECT * FROM persons WHERE phone=$1", phoneNumber)

	// Check if the call is full
	acceptedPersons, err := b.GetAcceptedPersons(lastCallOfPerson.ID)

	if lastCallOfPerson.Capacity > len(acceptedPersons) {

		log.Debugf("Accepting number %s for call \n", phoneNumber)

		if err = b.sender.SendMessageAccept(
			phoneNumber,
			lastCallOfPerson.TimeStart.Format("14:12"),
			lastCallOfPerson.TimeEnd.Format("14:12"),
			lastCallOfPerson.Location,
			genOTP(phoneNumber, lastCallOfPerson.ID),
		); err != nil {
			return err
		}

		_, err = bridge.db.NamedExec(
			`UPDATE persons SET last_call_accepted = last_call WHERE phone=:phone`,
			map[string]interface{}{
				"phone": phoneNumber,
			},
		)
	} else {
		log.Debugf("number %s rejected for call (is full)\n", phoneNumber)
		b.sender.SendMessageReject(phoneNumber)
	}

	return err
}

func (b *Bridge) PersonCancelCall(phoneNumber string) error {

	log.Debugf("Cancelling call for number %s\n", phoneNumber)

	_, err := bridge.db.NamedExec(
		`UPDATE persons SET last_call_accepted = NULL WHERE phone=:phone`,
		map[string]interface{}{
			"phone": phoneNumber,
		},
	)

	return err
}

func (b *Bridge) PersonDelete(phoneNumber string) error {

	log.Debugf("Deleting number %s\n", phoneNumber)

	m := map[string]interface{}{"phone": phoneNumber}
	result, err := b.db.NamedExec("DELETE FROM persons WHERE phone=:phone", m)

	if err != nil {
		return err
	}

	numrows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	b.sender.SendMessageDelete(phoneNumber)

	log.Info("Number of persons deleted: ", numrows)
	return nil
}
