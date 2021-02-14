package main

import (
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// Call represents a call issued by a user. It contains data on the
// availability, times and location. Invitations will then be send out for that
// call
type Call struct {
	ID         int       `db:"id"`
	Title      string    `db:"title"`
	CenterID   int       `db:"center_id"`
	Capacity   int       `db:"capacity"`
	TimeStart  time.Time `db:"time_start"`
	TimeEnd    time.Time `db:"time_end"`
	LocName    string    `db:"loc_name"`
	LocStreet  string    `db:"loc_street"`
	LocHouseNr string    `db:"loc_housenr"`
	LocPLZ     string    `db:"loc_plz"`
	LocCity    string    `db:"loc_city"`
	LocOpt     string    `db:"loc_opt"`
}

func todayAt(input string) (time.Time, error) {

	now := time.Now()

	year, month, day := now.Date()

	tmp, err := time.Parse("15:4", input)
	if err != nil {
		return now, err
	}

	hour, min, _ := tmp.Clock()
	return time.Date(year, month, day, hour, min, 0, 0, now.Location()), nil
}

// NewCall creates a new call
func NewCall(data url.Values) (Call, []string, error) {

	var errorStrings []string

	// Validate capacity > 0
	capacity, err := strconv.Atoi(data.Get("capacity"))
	if err != nil || capacity < 1 {
		errorStrings = append(errorStrings, "Ungültige Kapazität")
	}

	// Validate start and end times make sense
	log.Debug("start-time: ", data.Get("start-time"))
	log.Debug("end-time: ", data.Get("end-time"))

	timeStart, err := todayAt(data.Get("start-time"))
	if err != nil {
		errorStrings = append(errorStrings, "Ungültige Startzeit")
	}

	timeEnd, err := todayAt(data.Get("end-time"))
	if err != nil {
		errorStrings = append(errorStrings, "Ungültige Endzezeit")
	}

	if !timeStart.Before(timeEnd) {
		errorStrings = append(errorStrings, "Endzezeit ist nicht nach Startzeit")
	}

	// Get text fields and check that they are not empty strings
	var locName, locStreet, locHouseNr, locPlz, locCity, locOpt, title string

	locName, errorStrings = getFormFieldWithErrors(data, "loc_name", errorStrings)
	locStreet, errorStrings = getFormFieldWithErrors(data, "loc_street", errorStrings)
	locHouseNr, errorStrings = getFormFieldWithErrors(data, "loc_housener", errorStrings)
	locPlz, errorStrings = getFormFieldWithErrors(data, "loc_plz", errorStrings)
	locCity, errorStrings = getFormFieldWithErrors(data, "loc_city", errorStrings)
	locOpt, errorStrings = getFormFieldWithErrors(data, "loc_opt", errorStrings)
	title, errorStrings = getFormFieldWithErrors(data, "title", errorStrings)

	return Call{
		Title:      title,
		CenterID:   0, //TODO set centerID from contextString
		Capacity:   capacity,
		TimeStart:  timeStart,
		TimeEnd:    timeEnd,
		LocName:    locName,
		LocStreet:  locStreet,
		LocHouseNr: locHouseNr,
		LocPLZ:     locPlz,
		LocCity:    locCity,
		LocOpt:     locOpt,
	}, errorStrings, nil
}

func getFormFieldWithErrors(data url.Values, formID string, errorStrings []string) (string, []string) {

	value := data.Get(formID)
	if value == "" {
		errorStrings = append(errorStrings, "Ungültige Eingabe für: "+formID)
	}

	return value, errorStrings
}
