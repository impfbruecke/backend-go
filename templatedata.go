package main

// TmplData bundles the data that is passed to the html templates for easier
// and uniform acces to it. The fields which are not used in a certain template
// may be nil, it is up to the template to check for nil values where they
// might occurr
type TmplData struct {
	CurrentUser        string
	DefaultCapacity    string
	DefaultEndHour     string
	DefaultEndMinute   string
	DefaultLocation    string
	DefaultStartHour   string
	DefaultStartMinute string
	DefaultTitle       string
	AppMessages        []string
	AppMessageSuccess  string
	Calls              []Call
	CallStatus         CallStatus
	Persons            []Person
}
