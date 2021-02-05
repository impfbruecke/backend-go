package main

import (
	log "github.com/sirupsen/logrus"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
			log.Warn(err)
			templates.ExecuteTemplate(w, "error.html", "Eingaben ungültig, Ruf wurde nicht erstellt")
			return
		}

		// Add call to bridge
		if err := bridge.AddCall(call); err != nil {
			log.Warn(err)
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
			log.Warn(err)
			io.WriteString(w, "Failed to retrieve calls")
			return
		}

		// Show all active calls
		templates.ExecuteTemplate(w, "active.html", calls)

	} else {
		io.WriteString(w, "Invalid request")
	}
}
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
		r.ParseForm()

		groupNum, err := strconv.Atoi(group)

		if err != nil {
			log.Fatal(err)
		}

		person, err := NewPerson(0, groupNum, phone)
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

	if os.Getenv("IMPF_MODE") == "DEVEL" {

		// Output to stdout instead of the default stderr
		// Can be any io.Writer, see below for File example
		log.SetOutput(os.Stdout)
		// Only log the warning severity or above.
		log.SetLevel(log.DebugLevel)
		log.Info("Starting in DEVEL mode")
	}

	// Intial setup. Instanciate bridge and parse html templates
	log.Info("Parsing templates")
	templates = parseTemplates()

	log.Info("Creating new bridge")
	bridge = NewBridge()

	// Routes
	log.Info("Setting up routes")
	r := mux.NewRouter()

	// Serve static files like css and images
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Redirect / to /call for convenience
	r.HandleFunc("/", redirectHandler)

	// Handler functions to endpoints
	r.HandleFunc("/call", handlerSendCall)
	r.HandleFunc("/call/{id}", handlerStatus)
	r.HandleFunc("/active", handlerActiveCalls)
	r.HandleFunc("/add", handlerAddPerson)
	r.HandleFunc("/upload", handlerUpload)

	// Handle incoming webhooks
	r.HandleFunc("/api/{endpoint}", handlerApi)

	// Bind to a port and pass our router in

	bindAddress := "localhost:12000"
	log.Info("Starting server on: ", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, r))
}
