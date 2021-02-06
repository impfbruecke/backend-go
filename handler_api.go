package main

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func handlerApi(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

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
