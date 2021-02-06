package main

import (
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// All templates inside of ./templates and it's subfolders are parsed and can be executed by it's filename
var templates *template.Template
var bridge *Bridge

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/call", 301)
}

func main() {

	// Show more logs if IMPF_MODE=DEVEL is set
	if os.Getenv("IMPF_MODE") == "DEVEL" {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.DebugLevel)
		log.Info("Starting in DEVEL mode")
	}

	// Intial setup. Instanciate bridge and parse html templates
	log.Info("Parsing templates")
	templates = parseTemplates()

	bridge = NewBridge()

	// Routes
	log.Info("Setting up routes")
	router := mux.NewRouter()

	// Serve static files like css and images
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Handler functions to endpoints
	router.HandleFunc("/", redirectHandler)          // Redirect / to /call
	router.HandleFunc("/call", handlerSendCall)      // Send a call
	router.HandleFunc("/call/{id}", handlerStatus)   // Get call details
	router.HandleFunc("/active", handlerActiveCalls) // List active calls
	router.HandleFunc("/add", handlerAddPerson)      // Add single person
	router.HandleFunc("/upload", handlerUpload)      // CSV upload
	router.HandleFunc("/api/{endpoint}", handlerApi) // Handle incoming webhooks

	// Wrap the router in a function to log all requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		router.ServeHTTP(w, r)
	})

	// Bind to addrerss, with specified routing
	bindAddress := "localhost:12000"
	log.Info("Starting server on: ", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, handler))
}
