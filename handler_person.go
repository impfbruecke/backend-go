package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

func handlerAddPerson(w http.ResponseWriter, r *http.Request) {

	// GET requests show import page
	if r.Method == http.MethodGet {

		persons, err := bridge.GetPersons()
		if err != nil {
			log.Warn(err)
			io.WriteString(w, "Failed to retrieve persons")
			return
		}

		data := struct {
			Data        []Person
			CurrentUser string
		}{
			Data:        persons,
			CurrentUser: contextString("current_user", r),
		}

		log.Info(templates.ExecuteTemplate(w, "add.html", data))

	} else if r.Method == http.MethodPost {

		// TODO validate data, ignore empty
		r.ParseForm()
		data := r.Form
		phone := data.Get("phone")
		group := data.Get("group")

		// Try to create new call from input data
		groupNum, err := strconv.Atoi(group)

		if err != nil {
			templates.ExecuteTemplate(w, "error.html", "Ungültige Gruppe")
			return
		}

		if phone == "" {
			templates.ExecuteTemplate(w, "error.html", "Fehlende Rufnummer")
			return
		}

		person, err := NewPerson(0, groupNum, phone, false)
		if err != nil {
			log.Info(err)
			log.Debug(person)
			templates.ExecuteTemplate(w, "error.html", "Eingaben ungültig")
			return
		}

		// Add call to bridge
		if err := bridge.AddPerson(person); err != nil {
			log.Warn(err)
			log.Warn(person)
			templates.ExecuteTemplate(w, "error.html", "Personen konnten nicht gespeichert werden")
			return
		}

		tmpldata := struct {
			Message     string
			CurrentUser string
		}{
			Message:     "Import Erfolgreich!",
			CurrentUser: contextString("current_user", r),
		}

		log.Info(templates.ExecuteTemplate(w, "success.html", tmpldata))

	} else {
		io.WriteString(w, "Invalid request")
	}
}
