package main

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Handle endpoints. GET will usually render a html template, POST will be used
// for data import and specific requests
func handlerSendCall(w http.ResponseWriter, r *http.Request) {

	// Default start time is now, default end time is in 3 hours from now.
	// If adding 3 hours to the current time results in being the next day,
	// we just set a predefined value of 23:00 - 23:59. This will probably
	// not happen during normal office hours, but developing at night
	// sometimes causes unxepeted errors ;)

	startHour, startMin, _ := time.Now().Clock()
	endHour, endMin, _ := time.Now().Add(3 * time.Hour).Clock()

	if endHour < startHour {
		startMin = 0
		startHour = 23
		endHour = 23
		endMin = 59
	}

	tData := TmplData{
		CurrentUser:           contextString(contextKeyCurrentUser, r),
		DefaultTitle:          "IZ Duisburg",                 // TODO add collumn to users table
		DefaultCapacity:       "10",                          // TODO add collumn to users table
		DefaultLocationName:   "Impfzentrum Duisburg am TAM", // TODO add collumn to users table
		DefaultLocationStreet: "Plessingstraße",              // TODO add collumn to users table
		DefaultHouseNumber:    "20",                          // TODO add collumn to users table
		DefaultPostCode:       "47051",                       // TODO add collumn to users table
		DefaultCity:           "Duisburg",                    // TODO add collumn to users table
		DefaultStartHour:      strconv.Itoa(startHour),
		DefaultStartMinute:    strconv.Itoa(startMin),
		DefaultEndHour:        strconv.Itoa(endHour),
		DefaultEndMinute:      strconv.Itoa(endMin),
	}

	if r.Method == http.MethodGet {

		log.Info(templates.ExecuteTemplate(w, "newCall.html", tData))

	} else if r.Method == http.MethodPost {

		// Try to create new call from input data
		r.ParseForm()
		call, errStrings, err := NewCall(r.Form)
		if err != nil {
			log.Warn(err)
			tData.AppMessages = errStrings
			log.Info(templates.ExecuteTemplate(w, "newCall.html", tData))
			return
		}

		// Add call to bridge
		if err := bridge.AddCall(call); err != nil {
			log.Warn(err)
			tData.AppMessages = []string{"Ruf konnte nicht gespeichert werden"}
			templates.ExecuteTemplate(w, "newCall.html", tData)
			return
		}

		tData.AppMessageSuccess = "Ruf erfolgreich erstellt!"
		log.Info(templates.ExecuteTemplate(w, "newCall.html", tData))

	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerActiveCalls(w http.ResponseWriter, r *http.Request) {

	callID := mux.Vars(r)["id"]
	details, err := bridge.GetCallStatus(callID)
	tData := TmplData{
		CurrentUser: contextString(contextKeyCurrentUser, r),
	}

	if err == nil {
		tData.CallStatus = details
	}

	if r.Method == http.MethodGet {

		calls, err := bridge.GetActiveCalls()
		if err != nil {
			log.Warn(err)
			io.WriteString(w, "Failed to retrieve calls")
			return
		}

		// Show all active calls
		tData.Calls = calls
		log.Info(templates.ExecuteTemplate(w, "calls.html", tData))

	} else {
		io.WriteString(w, "Invalid request")
	}
}
