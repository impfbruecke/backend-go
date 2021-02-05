package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Sender interface {
	SendMessage(msg_to, msg_type, msg_body string) error
}

type TwillioSender struct {
	endpoint, user, token, from string
}

func NewTwillioSender(endpoint, user, token, from string) *TwillioSender {

	return &TwillioSender{
		endpoint: endpoint,
		token:    token,
		user:     user,
		from:     from,
	}
}

func (s TwillioSender) SendMessage(msg_to, msg_type, msg_body string) error {

	var err error
	var resp *http.Response
	var req *http.Request

	// Set request data
	v := url.Values{
		"To":   {msg_to},
		"From": {s.from},
	}

	v.Set("To", msg_to)
	v.Set("From", s.from)
	v.Set("Parameters", `{"type": "nachricht", "message":"`+url.QueryEscape(msg_body)+`"}`)

	// Pass the values to the request's body
	if req, err = http.NewRequest("POST", s.endpoint, strings.NewReader(v.Encode())); err != nil {
		return err
	}

	// Set authentication
	req.SetBasicAuth(s.user, s.token)

	// Send HTTP request
	client := &http.Client{}
	if resp, err = client.Do(req); err != nil {
		return err
	}

	// Log result or error
	if bodyText, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Warn(err)
	} else {
		log.Info(string(bodyText))
	}

	return nil
}
