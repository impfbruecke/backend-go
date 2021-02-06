package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// Handle endpoints. GET will usually render a html template, POST will be used
// for data import and specific requests
func handlerSendCall(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Show the "Send Call" page
		templates.ExecuteTemplate(w, "call.html", nil)

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

		templates.ExecuteTemplate(w, "success.html", "Ruf erfolgreich erstellt")

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

		// Show all active calls
		templates.ExecuteTemplate(w, "active.html", calls)

	} else {
		io.WriteString(w, "Invalid request")
	}
}
