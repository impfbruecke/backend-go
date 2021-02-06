package main

import (
	"crypto/rsa"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

// location of the files used for signing and verification
const (
	privKeyPath = "keys/app.rsa"     // openssl genrsa -out app.rsa keysize
	pubKeyPath  = "keys/app.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
	tokenName   = "AccessToken"
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

// reads the form values, checks them and creates the token
func authHandler(w http.ResponseWriter, r *http.Request) {
	// make sure its post
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		templates.ExecuteTemplate(w, "login.html", "Ungültiger Request")
		return
	}

	inputUser := r.FormValue("user")
	inputPass := r.FormValue("pass")

	log.Debug("Trying to authenticate: user[%s] pass[%s]\n", inputUser, inputPass)

	// Create an instance of `Credentials` to store the credentials we get from
	// the database
	storedCreds := Credentials{}

	// Get the existing entry present in the database for the given username
	if err := bridge.db.Get(&storedCreds, "SELECT * FROM users WHERE username=$1", inputUser); err != nil {

		log.Error("the error")

		// TODO use something like this to implement checking if the error was
		// caused by a non-existing user

		// if err == sql.ErrNoRows {
		// 	// User not present in the database
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }
		// If the error is of any other type, send a 500 status

		//TODO better error handling. For now attempting to login with an user
		//that does not exist results in a server error being displayed
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Compare the stored hashed password, with the hashed version of the password that was received
	if err := bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(inputPass)); err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusForbidden)
		templates.ExecuteTemplate(w, "login.html", "Ungültiger Login")
		return
	}

	// If we reach this point, that means the users password was correct and
	// that they are authorized. The default 200 status is sent

	// Create a signer for rsa 256
	token := jwt.New(jwt.GetSigningMethod("RS256"))

	// Set token properties and permissions
	token.Claims = jwt.MapClaims{
		// Set access level in claims. This will allow to implement multiple
		// access levels (e.g. "normal", "admin", "moderator") if needed later
		// on. For now all users have access level 1 since there is only one
		// user.

		"AccessToken": "level1",

		// Set the expire time to one hour. After this period the token will be
		// invalidated requiring the user to login again
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	}

	// Sign the token. We don't expect any errors here. If the occur anyway,
	// something went wrong. Show login again and log a warning to look into
	// the problem
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.ExecuteTemplate(w, "login.html", "Token Fehler")
		log.Warn("Token Signing error: %v\n", err)
		return
	}

	// Storing the token in a cookie may break cross-domain API usage, but
	// there is no need for it now and we can avoid using javascript. The calls
	// to the API will be secured by basic auth in the reverse proxy and not by
	// the application itself
	http.SetCookie(w, &http.Cookie{
		Name:       tokenName,
		Value:      tokenString,
		Path:       "/",
		RawExpires: "0",
	})

	// Successful authentication. Redirect to call.html for convenience
	http.Redirect(w, r, "/auth/call", 301)
}

// Serve login page
func loginHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", "Bitte einloggen")
}

// middlewareAuth  prepended to all handlers to handle authentication
func middlewareAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// check if we have a cookie with out tokenName
		tokenCookie, err := r.Cookie(tokenName)
		switch {
		case err == http.ErrNoCookie:
			w.WriteHeader(http.StatusUnauthorized)
			templates.ExecuteTemplate(w, "login.html", "Login Token ungültig, bitte einloggen")
			return
		case err != nil:
			w.WriteHeader(http.StatusUnauthorized)
			templates.ExecuteTemplate(w, "login.html", "Login Cookie ungültig, bitte einloggen")
			return
		}

		// Check if the cookie is empty. This is not strictly necessary, as
		// parsing should fail anyway
		if tokenCookie.Value == "" {
			w.WriteHeader(http.StatusUnauthorized)
			templates.ExecuteTemplate(w, "login.html", "Login erforderlich")
			return
		}

		// Try to validate the token. Since we only use the one private key to
		// sign the tokens, we also only use its public counter part to verify
		token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})

		// Handle possible signing errors and show messages accordingly to the user and in the application log
		switch err.(type) {

		case nil:
			// No error occured but the token might still be invalid!
			// Check if the token is actually valid and only show the login page agin if it's not

			if !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				templates.ExecuteTemplate(w, "login.html", "Session ungültig. Bitte einloggen")
				return
			}

			// Token is valid. At this point the login is succesfull and this
			// middleware has done it's job. Pass on to the next handler and
			// record the login in the application log for good measure
			log.Debugf("Login successful with token:%+v\n", token)
			next.ServeHTTP(w, r)

		case *jwt.ValidationError:

			// The error has something to do with the validation process
			vErr := err.(*jwt.ValidationError)

			switch vErr.Errors {
			case jwt.ValidationErrorExpired:

				// Something went wrong during the validation process. The error
				// might just be an expired token, so we ask the user to login
				// again.
				log.Info("User tried to login with expired token")
				w.WriteHeader(http.StatusUnauthorized)
				templates.ExecuteTemplate(w, "login.html", "Session abgelaufen. Bitte loggen Sie sich erneut ein.")
				return

			default:
				//	If not, there is a problem with the application and we
				// show a warning in the log. At this point the only thing the user
				// can do, is to try to login again and hope the problem was
				// temporary until the bug is fixed. This should not occur.
				log.Errorf("ValidationError error: %+v\n", vErr.Errors)
				w.WriteHeader(http.StatusInternalServerError)
				templates.ExecuteTemplate(w, "login.html", "Ein Fehler ist beim Login aufgetreten.")
				return
			}

		default:
			// Something different went wrong. Log an error to and show login
			// page again. This should not happen and means there is a bug somewhere
			log.Errorf("Token parse error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			templates.ExecuteTemplate(w, "login.html", "Ein Fehler ist beim Login aufgetreten.")
			return
		}
	})
}

// func Signup(w http.ResponseWriter, r *http.Request) {
// 	// Parse and decode the request body into a new `Credentials` instance
// 	creds := &Credentials{}
// 	err := json.NewDecoder(r.Body).Decode(creds)
// 	if err != nil {
// 		// If there is something wrong with the request body, return a 400 status
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	// Salt and hash the password using the bcrypt algorithm
// 	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

// 	// Next, insert the username, along with the hashed password into the database
// 	if _, err = db.Query("insert into users values ($1, $2)", creds.Username, string(hashedPassword)); err != nil {
// 		// If there is any issue with inserting into the database, return a 500 error
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	// We reach this point if the credentials we correctly stored in the database, and the default status of 200 is sent back
// }
