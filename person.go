package main

import (
	"database/sql"
	"errors"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Person struct {
	ID       int           `db:"id"`        // Internal ID of the person, created by DB
	CenterID int           `db:"center_id"` // ID of center that added this person
	Group    int           `db:"group_num"` // Vaccination group
	Phone    string        `db:"phone"`     // Telephone number
	LastCall sql.NullInt64 `db:"last_call"` // ID of last call the person was called to
	Status   bool          `db:"status"`    // Vaccination status
}

// NewPerson receives the input data and returns a slice of person objects. For
// single import this will just be an array with a single entry, for CSV upload
// it may be longer.
func NewPerson(centerID, group int, phone string, status bool) (Person, error) {
	person := Person{
		CenterID: centerID,
		LastCall: sql.NullInt64{},
		Status:   status,
	}

	// Validate that phone number is not empty
	if phone == "" {
		return person, errors.New("Ungültige Rufnummer: " + phone)
	}
	person.Phone = phone

	// Validate that group number is not empty
	if group == 0 {
		return person, errors.New("Ungültige Gruppe: " + strconv.Itoa(group))
	}

	person.Group = group

	return person, nil
}

func (p *Person) Notify(text string) error {

	// TODO implement
	// TODO send actual SMS, for now we just print to the log
	log.Debugf("Sending SMS to: %v\n%v\n\n", p.Phone, text)

	_, err := bridge.db.NamedExec(
		`UPDATE persons SET last_notified = :last_notified WHERE
		phone = ':phone'`,
		map[string]interface{}{
			"last_notified": time.Now(),
			"phone":         p.Phone,
		},
	)

	return err
}
