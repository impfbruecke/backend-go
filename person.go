package main

import (
	"database/sql"
	"errors"
	"strconv"
)

type Person struct {
	ID       int           `db:"id"`
	CenterID int           `db:"center_id"`
	Group    int           `db:"group_num"`
	Phone    string        `db:"phone"`
	LastCall sql.NullInt64 `db:"last_call"`
}

// NewPerson receives the input data and returns a slice of person objects. For
// single import this will just be an array with a single entry, for CSV upload
// it may be longer.
func NewPerson(centerID, group int, phone string) (Person, error) {
	person := Person{CenterID: centerID}

	if phone == "" {
		return person, errors.New("Ungültige Rufnummer: " + phone)
	}
	person.Phone = phone

	if group == 0 {
		return person, errors.New("Ungültige Gruppe: " + strconv.Itoa(group))
	}

	person.Group = group

	return person, nil
}
