package main

import (
	"net/url"
)

type Person struct {
	CenterID int
	Name     string
	Phone    string
}

// NewPerson receives the input data and returns a slice of person objects. For
// single import this will just be an array with a single entry, for CSV upload
// it may be longer.
func NewPersons(data url.Values) ([]Person, error) {
	return []Person{}, nil
}
