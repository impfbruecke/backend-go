package main

import (
	jwt "github.com/dgrijalva/jwt-go"

	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"io/ioutil"
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

	router.HandleFunc("/login", loginHandler)

	// Handler functions to endpoints
	router.HandleFunc("/", loginHandler)            // Login
	router.HandleFunc("/authenticate", authHandler) // Authenticate

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

// middlewareLog is prepended to all handlers to log http requsts uniformly
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		next.ServeHTTP(w, r)
	})
}

// read the key files before starting http handlers
func init() {
	var err error
	var signBytes []byte
	var verifyBytes []byte

	if signBytes, err = ioutil.ReadFile(privKeyPath); err != nil {
		log.Fatal(err)
	}

	if signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes); err != nil {
		log.Fatal(err)
	}

	if verifyBytes, err = ioutil.ReadFile(pubKeyPath); err != nil {
		log.Fatal(err)
	}

	if verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes); err != nil {
		log.Fatal(err)
	}
}
