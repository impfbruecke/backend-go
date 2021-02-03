package main

import (
	"log"
	"net/url"
	"strconv"
	"time"
)

type Call struct {
	Title     string
	CenterID  int
	Capacity  int
	TimeStart time.Time
	TimeEnd   time.Time
	Location  string
}

func parseInputTime(h, m string) (time.Time, error) {
	currentDateString := time.Now().Format("2006-01-02")
	return time.Parse(time.RFC3339, currentDateString+"T"+h+":"+m+":00Z")
}

func NewCall(data url.Values) (Call, error) {

	call := Call{}

	title := data.Get("title")

	capacity, err := strconv.Atoi(data.Get("capacity"))
	if err != nil {
		return call, err
	}

	timeStart, err := parseInputTime(data.Get("start-hour"), data.Get("start-min"))
	if err != nil {
		log.Println("Invalid start time:", err)
		return call, err
	}

	timeEnd, err := parseInputTime(data.Get("end-hour"), data.Get("end-min"))
	if err != nil {
		log.Println("Invalid end time", err)
		return call, err
	}

	location := data.Get("location")

	call.Title = title
	call.CenterID = 0 // TODO
	call.Capacity = capacity
	call.TimeStart = timeStart
	call.TimeEnd = timeEnd
	call.Location = location

	// TODO Validate data
	return call, nil

}
