package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func handlerStatus(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		// Get the id we want to query
		callID := mux.Vars(r)["id"]

		details, err := bridge.GetCallStatus(callID)

		if err != nil {
			log.Warn("Failed to retrieve call details for call ID:", callID)
			return
		}

		w.WriteHeader(http.StatusOK)

		data := struct {
			Data        callstatus
			CurrentUser string
		}{
			Data:        details,
			CurrentUser: "someone",
		}

		// Show Call Details
		log.Info(templates.ExecuteTemplate(w, "status.html", data))
	} else {
		io.WriteString(w, "Invalid request")
	}
}
