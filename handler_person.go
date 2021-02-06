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

		templates.ExecuteTemplate(w, "add.html", persons)

	} else if r.Method == http.MethodPost {

		// TODO validate data, ignore empty
		r.ParseForm()
		data := r.Form
		phone := data.Get("phone")
		group := data.Get("group")

		// Try to create new call from input data
		// r.ParseForm()

		groupNum, err := strconv.Atoi(group)

		if err != nil {
			log.Fatal(err)
		}

		person, err := NewPerson(0, groupNum, phone, false)
		if err != nil {
			log.Println(err)
			templates.ExecuteTemplate(w, "error.html", "Eingaben ungültig")
			return
		}

		// Add call to bridge
		if err := bridge.AddPerson(person); err != nil {
			log.Println(err)
			templates.ExecuteTemplate(w, "error.html", "Personen konnten nicht gespeichert werden")
			return
		}

		templates.ExecuteTemplate(w, "success.html", "Import Erfolgreich!")

	} else {
		io.WriteString(w, "Invalid request")
	}
}
