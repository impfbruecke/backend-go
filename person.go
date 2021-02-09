package main

import (
	"database/sql"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
	"strconv"
)

type Person struct {
	ID               int           `db:"id"`                 // Internal ID of the person, created by DB
	CenterID         int           `db:"center_id"`          // ID of center that added this person
	Group            int           `db:"group_num"`          // Vaccination group
	Phone            string        `db:"phone"`              // Telephone number
	LastCall         sql.NullInt64 `db:"last_call"`          // ID of last call the person was called to
	LastCallAccepted sql.NullInt64 `db:"last_call_accepted"` // ID of last call the person has accepted
	Status           bool          `db:"status"`             // Vaccination status
}

// NewPerson receives the input data and returns a slice of person objects. For
// single import this will just be an array with a single entry, for CSV upload
// it may be longer.
func NewPerson(centerID, group int, phone string, status bool) (Person, error) {

	person := Person{
		CenterID:         centerID,
		LastCall:         sql.NullInt64{Valid: false},
		LastCallAccepted: sql.NullInt64{Valid: false},
		Status:           status,
	}

	num, err := libphonenumber.Parse(phone, "DE")
	if err != nil {
		log.Warn("Error parsing phone number: ", phone)
		return person, errors.New("Ungültige Rufnummer: " + phone)
	}

	person.Phone = libphonenumber.Format(num, libphonenumber.E164)
	log.Debug("parsed number: ", person.Phone)

	// Validate that group number is not empty
	if group == 0 {
		return person, errors.New("Ungültige Gruppe: " + strconv.Itoa(group))
	}

	person.Group = group

	return person, nil
}
