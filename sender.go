package main

import (
	// "bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// TwillioSender abstracts interactions with the twilio API
type TwillioSender struct {
	endpoint, user, token, from string
}

// SendMessage sends a SMS with the provided body to the number. It will return
// an error if sending fails
func (s TwillioSender) SendMessage(msgTo, msgBody string) error {

	var jsonData []byte
	var err error

	// Set required data
	data := url.Values{}
	data.Set("To", msgTo)
	data.Set("From", s.from)

	values := map[string]string{
		"type":    "nachricht",
		"message": msgBody,
	}

	if jsonData, err = json.Marshal(values); err != nil {
		return err
	}

	data.Set("Parameters", string(jsonData))

	// Create a new client and request
	client := &http.Client{}
	r, err := http.NewRequest("POST", s.endpoint, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return err
	}

	// Set necessary headers
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.SetBasicAuth(s.user, s.token)

	if disableSMS != "" {
		log.Info("SMS sending disabled. Unset IMPF_DISABLE_SMS to enable")
		return nil
	}

	// Execute the request
	res, err := client.Do(r)
	if err != nil {
		return err
	}

	// Print the result status code
	log.Debug(res.Status)

	// Read the body of the response for logging
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(string(body))

	return nil
}

// SendMessageOnboarding sends the onboarding message, notifying a person that
// his/her number has been added to the application
func (s TwillioSender) SendMessageOnboarding(toPhone string) error {
	msg := "Willkommen bei der kurzfristigen Impfterminvergabe der Feuerwehr Duisburg. Möchten Sie diesen Service nicht benutzen, antworten Sie jederzeit mit \"LÖSCHEN\"."
	return s.SendMessage(toPhone, msg)
}

// SendMessageNotify send the invitation message when the person is invited to a call
func (s TwillioSender) SendMessageNotify(toPhone, start, end, location string) error {
	msg := "Sie haben die Möglichkeit zur Corona-Impfung, heute " + start + "-" + end + "h " + location + ". Antworten Sie für Zusage mit \"JA\""
	return s.SendMessage(toPhone, msg)
}

// SendMessageReject sends the rejection message when the person tries to
// accept a call but it is already wull
func (s TwillioSender) SendMessageReject(toPhone string) error {
	msg := "Leider wurden zwischenzeitlich schon alle Termine vergeben. Sie bleiben im System und werden ggf. wieder benachrichtigt."
	return s.SendMessage(toPhone, msg)
}

// SendMessageAccept sends the acceptance message when a person replies to a
// call in time and has been given a spot
func (s TwillioSender) SendMessageAccept(toPhone, start, end, location, otp string) error {
	msg := "Termin bestätigt. " + location + ", heute " + start + "-" + end + "h . ID: " + otp + ". Falls Sie den Termin nicht wahrnehmen können, bitte \"STORNO\" antworten."
	return s.SendMessage(toPhone, msg)
}

// SendMessageDelete notifies a person that their number has been deleted from
// the database
func (s TwillioSender) SendMessageDelete(toPhone string) error {
	msg := "Sie wurden erfolgreich entfernt und erhalten keine weiteren Nachrichten von uns."
	return s.SendMessage(toPhone, msg)
}

// NewTwillioSender creates a new instance of the sender from the parameters
// passed
func NewTwillioSender(endpoint, user, token, from string) *TwillioSender {

	return &TwillioSender{
		endpoint: endpoint,
		token:    token,
		user:     user,
		from:     from,
	}
}
