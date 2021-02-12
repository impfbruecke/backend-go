package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

func handlerAddPerson(w http.ResponseWriter, r *http.Request) {

	tData := TmplData{
		CurrentUser: contextString(contextKeyCurrentUser, r),
	}

	// GET requests show import page
	if r.Method == http.MethodGet {

		persons, err := bridge.GetPersons()
		if err != nil {
			log.Warn(err)
			io.WriteString(w, "Failed to retrieve persons")
			return
		}

		tData.Persons = persons
		log.Info(templates.ExecuteTemplate(w, "add.html", tData))

	} else if r.Method == http.MethodPost {

		// TODO validate data, ignore empty
		r.ParseForm()
		data := r.Form
		phone := data.Get("phone")
		group := data.Get("group")

		// Try to create new call from input data
		groupNum, err := strconv.Atoi(group)

		if err != nil {
			log.Debug(err)
			tData.AppMessages = append(tData.AppMessages, "Ungültige Gruppe")
			templates.ExecuteTemplate(w, "add.html", tData)
			return
		}

		if phone == "" {
			tData.AppMessages = append(tData.AppMessages, "Fehlende Rufnummer")
			templates.ExecuteTemplate(w, "add.html", tData)
			return
		}

		person, err := NewPerson(0, groupNum, phone, false)
		if err != nil {
			log.Debug(err)
			tData.AppMessages = append(tData.AppMessages, "Eingaben ungültig")
			templates.ExecuteTemplate(w, "add.html", tData)
			return
		}

		// Add call to bridge
		if err := bridge.AddPerson(person); err != nil {
			log.Warn(err)
			log.Warn(person)

			tData.AppMessages = append(tData.AppMessages, "Personen konnten nicht gespeichert werden. Rufnummer schon vorhanden?")
			templates.ExecuteTemplate(w, "add.html", tData)
			return
		}

		// Send onboarding notificatino
		if err := bridge.sender.SendMessageOnboarding(person.Phone); err != nil {
			log.Error(err)
		}

		log.Debug("Person was added")
		tData.AppMessageSuccess = "Import Erfolgreich!"
		tData.CurrentUser = contextString(contextKeyCurrentUser, r)
		log.Info(templates.ExecuteTemplate(w, "add.html", tData))

	} else {
		io.WriteString(w, "Invalid request")
	}
}
