package main

import (
	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
)

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

func logRequest(r *http.Request) {

	log.Println(r.Method + " " + r.URL.String())

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)

	if err != nil {
		log.Debug(err)
	}

	log.Debug(string(requestDump))

}

// middlewareLog is prepended to all handlers to log http requsts uniformly
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		next.ServeHTTP(w, r)
	})
}

func extractClaims(tokenStr string) (jwt.MapClaims, bool) {

	var err error
	var token *jwt.Token

	token, err = jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	if err != nil {
		log.Error(err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		log.Printf("Invalid JWT Token")
		return nil, false
	}
}
