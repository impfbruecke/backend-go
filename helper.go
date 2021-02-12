package main

import (
	"crypto/sha1"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
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

func contextString(key contextKey, r *http.Request) string {

	v := r.Context().Value(key)

	if v != nil {
		return v.(string)
	}
	return ""
}

// genOTP generates a OTP to verify the person on-site. The OTP is the first 5
// chars of the SHA-1 hash of phonenumber+callID+tokenSecret
func genOTP(phone string, callID int) string {
	h := sha1.New()
	h.Write([]byte(phone + strconv.Itoa(callID) + tokenSecret))
	return hex.EncodeToString(h.Sum(nil))[1:5]
}
