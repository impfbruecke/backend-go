package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handlerSendCall(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Show the "Send Call" page
		templates.ExecuteTemplate(w, "call.html", nil)

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
		templates.ExecuteTemplate(w, "active.html", nil)
	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerImport(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		// Show import page
		templates.ExecuteTemplate(w, "import.html", nil)

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
		templates.ExecuteTemplate(w, "status.html", nil)

	} else {
		io.WriteString(w, "Invalid request")
	}
}

func ParseTemplates() *template.Template {
	templ := template.New("")
	err := filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				log.Println(err)
			}
		}

		return err
	})

	if err != nil {
		panic(err)
	}

	return templ
}

var templates *template.Template

func redirectHandler(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "http://api.impfbruecke.de", 301)
}

func main() {
	r := mux.NewRouter()

	templates = ParseTemplates()
	// Routes consist of a path and a handler function.

	// POST

	// Send out a call from JSON input.
	// Input data fields:
	// - Number of Doses
	// - Open Time
	// - Close Time
	// - Location

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	r.HandleFunc("/", redirectHandler)
	r.HandleFunc("/call", handlerSendCall)
	r.HandleFunc("/call/{id}", handlerStatus)
	r.HandleFunc("/active", handlerActiveCalls)
	r.HandleFunc("/import", handlerImport)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe("localhost:12000", r))
}
