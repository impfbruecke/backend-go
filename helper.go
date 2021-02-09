package main

import (
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
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

func contextString(key string, r *http.Request) string {

	v := r.Context().Value(key)

	if v != nil {
		return v.(string)
	} else {
		return ""
	}
}

// genOTP generates a OTP to verify the person on-site. The OTP is the first 5
// chars of the SHA-1 hash of phonenumber+callID+tokenSecret
func genOTP(phone string, callID int) string {
	h := sha1.New()
	h.Write([]byte(phone + strconv.Itoa(callID) + tokenSecret))
	return hex.EncodeToString(h.Sum(nil))[1:5]
}
