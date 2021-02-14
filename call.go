package main

import (
	"errors"
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

	// Validate location and title are not empty
	location := data.Get("location")
	if location == "" {
		errorStrings = append(errorStrings, "Fehlender Ort")
	}

	title := data.Get("title")
	if title == "" {
		errorStrings = append(errorStrings, "Fehlender Titel")
	}

	if len(errorStrings) != 0 {
		return Call{}, errorStrings, errors.New("Invalid data for call")
	}

	return Call{
		Title:     title,
		CenterID:  0, //TODO set centerID from contextString
		Capacity:  capacity,
		TimeStart: timeStart,
		TimeEnd:   timeEnd,
		Location:  location,
	}, errorStrings, nil
}
