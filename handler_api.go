package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"

	"encoding/json"
	"github.com/gorilla/mux"
)

func handlerApi(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		log.Info("FORM1:")
		r.ParseForm()

		for key, value := range r.Form {
			log.Printf("%s = %s\n", key, value)
		}

		log.Info("FORM2:")
		log.Println(r.Form)

		decoder := json.NewDecoder(r.Body)
		var t map[string]string
		err := decoder.Decode(&t)
		if err != nil {
			panic(err)
		}

		log.Println("the json")
		log.Println(t)
		log.Println("the json end")

		if phoneNumber, ok := t["number"]; ok {
			w.WriteHeader(http.StatusOK)
			switch mux.Vars(r)["endpoint"] {
			case "ja":
				if err := bridge.PersonAcceptCall(phoneNumber); err != nil {
					log.Error(err)
				}
			case "storno":
				if err := bridge.PersonCancelCall(phoneNumber); err != nil {
					log.Error(err)
				}
			case "loeschen":
				if err := bridge.PersonDelete(phoneNumber); err != nil {
					log.Error(err)
				}
			default:
				log.Debug("Invalid request to API recieved")
				io.WriteString(w, "Invalid request")
			}
		}

	} else {
		io.WriteString(w, "Invalid request")
	}
}
