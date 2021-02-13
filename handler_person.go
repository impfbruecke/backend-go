package main

import (
	"io"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
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
			if _, err := io.WriteString(w, "Failed to retrieve persons"); err != nil {
				log.Error(err)
			}
			return
		}

		tData.Persons = persons
		if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
			log.Error(err)
		}

	} else if r.Method == http.MethodPost {

		if err := r.ParseForm(); err != nil {
			tData.AppMessages = append(tData.AppMessages, "Ungültige Eingaben")
			if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
				log.Error(err)
			}
			return
		}

		data := r.Form
		phone := data.Get("phone")
		group := data.Get("group")

		// Try to create new call from input data
		groupNum, err := strconv.Atoi(group)

		if err != nil {
			log.Debug(err)
			tData.AppMessages = append(tData.AppMessages, "Ungültige Gruppe")
			if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
				log.Error(err)
			}
			return
		}

		if phone == "" {
			tData.AppMessages = append(tData.AppMessages, "Fehlende Rufnummer")
			if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
				log.Error(err)
			}
			return
		}

		person, err := NewPerson(0, groupNum, phone, false)
		if err != nil {
			log.Debug(err)
			tData.AppMessages = append(tData.AppMessages, "Eingaben ungültig")
			if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
				log.Error(err)
			}
			return
		}

		// Add call to bridge
		if err := bridge.AddPerson(person); err != nil {
			log.Warn(err)
			log.Warn(person)

			tData.AppMessages = append(tData.AppMessages, "Personen konnten nicht gespeichert werden. Rufnummer schon vorhanden?")
			if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
				log.Error(err)
			}
			return
		}

		// Send onboarding notificatino
		if err := bridge.sender.SendMessageOnboarding(person.Phone); err != nil {
			log.Error(err)
		}

		log.Debug("Person was added")
		tData.AppMessageSuccess = "Import Erfolgreich!"
		tData.CurrentUser = contextString(contextKeyCurrentUser, r)
		if err := templates.ExecuteTemplate(w, "importPersons.html", tData); err != nil {
			log.Error(err)
		}

	} else {
		if _, err := io.WriteString(w, "Invalid request"); err != nil {
			log.Error(err)
		}
	}
}
