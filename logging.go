package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
)

func logRequest(r *http.Request) {

	log.Debugf("%s %s\n", r.Method, r.URL.String())

	// Log a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)

	if err != nil {
		log.Debug(err)
	}

	log.Debugf("\n%s\n", string(requestDump))

}

// middlewareLog is prepended to all handlers to log http requsts uniformly
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		next.ServeHTTP(w, r)
	})
}
