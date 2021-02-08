package main

import (
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// Handle endpoints. GET will usually render a html template, POST will be used
// for data import and specific requests
func handlerSendCall(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Show the "Send Call" page

		startHour, startMin, _ := time.Now().Clock()
		endHour, endMin, _ := time.Now().Add(3 * time.Hour).Clock()

		if endHour < startHour {
			endHour = 23
		}

		data := struct {
			CurrentUser        string
			DefaultCapacity    string
			DefaultEndHour     string
			DefaultEndMinute   string
			DefaultLocation    string
			DefaultStartHour   string
			DefaultStartMinute string
			DefaultTitle       string
		}{
			CurrentUser:        contextString("current_user", r),
			DefaultTitle:       "Ruf IZ Duisburg",
			DefaultCapacity:    "10",
			DefaultLocation:    "Somewhere over the rainbow",
			DefaultStartHour:   strconv.Itoa(startHour),
			DefaultStartMinute: strconv.Itoa(startMin),
			DefaultEndHour:     strconv.Itoa(endHour),
			DefaultEndMinute:   strconv.Itoa(endMin),
		}

		log.Info(templates.ExecuteTemplate(w, "call.html", data))

	} else if r.Method == http.MethodPost {

		// Try to create new call from input data
		r.ParseForm()
		call, err := NewCall(r.Form)
		if err != nil {
			log.Warn(err)
			templates.ExecuteTemplate(w, "error.html", "Eingaben ungültig, Ruf wurde nicht erstellt")
			return
		}

		// Add call to bridge
		if err := bridge.AddCall(call); err != nil {
			log.Warn(err)
			templates.ExecuteTemplate(w, "error.html", "Ruf konnte nicht gespeichert werden")
			return
		}

		data := struct {
			CurrentUser string
			Message     string
		}{
			Message:     "Ruf erfolgreich erstellt",
			CurrentUser: contextString("current_user", r),
		}

		log.Info(templates.ExecuteTemplate(w, "success.html", data))

	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerActiveCalls(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		calls, err := bridge.GetActiveCalls()
		if err != nil {
			log.Warn(err)
			io.WriteString(w, "Failed to retrieve calls")
			return
		}

		data := struct {
			Data        []Call
			CurrentUser string
		}{
			Data:        calls,
			CurrentUser: contextString("current_user", r),
		}

		// Show all active calls
		log.Info(templates.ExecuteTemplate(w, "active.html", data))

	} else {
		io.WriteString(w, "Invalid request")
	}
}
