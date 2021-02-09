package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Call struct {
	ID        int       `db:"id"`
	Title     string    `db:"title"`
	CenterID  int       `db:"center_id"`
	Capacity  int       `db:"capacity"`
	TimeStart time.Time `db:"time_start"`
	TimeEnd   time.Time `db:"time_end"`
	Location  string    `db:"location"`
	Sent      bool      `db:"sent"`
}

func parseInputTime(h, m string) (time.Time, error) {
	log.Debugf("Parsing time hours: %s\nminutes:%s\n", h, m)
	return time.Parse("15:04", h+":"+m)
}

func NewCall(data url.Values) (Call, error) {

	call := Call{}

	// Validate capacity > 0
	capacity, err := strconv.Atoi(data.Get("capacity"))
	if err != nil || capacity < 1 {
		log.Warn("Invalid start time:", err)
		return call, err
	}

	if capacity < 1 {
		log.Warn("Capacity has to be at least 1")
		return call, errors.New("Capacity should be > 0")
	}

	// Validate start and end times make sense
	timeStart, err := parseInputTime(data.Get("start-hour"), data.Get("start-min"))
	if err != nil {
		log.Warn("Invalid start time:", err)
		return call, err
	}

	timeEnd, err := parseInputTime(data.Get("end-hour"), data.Get("end-min"))
	if err != nil {
		log.Warn("Invalid end time", err)
		return call, err
	}

	if !timeStart.Before(timeEnd) {
		log.Warn("Capacity has to be at least 1")
		return call, errors.New(fmt.Sprintf("Start time %v is not before end time %v\n", timeStart.String(), timeEnd.String()))
	}

	// Validate location and title are not empty
	location := data.Get("location")
	if location == "" {
		log.Warn("Invalid location")
		return call, errors.New("Empty location now allowed")
	}

	title := data.Get("title")
	if title == "" {
		log.Warn("Invalid title")
		return call, errors.New("Empty title now allowed")
	}

	return Call{
		Title:     title,
		CenterID:  0, //TODO set centerID from contextString
		Capacity:  capacity,
		TimeStart: timeStart,
		TimeEnd:   timeEnd,
		Location:  location,
		Sent:      false,
	}, nil
}
