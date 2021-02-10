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

type TwillioSender struct {
	endpoint, user, token, from string
}

func (s TwillioSender) SendMessage(msg_to, msg_body string) error {

	// Set required data
	data := url.Values{}
	data.Set("To", msg_to)
	data.Set("From", s.from)

	values := map[string]string{
		"type":    "nachricht",
		"message": msg_body,
	}

	json_data, err := json.Marshal(values)
	data.Set("Parameters", string(json_data))

	// Create a new client and request
	client := &http.Client{}
	r, err := http.NewRequest("POST", s.endpoint, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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

func (s TwillioSender) SendMessageOnboarding(toPhone string) error {
	msg := "Willkommen bei der kurzfristigen Impfterminvergabe der Feuerwehr Duisburg. Möchten Sie diesen Service nicht benutzen, antworten Sie jederzeit mit \"LÖSCHEN\"."
	return s.SendMessage(toPhone, msg)
}

func (s TwillioSender) SendMessageNotify(toPhone, start, end, location string) error {
	msg := "Sie haben die Möglichkeit zur Corona-Impfung, heute " + start + "-" + end + "h " + location + ". Antworten Sie für Zusage mit \"JA\""
	return s.SendMessage(toPhone, msg)
}

func (s TwillioSender) SendMessageReject(toPhone string) error {
	msg := "Leider wurden zwischenzeitlich schon alle Termine vergeben. Sie bleiben im System und werden ggf. wieder benachrichtigt."
	return s.SendMessage(toPhone, msg)
}

func (s TwillioSender) SendMessageAccept(toPhone, start, end, location, otp string) error {
	msg := "Termin bestätigt. " + location + ", heute " + start + "-" + end + "h . ID: " + otp + ". Falls Sie den Termin nicht wahrnehmen können, bitte \"STORNO\" antworten."
	return s.SendMessage(toPhone, msg)
}

func (s TwillioSender) SendMessageDelete(toPhone string) error {
	msg := "Sie wurden erfolgreich entfernt und erhalten keine weiteren Nachrichten von uns."
	return s.SendMessage(toPhone, msg)
}

func NewTwillioSender(endpoint, user, token, from string) *TwillioSender {

	return &TwillioSender{
		endpoint: endpoint,
		token:    token,
		user:     user,
		from:     from,
	}
}
