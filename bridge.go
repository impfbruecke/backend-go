package main

import (
	"log"
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
	phone TEXT,
	group_num INTEGER,
	last_call INTEGER
);
`

var schemaCalls = `
CREATE TABLE IF NOT EXISTS calls (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT,
	center_id INTEGER,
	capacity INTEGER,
	time_start INTEGER,
	time_end INTEGER,
	location TEXT
);
`

func NewBridge() *Bridge {

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
	db.MustExec(schemaCalls)
	db.MustExec(schemaPersons)

	return &Bridge{db: db}
}

func (b *Bridge) AddCall(call Call) error {

	tx := b.db.MustBegin()
	tx.NamedExec(
		"INSERT INTO calls ( title, center_id, capacity, time_start, time_end, location) VALUES"+
			"( :title, :center_id, :capacity, :time_start, :time_end, :location)", &call)

	return tx.Commit()
}

// AddPerson adds a person to the databse
func (b *Bridge) AddPerson(person Person) error {
	tx := b.db.MustBegin()
	tx.NamedExec(
		"INSERT INTO persons (center_id, group_num, phone) VALUES "+
			"(:center_id, :group_num, :phone)", &person)

	return tx.Commit()
}

// AddPersons adds multiple persons at once. We don't just reuse the
// AddPerson() function here, to optimize performance. Named transactions are
// created for each person, but the commit is only done once
func (b *Bridge) AddPersons(persons []Person) error {
	tx := b.db.MustBegin()
	for k := range persons {
		tx.NamedExec(
			"INSERT INTO persons (center_id, group_num, phone) VALUES"+
				"(:center_id, :group_num, :phone)", &persons[k])

	}
	return tx.Commit()
}

func (b *Bridge) GetActiveCalls() ([]Call, error) {
	// Query the database, storing results in a []User (wrapped in []interface{})
	calls := []Call{}
	b.db.Select(&calls, "SELECT * FROM calls ORDER BY time_start ASC")
	return calls, nil
}
