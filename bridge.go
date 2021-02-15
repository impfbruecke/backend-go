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

// Bridge is the main struct of the application. It abstracts the connection to
// the database and twilio providing methods to act on both
type Bridge struct {
	// TODO handle duplicates and validate data
	db     *sqlx.DB
	sender *TwillioSender
}

var schemaPersons = `
CREATE TABLE IF NOT EXISTS persons (
	phone TEXT PRIMARY KEY,
	center_id INTEGER NOT NULL,
	group_num INTEGER NOT NULL,
	status INTEGER NOT NULL,
	age INTEGER NOT NULL
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
	age_min INTEGER NOT NULL,
	age_max INTEGER NOT NULL,
	loc_name TEXT NOT NULL,
	loc_street TEXT NOT NULL,
	loc_housenr TEXT NOT NULL,
	loc_plz TEXT NOT NULL,
	loc_city TEXT NOT NULL,
	loc_opt TEXT NOT NULL
);
`

var schemaUsers = `
CREATE TABLE IF NOT EXISTS users (
  username text primary key,
  password text
);
`

var schemaNotifications = `
CREATE TABLE IF NOT EXISTS invitations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  phone TEXT NOT NULL,
  call_id INTEGER NOT NULL,
  status TEXT NOT NULL,
  time DATETIME NOT NULL
);
`

// NewBridge creates a new instance of the bridge using predifined parameters
// from env vars, global vars and/or defaults
func NewBridge() *Bridge {

	log.Info("Creating new bridge")

	log.Info("Using database:", os.Getenv("IMPF_DB_FILE"))

	// Open connection to database file. Will be created if it does not already
	// exist. Exit application on errors, we can't continue without database
	db, err := sqlx.Connect("sqlite3", os.Getenv("IMPF_DB_FILE"))

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

	log.Debug("Verifying DB schema for notifications")
	db.MustExec(schemaNotifications)

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
	var err error
	var persons []Person
	var call CallStatus

	if persons, err = b.GetNextPersonsForCall(numPersons, id); err != nil {
		return err
	}

	if call, err = b.GetCallStatus(strconv.Itoa(id)); err != nil {
		return err
	}

	if err != nil {
		return err
	}

	for k := range persons {
		if err := b.sender.SendMessageNotify(
			persons[k].Phone,
			call.Call.TimeStart.Format("14:12"),
			call.Call.TimeEnd.Format("14:12"),
			call.Call.LocName,
			call.Call.LocStreet,
			call.Call.LocHouseNr,
			call.Call.LocPLZ,
			call.Call.LocCity,
			call.Call.LocOpt,
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
		`INSERT INTO calls (
			title,
			center_id,
			capacity,
			time_start,
			time_end,
			age_min,
			age_max,
			loc_name,
			loc_street,
			loc_housenr,
			loc_plz,
			loc_city,
			loc_opt
		) VALUES (
			:title,
			:center_id,
			:capacity,
			:time_start,
			:time_end,
			:age_min,
			:age_max,
			:loc_name,
			:loc_street,
			:loc_housenr,
			:loc_plz,
			:loc_city,
			:loc_opt
		)`, &call)

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
		"INSERT INTO persons (center_id, group_num, phone, status, age) VALUES "+
			"(:center_id, :group_num, :phone, :status, :age)", &person); err != nil {
		return err
	}

	if numrows, err := res.RowsAffected(); err != nil {
		return err
	} else {
		log.Debugf("Persons will be added: %v\n", numrows)
	}

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
			"INSERT INTO persons (center_id, group_num, phone, status, age) VALUES "+
				"(:center_id, :group_num, :phone, :status, :age)", &persons[k]); err != nil {
			return err
		}

	}
	return tx.Commit()
}

// CallStatus bundles a call and the persons who have accepted it for simpler
// rendering in the html templates. TODO see if we can just replace it with the
// existing structs
type CallStatus struct {
	Call    Call
	Persons []Person
}

// GetCallStatus returns the status of a call. This is used for the status
// template to bundle information about the call and the persons that have
// accepted it
func (b *Bridge) GetCallStatus(id string) (CallStatus, error) {

	var err error
	status := CallStatus{}

	//Retrieve call information
	call := Call{}
	if err = b.db.Get(&call, "SELECT * FROM calls WHERE id=$1", id); err != nil {
		log.Warn("Failed to find call with callID:", id)
		log.Warn(err)
	}

	status.Call = call
	status.Persons, err = b.GetAcceptedPersons(call.ID)
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

// GetAllCalls returns a list of all calls
func (b *Bridge) GetAllCalls() ([]Call, error) {

	// Query the database, storing results in a []User (wrapped in []interface{})
	calls := []Call{}
	// b.db.Select(&calls, "SELECT * FROM calls ORDER BY time_start ASC")j
	err := b.db.Select(&calls, "SELECT * FROM calls")

	if err != nil {
		log.Error(err)
		return calls, err
	}

	log.Debugf("Found calls: %+v\n", calls)
	return calls, err
}

// GetNextPersonsForCall finds the next `num` persons that should be notified
// for a callID. Selection is based on group_num
//TODO FIXME
func (b *Bridge) GetNextPersonsForCall(num, callID int) ([]Person, error) {

	// Selection is based on:
	// - group_num (lower first)
	// - random from group_num

	// Get all groups

	log.Debugf("Retrieving next persons %v for call ID: %v\n", num, callID)
	var err error
	var call CallStatus

	call, err = bridge.GetCallStatus(strconv.Itoa(callID))
	if err != nil {
		return []Person{}, err
	}

	persons := []Person{}
	err = b.db.Select(&persons,
		`SELECT * FROM persons
			WHERE phone NOT IN (
				SELECT phone FROM invitations
					WHERE call_id=$1
				)
			AND age<=$2
			AND age>=$3
			AND status=0
			ORDER BY group_num LIMIT $4`,
		callID, call.Call.AgeMax, call.Call.AgeMin, num)
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
	// err := b.db.Select(&persons, "SELECT * FROM persons where last_call_accepted=$1", id)
	err := b.db.Select(&persons,
		`SELECT persons.* FROM persons
		JOIN invitations ON
		persons.phone = invitations.phone
		AND invitations.status = 'accepted'
		AND invitations.call_id=$1`, id)
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

type Invitation struct {
	ID     int              `db:"id"`
	Phone  string           `db:"phone"`
	CallID int              `db:"call_id"`
	Status invitationStatus `db:"status"`
	Time   time.Time        `db:"time"`
}

type invitationStatus string

const (
	InvitationRejected  invitationStatus = "rejected"
	InvitationAccepted  invitationStatus = "accepted"
	InvitationCancelled invitationStatus = "cancelled"
	InvitationNotified  invitationStatus = "notified"
)

func (b *Bridge) GetInvitations() ([]Invitation, error) {

	log.Debug("Retrieving invitations")

	invitations := []Invitation{}
	err := b.db.Select(&invitations, "SELECT * FROM invitations ORDER BY time DESC")
	if err != nil {
		log.Error(err)
		return invitations, err
	}

	// log.Debugf("Found invitations: %+v\n", invitations)
	return invitations, err
}

// CallFull is true if the call is full. Used to check call status
func (b *Bridge) CallFull(call Call) (bool, error) {

	var numAccpets int
	err := b.db.Get(&numAccpets, "select count(id) from invitations where call_id=$1 and status='accepted'", call.ID)
	log.Debugf("Checking full status for call %v: %v/%v", call.ID, numAccpets, call.Capacity)
	return numAccpets >= call.Capacity, err
}

// LastCallNotified retrieves the last call a person was notified to
func (b *Bridge) LastCallNotified(person Person) (Call, error) {
	lastCallOfPerson := Call{}
	err := b.db.Get(&lastCallOfPerson,
		`SELECT * FROM calls
		WHERE id = (
			SELECT call_id FROM invitations
			WHERE phone=$1
			ORDER BY time DESC
		)`, person.Phone)
	return lastCallOfPerson, err
}

// PersonAcceptLastCall retrieves the last call a person has accepted
func (b *Bridge) PersonAcceptLastCall(phoneNumber string) error {

	// "update invitations set status = \"accepted\" where phone=$1 and id in ( select id from ( select id from invitations order by time desc limit 1) tmp )", phoneNumber)

	log.Debugf("number %s trying to accept call\n", phoneNumber)

	var err error
	var lastCall Call
	var isFull bool

	if lastCall, err = b.LastCallNotified(Person{Phone: phoneNumber}); err != nil {
		log.Debugf("Phone %s has not been invited yet\n", phoneNumber)
		return err
	}

	if isFull, err = b.CallFull(lastCall); err != nil {
		log.Debugf("Last call %v, does not exist\n", lastCall.ID)
		return err
	}

	if isFull {
		log.Debugf("Rejecting number %s for call (is full)\n", phoneNumber)
		_, err = bridge.db.NamedExec(
			`UPDATE invitations SET
				status=:status,
				time=:time
			WHERE
				phone=:phone
				AND call_id=:call_id
				AND status=:oldstatus`,
			map[string]interface{}{
				"status":    InvitationAccepted,
				"oldstatus": InvitationRejected,
				"phone":     phoneNumber,
				"time":      time.Now(),
				"call_id":   lastCall.ID,
			},
		)

		if err != nil {
			log.Errorf("Failed to set accepted status for last invitation of %s\n", phoneNumber)
			return err
		}

		if err := b.sender.SendMessageReject(phoneNumber); err != nil {
			log.Errorf("Failed to send reject message for phone %s\n", phoneNumber)
			log.Error(err)
		}
	} else {

		log.Debugf("Accepting number %s for call \n", phoneNumber)

		log.Debugf("Setting status=accepted for phone %s\n", phoneNumber)

		var res sql.Result
		res, err = bridge.db.NamedExec(
			`UPDATE invitations SET
				status=:status,
				time=:time
			WHERE
				phone=:phone
				AND call_id=:call_id
				AND status=:oldstatus`,
			map[string]interface{}{
				"status":    InvitationAccepted,
				"oldstatus": InvitationNotified,
				"phone":     phoneNumber,
				"time":      time.Now(),
				"call_id":   lastCall.ID,
			},
		)

		if err != nil {
			log.Errorf("Failed to set accepted status for last invitation of %s\n", phoneNumber)
			return err
		}

		rowNum, err := res.RowsAffected()

		if err != nil {
			log.Error("Failed to get number of affected invitations")
			return err
		}

		log.Debugf("Updated %v invitations\n", rowNum)

		if rowNum == 0 {
			log.Warn("No invitations updated, call might have been already accepted")
		} else {

			log.Debugf("Sending accept message to phone %s\n", phoneNumber)

			if err = b.sender.SendMessageAccept(
				phoneNumber,
				lastCall.TimeStart.Format("14:12"),
				lastCall.TimeEnd.Format("14:12"),
				lastCall.LocName,
				lastCall.LocStreet,
				lastCall.LocHouseNr,
				lastCall.LocPLZ,
				lastCall.LocCity,
				lastCall.LocOpt,
				genOTP(phoneNumber, lastCall.ID),
			); err != nil {
				log.Errorf("Failed to send accept message for phone %s: %v\n", phoneNumber, err)
				return err
			}
		}
	}

	return err
}

// PersonCancelAllCalls cancels all accepted calls
func (b *Bridge) PersonCancelAllCalls(phoneNumber string) error {

	log.Debugf("Cancelling call for number %s\n", phoneNumber)

	_, err := bridge.db.NamedExec(
		`UPDATE invitations SET status=:newstatus, time=:time WHERE phone=:phone AND status=:oldstatus`,

		map[string]interface{}{
			"phone":     phoneNumber,
			"oldstatus": InvitationAccepted,
			"newstatus": InvitationCancelled,
			"time":      time.Now(),
		},
	)

	return err
}

// PersonDelete removes a person from the imported data
// TODO we shoud keep some kind of reference of the person, so that it won't be
// reimported
func (b *Bridge) PersonDelete(phoneNumber string) error {

	log.Debugf("Deleting number %s\n", phoneNumber)

	m := map[string]interface{}{"phone": phoneNumber}
	result, err := b.db.NamedExec("DELETE FROM persons WHERE phone=:phone", m)

	if err != nil {
		log.Warnf("Phone %s not deleted %v\n", phoneNumber, err)
		return err
	}

	numrows, err := result.RowsAffected()
	if err != nil {
		log.Warnf("Failed to get rows affected by deletion %v\n", err)
		return err
	}

	if err := b.sender.SendMessageDelete(phoneNumber); err != nil {
		log.Warnf("Failed to send deletion confirmation to %s: %v\n", phoneNumber, err)
		return err
	}

	log.Info("Number of persons deleted: ", numrows)
	return nil
}
