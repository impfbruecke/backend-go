package main

import (
	jwt "github.com/dgrijalva/jwt-go"

	"fmt"
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

// middlewareAuth  prepended to all handlers to handle authentication
func middlewareAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// check if we have a cookie with out tokenName
		tokenCookie, err := r.Cookie(tokenName)
		switch {
		case err == http.ErrNoCookie:
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "No Token, no fun!")
			return
		case err != nil:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error while Parsing cookie!")
			log.Printf("Cookie parse error: %v\n", err)
			return
		}

		// just for the lulz, check if it is empty.. should fail on Parse anyway..
		if tokenCookie.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "No Token, no fun!")
			return
		}

		// validate the token
		token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return verifyKey, nil
		})

		// branch out into the possible error from signing
		switch err.(type) {

		case nil: // no error

			if !token.Valid { // but may still be invalid
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, "Invalid token!")
				return
			}

			// see stdout and watch for the CustomUserInfo, nicely unmarshalled
			log.Printf("Someone accessed resricted area! Token:%+v\n", token)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)

			next.ServeHTTP(w, r)

		case *jwt.ValidationError: // something was wrong during the validation
			vErr := err.(*jwt.ValidationError)

			switch vErr.Errors {
			case jwt.ValidationErrorExpired:
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintln(w, "Token Expired, get a new one.")
				return

			default:
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error while Parsing Token!")
				log.Printf("ValidationError error: %+v\n", vErr.Errors)
				return
			}

		default: // something else went wrong
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error while Parsing Token!")
			log.Printf("Token parse error: %v\n", err)
			return
		}

	})
	// return http.HandlerFunc(restrictedHandler)
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
