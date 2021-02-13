package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"

	"encoding/json"
	"github.com/gorilla/mux"
)

func handlerAPI(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		if err := r.ParseForm(); err != nil {
			log.Error(err)
			return
		}

		for key, value := range r.Form {
			log.Debugf("%s = %s\n", key, value)
		}

		decoder := json.NewDecoder(r.Body)
		var t map[string]string
		err := decoder.Decode(&t)
		if err != nil {
			log.Error(err)
			return
		}

		header := http.StatusOK

		if phoneNumber, ok := t["number"]; ok {
			switch mux.Vars(r)["endpoint"] {
			case "ja":
				if err := bridge.PersonAcceptLastCall(phoneNumber); err != nil {
					log.Error(err)
					header = http.StatusBadRequest
				}
			case "storno":
				if err := bridge.PersonCancelCall(phoneNumber); err != nil {
					log.Error(err)
					header = http.StatusBadRequest
				}
			case "loeschen":
				if err := bridge.PersonDelete(phoneNumber); err != nil {
					log.Error(err)
					header = http.StatusBadRequest
				}
			default:
				log.Debug("Invalid request to API recieved")
				header = http.StatusBadRequest
				if _, err := io.WriteString(w, "Invalid request"); err != nil {
					log.Error(err)
				}
			}

			w.WriteHeader(header)
			return
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid request")
	}
}
