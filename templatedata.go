package main

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
