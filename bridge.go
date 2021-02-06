package main

import (
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type Bridge struct {
	// TODO handle duplicates and validate data
	db *sqlx.DB
}

var schemaPersons = `
CREATE TABLE IF NOT EXISTS persons (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	phone TEXT NOT NULL,
	center_id INTEGER NOT NULL,
	group_num INTEGER NOT NULL,
	last_call INTEGER,
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
	location TEXT NOT NULL
);
`

func NewBridge() *Bridge {

	log.Info("Creating new bridge")

	// Use the path in envrionment variable if specified, default fallback to
	// ./data.db for testing
	dbPath := "./data.db"

	// Check if path is set
	if os.Getenv("IMPF_DB_FILE") != "" {
		dbPath = os.Getenv("IMPF_DB_FILE")
	}

	log.Println("Using database:", dbPath)

	// Open connection to database file. Will be created if it does not already
	// exist. Exit application on errors, we can't continue without database
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Exec the schema or fail; multi-statement Exec behavior varies between
	// database drivers

	log.Debug("Verifying DB schema for calls")
	db.MustExec(schemaCalls)

	log.Debug("Verifying DB schema for persons")
	db.MustExec(schemaPersons)

	return &Bridge{db: db}
}

func (b *Bridge) AddCall(call Call) error {

	log.Debugf("Adding call %+v\n", call)

	tx := b.db.MustBegin()
	tx.NamedExec(
		"INSERT INTO calls ( title, center_id, capacity, time_start, time_end, location) VALUES"+
			"( :title, :center_id, :capacity, :time_start, :time_end, :location)", &call)

	return tx.Commit()
}

// AddPerson adds a person to the databse
func (b *Bridge) AddPerson(person Person) error {

	log.Debugf("Adding person %+v\n", person)

	tx := b.db.MustBegin()
	if _, err := tx.NamedExec(
		"INSERT INTO persons (center_id, group_num, phone) VALUES "+
			"(:center_id, :group_num, :phone)", &person); err != nil {
		return err
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
			"INSERT INTO persons (center_id, group_num, phone) VALUES"+
				"(:center_id, :group_num, :phone)", &persons[k]); err != nil {
			return err
		}

	}
	return tx.Commit()
}

type callstatus struct {
	Call    Call
	Persons []Person
}

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
	if err = b.db.Select(&persons, "SELECT * FROM persons WHERE last_call=$1", id); err != nil {
		log.Warn("Failed to find persons for callID:", id)
		log.Warn(err)
	}
	status.Persons = persons

	return status, err
}

func (b *Bridge) GetActiveCalls() ([]Call, error) {

	log.Debug("Retrieving active calls")

	// Query the database, storing results in a []User (wrapped in []interface{})
	calls := []Call{}
	// b.db.Select(&calls, "SELECT * FROM calls ORDER BY time_start ASC")
	err := b.db.Select(&calls, "SELECT * FROM calls")
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Found calls: %+v\n", calls)
	return calls, err
}

func (b *Bridge) GetPersons() ([]Person, error) {

	log.Debug("Retrieving persons")

	// Query the database, storing results in a []User (wrapped in []interface{})
	persons := []Person{}
	// b.db.Select(&calls, "SELECT * FROM calls ORDER BY time_start ASC")
	err := b.db.Select(&persons, "SELECT * FROM persons")
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Found persons: %+v\n", persons)
	return persons, err
}

func (b *Bridge) PersonAcceptCall() error {

	log.Debug("Accepting call")

	//TODO implement
	return nil
}

func (b *Bridge) PersonCancelCall() error {

	log.Debug("Cancelling call")

	//TODO implement
	return nil
}

func (b *Bridge) PersonDelete() error {

	log.Debug("Deleting person")

	//TODO implement
	return nil
}
