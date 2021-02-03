package main

type Bridge struct {
	calls []Call
}

func NewBridge() *Bridge {
	return &Bridge{}
}

func (b *Bridge) AddCall(call Call) error {
	b.calls = append(b.calls, call)
	return nil
}

func (b *Bridge) GetActiveCalls() ([]Call, error) {
	return b.calls, nil
}

// Import imports a a person into the database
func (b *Bridge) Import(p Person) error {
	return nil
}

// ImportMulti is just a helper function to call Import for multiple persons.
// It may be used for the csv import functionality
func (b *Bridge) ImportMulti(persons []Person) error {

	for k := range persons {
		// If a single entry fails to import, abort and return
		if err := b.Import(persons[k]); err != nil {
			return err
		}
	}
	return nil
}
