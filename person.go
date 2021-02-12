package main

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/ttacon/libphonenumber"
	"strconv"
)

// Person represents a person that has been imported to be notified for calls.
// It includes the phone number aswell as the user id that imported information
// about the Person itself
type Person struct {
	Phone    string `db:"phone"`     // Telephone number
	CenterID int    `db:"center_id"` // ID of center that added this person
	Group    int    `db:"group_num"` // Vaccination group
	Status   bool   `db:"status"`    // Vaccination status
}

// NewPerson receives the input data and returns a slice of person objects. For
// single import this will just be an array with a single entry, for CSV upload
// it may be longer.
func NewPerson(centerID, group int, phone string, status bool) (Person, error) {

	person := Person{
		CenterID: centerID,
		Status:   status,
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
