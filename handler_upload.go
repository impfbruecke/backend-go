package main

import (
	"encoding/csv"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
)

func handlerUpload(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		io.WriteString(w, "Invalid request")
		return
	}

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 200 MB files.
	r.ParseMultipartForm(200 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("datei")
	if err != nil {
		log.Warn("Error Retrieving the File: ", err)
		return
	}

	defer file.Close()

	log.Debugf("Uploaded File: %+v\n", handler.Filename)
	log.Debugf("File Size: %+v\n", handler.Size)
	log.Debugf("MIME Header: %+v\n", handler.Header)

	csvReader := csv.NewReader(file)

	// List of new persons
	var persons []Person

	for {
		// Read each record from csv
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			return
		}

		// Parse the group number into an integer and return on errors
		groupNum, err := strconv.Atoi(record[1])
		if err != nil {
			log.Warn(err)
			return
		}

		// Try to create a new persion object from the data and return on
		// errors
		p, err := NewPerson(0, groupNum, record[1], false)
		if err != nil {
			log.Warn(err)
			return
		}

		// If everything went well until here, add the persion to the list of new persons
		persons = append(persons, p)

	}

	// Add the list of new persons to the bridge/database
	bridge.AddPersons(persons)
}
