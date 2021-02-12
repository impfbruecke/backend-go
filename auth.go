package main

import (
	"context"
	"database/sql"
	// "crypto/rsa"
	// "encoding/gob"
	// "github.com/gorilla/mux"
	// "github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	// "html/template"
	"net/http"
	// "time"
)

// ImpfUser is a row from the users table. It has the username and hash of the
// password the user uses to login, aswell as the default values to populate
// new calls
type ImpfUser struct {
	Password string `db:"password"`
	Username string `db:"username"`
}

func authenticateUser(user, pass string) bool {

	log.Debugf("Trying to authenticate: user[%s] pass[%s]\n", user, pass)

	// Create an instance of `Credentials` to store the credentials from DB
	storedCreds := ImpfUser{}

	// Get the existing entry present in the database for the given username
	if err := bridge.db.Get(&storedCreds, "SELECT * FROM users WHERE username=$1", user); err != nil {

		if err == sql.ErrNoRows {
			log.Debug(err)
			// User not present in the database
		} else {
			// Something else went wrong. This should not happen, log and return
			log.Error(err)
		}
		return false
	}

	// Compare the stored hashed password, with the received and hashed
	// password
	if err := bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(pass)); err != nil {
		return false
	}

	// Session is valid
	log.Infof("Successfull authentication for user: [%s]\n", user)
	return true
}

// reads the form values, checks them and creates a session
func authHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "impf-auth")
	if err != nil {
		log.Warn("Could not retrieve session cookie from store")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Read form values of login page
	inputUsername := r.FormValue("user")
	inputPassword := r.FormValue("pass")

	if !authenticateUser(inputUsername, inputPassword) {

		session.AddFlash("Ungültige Zugangsdaten")

		err = session.Save(r, w)
		if err != nil {
			log.Warn("Error saving session cookie: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/forbidden", http.StatusFound)
		return
	}

	user := &User{
		Username:      inputUsername,
		Authenticated: true,
	}

	session.Values["user"] = user

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/auth/call", http.StatusFound)
}

// Serve login page
func loginHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", TmplData{})
}

// middlewareAuth  prepended to all handlers to handle authentication
func middlewareAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session, err := store.Get(r, "impf-auth")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := getUser(session)

		if auth := user.Authenticated; !auth {
			session.AddFlash("Login notwendig!")
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/forbidden", http.StatusFound)
			return
		}

		// Session is valid. At this point the login is succesfull and this
		// middleware has done it's job. Pass on to the next handler and
		// record the login in the application log for good measure
		log.Debugf("Login successful for user: [%v]\n", user)
		ctx := context.WithValue(r.Context(), "current_user", user.Username)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func forbiddenHandler(w http.ResponseWriter, r *http.Request) {
	log.Warn("forbidden reached")

	session, err := store.Get(r, "impf-auth")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Error getting cookie: ", err)
		return
	}

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Error saving cookie", err)
		return
	}

	tData := TmplData{AppMessages: []string{"Login Fehlgeschlagen"}}
	templates.ExecuteTemplate(w, "login.html", tData)
}

// getUser returns a user from session s
// on error returns an empty user
func getUser(s *sessions.Session) User {
	val := s.Values["user"]
	var user = User{}
	user, ok := val.(User)
	if !ok {
		return User{Authenticated: false}
	}
	return user
}
