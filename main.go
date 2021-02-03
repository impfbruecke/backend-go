package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// Handle endpoints. GET will usually render a html template, POST will be used
// for data import and specific requests
func handlerSendCall(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Show the "Send Call" page
		templates.ExecuteTemplate(w, "call.html", nil)

	} else if r.Method == http.MethodPost {

		// Try to create new call from input data
		r.ParseForm()
		call, err := NewCall(r.Form)
		if err != nil {
			log.Println(err)
			templates.ExecuteTemplate(w, "error.html", "Eingaben ungültig, Ruf wurde nicht erstellt")
			return
		}

		// Add call to bridge
		if err := bridge.AddCall(call); err != nil {
			log.Println(err)
			templates.ExecuteTemplate(w, "error.html", "Ruf konnte nicht gespeichert werden")
			return
		}

		templates.ExecuteTemplate(w, "success.html", "Ruf erfolgreich erstellt")

	} else {
		io.WriteString(w, "Invalid request")
	}
}

func handlerActiveCalls(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		calls, err := bridge.GetActiveCalls()
		if err != nil {
			log.Println(err)
			io.WriteString(w, "Failed to retrieve calls")
			return
		}

		// Show all active calls
		templates.ExecuteTemplate(w, "active.html", calls)

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

// All templates inside of ./templates and it's subfolders are parsed and can be executed by it's filename
var templates *template.Template
var bridge *Bridge

func parseTemplates() *template.Template {
	templ := template.New("")
	err := filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				log.Fatal(err)
			}
		}

		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return templ
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/call", 301)
}

func main() {

	r := mux.NewRouter()

	templates = parseTemplates()
	bridge = NewBridge()

	// Serve static files like css and images
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Redirect to /call
	r.HandleFunc("/", redirectHandler)

	// Handler functions to endpoints
	r.HandleFunc("/call", handlerSendCall)
	r.HandleFunc("/call/{id}", handlerStatus)
	r.HandleFunc("/active", handlerActiveCalls)
	r.HandleFunc("/import", handlerImport)

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe("localhost:12000", r))
}
