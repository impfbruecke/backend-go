package main

import (
	"encoding/gob"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/sessions"
)

// All templates inside of ./templates and it's subfolders are parsed and can be executed by it's filename
var templates *template.Template
var bridge *Bridge

const tokenName = "AccessToken"

// User holds a users account information
type User struct {
	Username      string
	Authenticated bool
}

// store will hold all session data
var store *sessions.CookieStore

// tpl holds all parsed templates
var tpl *template.Template

func init() {

	store = sessions.NewCookieStore([]byte("asdaskdhasdhgsajdgasdsadksakdhasidoajsdousahdopj"))

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 15,
		HttpOnly: true,
	}

	gob.Register(User{})

	// Show more logs if IMPF_MODE=DEVEL is set
	if os.Getenv("IMPF_MODE") == "DEVEL" {
		log.SetOutput(os.Stdout)
		log.SetLevel(log.DebugLevel)
		log.Info("Starting in DEVEL mode")
	}

	// Intial setup. Instanciate bridge and parse html templates
	log.Info("Parsing templates")
	templates = parseTemplates()

}

func main() {

	bridge = NewBridge()

	// Routes
	log.Info("Setting up routes")
	router := mux.NewRouter()

	// Serve static files like css and images
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/login", loginHandler)

	// Handler functions to endpoints
	router.HandleFunc("/", loginHandler)            // Login
	router.HandleFunc("/authenticate", authHandler) // Authenticate
	router.HandleFunc("/forbidden", forbidden)

	// Router for all routes under https://domain.tld/auth/ will have to pass
	// through the authentication middleware. Put any routes here, that should
	// be protected by user and password

	subRouterAuth := router.PathPrefix("/auth").Subrouter()
	subRouterAuth.Use(middlewareAuth)
	subRouterAuth.HandleFunc("/call", handlerSendCall)      // Send a call
	subRouterAuth.HandleFunc("/call/{id}", handlerStatus)   // Get call details
	subRouterAuth.HandleFunc("/active", handlerActiveCalls) // List active calls
	subRouterAuth.HandleFunc("/add", handlerAddPerson)      // Add single person
	subRouterAuth.HandleFunc("/upload", handlerUpload)      // CSV upload
	subRouterAuth.HandleFunc("/api/{endpoint}", handlerApi) // Handle incoming webhooks

	handler := middlewareLog(router)

	// Bind to addrerss, with specified routing
	bindAddress := "localhost:12000"
	log.Info("Starting server on: ", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, handler))
}
