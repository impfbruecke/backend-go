package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func handlerApi(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		log.Info("FORM:")
		r.ParseForm()
		log.Println(r.Form)

		w.WriteHeader(http.StatusOK)
		switch mux.Vars(r)["endpoint"] {
		case "ja":
			bridge.PersonAcceptCall()
		case "storno":
			bridge.PersonCancelCall()
		case "loeschen":
			bridge.PersonDelete()
		default:
			io.WriteString(w, "Invalid request")
		}

	} else {
		io.WriteString(w, "Invalid request")
	}
}
