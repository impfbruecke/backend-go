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

var (

	// Instance of the main application
	bridge *Bridge

	// All templates inside of ./templates and it's subfolders are parsed and can be executed by it's filename
	templates *template.Template

	// store will hold all session data
	store *sessions.CookieStore

	// API auth for twilio
	apiUser     string
	apiPass     string
	tokenSecret string
	disableSMS  string
	dbPath      string
)

// User holds a users account information
type User struct {
	Username      string
	Authenticated bool
}

func init() {

	store = sessions.NewCookieStore([]byte(os.Getenv("IMPF_SESSION_SECRET")))

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

	// Show more logs if IMPF_MODE=DEVEL is set
	apiUser = os.Getenv("IMPF_TWILIO_USER")
	apiPass = os.Getenv("IMPF_TWILIO_PASS")
	tokenSecret = os.Getenv("IMPF_TOKEN_SECRET")
	disableSMS = os.Getenv("IMPF_DISABLE_SMS")
	dbPath = os.Getenv("IMPF_DB_FILE")

	// Add default if not set
	if dbPath == "" {
		dbPath = "./data.db"
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
	router.HandleFunc("/forbidden", forbiddenHandler)

	// Router for all routes under https://domain.tld/auth/ will have to pass
	// through the authentication middleware. Put any routes here, that should
	// be protected by user and password

	subRouterAPI := router.PathPrefix("/api").Subrouter()
	subRouterAPI.Use(middlewareAPI)
	subRouterAPI.HandleFunc("/{endpoint}", handlerAPI)

	subRouterAuth := router.PathPrefix("/auth").Subrouter()
	subRouterAuth.Use(middlewareAuth)
	subRouterAuth.HandleFunc("/call", handlerSendCall)           // Send a call
	subRouterAuth.HandleFunc("/active/{id}", handlerActiveCalls) // Get call details
	subRouterAuth.HandleFunc("/active", handlerActiveCalls)      // List active calls
	subRouterAuth.HandleFunc("/add", handlerAddPerson)           // Add single person
	subRouterAuth.HandleFunc("/upload", handlerUpload)           // CSV upload

	handler := middlewareLog(router)

	// Bind to addrerss, with specified routing
	bindAddress := "localhost:12000"
	log.Info("Starting server on: ", bindAddress)
	log.Fatal(http.ListenAndServe(bindAddress, handler))
}

func middlewareAPI(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		// Check we were able to get a username and password
		if !ok {
			log.Error("Failed to retrieve username and password API request")
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		// Check if the match the env vars set for the application
		if apiUser != user || apiPass != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
