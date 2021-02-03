package main

type Bridge struct {
	calls   []Call
	persons []Person
}

func NewBridge() *Bridge {
	return &Bridge{}
}

func (b *Bridge) AddCall(call Call) error {
	// TODO Replace with sql query
	b.calls = append(b.calls, call)
	return nil
}

func (b *Bridge) AddPersons(persons []Person) error {
	// TODO Replace with sql query that filters duplicates
	b.persons = append(b.persons, persons...)
	return nil
}

func (b *Bridge) GetActiveCalls() ([]Call, error) {
	// TODO Replace with sql query
	return b.calls, nil
}
