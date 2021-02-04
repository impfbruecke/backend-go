package main

type Person struct {
	ID       int    `db:"id"`
	CenterID int    `db:"center_id"`
	Group    int    `db:"group_num"`
	Phone    string `db:"phone"`
}

// NewPerson receives the input data and returns a slice of person objects. For
// single import this will just be an array with a single entry, for CSV upload
// it may be longer.
func NewPerson(centerID, group int, phone string) (Person, error) {
	return Person{
		CenterID: centerID,
		Phone:    phone,
		Group:    group,
	}, nil
}
