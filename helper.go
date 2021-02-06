package main

import (
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
