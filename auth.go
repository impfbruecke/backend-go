package main

import (
	"crypto/rsa"
	// "database/sql"
	// "encoding/gob"
	// "github.com/gorilla/mux"
	// "github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	// "golang.org/x/crypto/bcrypt"
	// "html/template"
	"net/http"
	// "time"
)

// keys are held in global variables
// i havn't seen a memory corruption/info leakage in go yet
// but maybe it's a better idea, just to store the public key in ram?
// and load the signKey on every signing request? depends on  your usage i guess
var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

type Credentials struct {
	Password string `db:"password"`
	Username string `db:"username"`
}

// reads the form values, checks them and creates a session
func authHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "cookie-name")
	if err != nil {
		log.Warn("Could not retrieve session cookie from store")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("pass") != "pass" {

		session.AddFlash("The code was incorrect")

		err = session.Save(r, w)
		if err != nil {
			log.Warn("Error saving session cookie: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/forbidden", http.StatusFound)
		return
	}

	username := r.FormValue("username")

	user := &User{
		Username:      username,
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
	templates.ExecuteTemplate(w, "login.html", "Bitte einloggen")
}

// middlewareAuth  prepended to all handlers to handle authentication
func middlewareAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session, err := store.Get(r, "cookie-name")
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

		// Token is valid. At this point the login is succesfull and this
		// middleware has done it's job. Pass on to the next handler and
		// record the login in the application log for good measure
		log.Debugf("Login successful for user:%v\n", user)

		// tpl.ExecuteTemplate(w, "secret.gohtml", user.Username)
		next.ServeHTTP(w, r)

	})
}

func forbidden(w http.ResponseWriter, r *http.Request) {
	log.Warn("forbidden reached")

	session, err := store.Get(r, "cookie-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Error getting cookie: ", err)
		return
	}

	flashMessages := session.Flashes()
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Error saving cookie", err)
		return
	}

	// tpl.ExecuteTemplate(w, "forbidden.gohtml", flashMessages)
	// tpl.ExecuteTemplate(w, "error.html", flashMessages)
	templates.ExecuteTemplate(w, "error.html", flashMessages)
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
