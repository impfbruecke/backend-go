package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log"
	"net/http"
)

func handlerSendCall(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Show the "Send Call" page
		tmpl := template.Must(template.ParseFiles("call.html"))
		tmpl.Execute(w, "data goes here")

	} else if r.Method == http.MethodPost {
		// Create call with entered details
		io.WriteString(w, "This is a post request")
	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerActiveCalls(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show all active calls
		tmpl := template.Must(template.ParseFiles("calls.html"))
		tmpl.Execute(w, "data goes here")
	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerImport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		// Show import page
		tmpl := template.Must(template.ParseFiles("import.html"))
		tmpl.Execute(w, "data goes here")

	} else if r.Method == http.MethodPost {

		// Upload data and save
		io.WriteString(w, "This is a post request")

	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		// Show Call Details
		tmpl := template.Must(template.ParseFiles("status.html"))
		tmpl.Execute(w, "data goes here")

	} else {
		io.WriteString(w, "Invalid request")
	}
}

func main() {
	r := mux.NewRouter()
	// Routes consist of a path and a handler function.

	// POST

	// Send out a call from JSON input.
	// Input data fields:
	// - Number of Doses
	// - Open Time
	// - Close Time
	// - Location
	r.HandleFunc("/call", handlerSendCall)
	r.HandleFunc("/active", handlerActiveCalls)
	r.HandleFunc("/import", handlerImport)
	r.HandleFunc("/status/:id", handlerStatus)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))
}
