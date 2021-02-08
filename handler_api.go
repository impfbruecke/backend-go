package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"

	"encoding/json"
	"github.com/gorilla/mux"
)

func handlerApi(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		log.Info("FORM1:")
		r.ParseForm()

		for key, value := range r.Form {
			log.Printf("%s = %s\n", key, value)
		}

		log.Info("FORM2:")
		log.Println(r.Form)

		decoder := json.NewDecoder(r.Body)
		var t map[string]string
		err := decoder.Decode(&t)
		if err != nil {
			panic(err)
		}

		log.Println("the json")
		log.Println(t)
		log.Println("the json end")

		if phoneNumber, ok := t["number"]; ok {
			w.WriteHeader(http.StatusOK)
			switch mux.Vars(r)["endpoint"] {
			case "ja":
				if err := bridge.PersonAcceptCall(phoneNumber); err != nil {
					log.Error(err)
				}
			case "storno":
				if err := bridge.PersonCancelCall(phoneNumber); err != nil {
					log.Error(err)
				}
			case "loeschen":
				if err := bridge.PersonDelete(phoneNumber); err != nil {
					log.Error(err)
				}
			default:
				log.Debug("Invalid request to API recieved")
				io.WriteString(w, "Invalid request")
			}
		}

	} else {
		io.WriteString(w, "Invalid request")
	}
}

// info msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: 8ca16db3-93eb-4cc5-b61e-1c02a21a3711\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: WoFNFzzu70y3GfXatgIFQigvOJ4=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
// info msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nAuthorization: Basic aW1wZmVuOm9vYWdncmd1c3dyeGRpeW9tbWJucndocHNpanlkeWRrdmVpaGZzbmZ5c2J1\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: 8ca16db3-93eb-4cc5-b61e-1c02a21a3711\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: WoFNFzzu70y3GfXatgIFQigvOJ4=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
// info msg="FORM1:"
// info msg="bodySHA256 = [952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d]\n"
// info msg="FORM2:"
// info msg="map[bodySHA256:[952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d]]"
// info msg="the json"
// info msg="map[number:+491727723996]"
// info msg="the json end"
// debug msg="Accepting call"
// info msg="POST /notify/storno?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /notify/storno?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: 26d8a0c4-35b4-4572-9ddf-c848b630eec3\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: x49wT6Y2r7fyGeYbFPbt9MTALOI=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"

// info msg="POST /notify/loeschen?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /notify/loeschen?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: b3f04bee-46eb-4382-8ce1-31d79f2e242f\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: uDdMtHOXZflC8vK31J9XY2/WSzM=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
// info msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: bd59bc0e-5c50-4c6f-a05a-3a422bc078a1\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: WoFNFzzu70y3GfXatgIFQigvOJ4=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
// info msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /api/ja?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nAuthorization: Basic aW1wZmVuOm9vYWdncmd1c3dyeGRpeW9tbWJucndocHNpanlkeWRrdmVpaGZzbmZ5c2J1\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: bd59bc0e-5c50-4c6f-a05a-3a422bc078a1\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: WoFNFzzu70y3GfXatgIFQigvOJ4=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
// info msg="FORM1:"
// info msg="bodySHA256 = [952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d]\n"
// info msg="FORM2:"
// info msg="map[bodySHA256:[952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d]]"
// info msg="the json"
// info msg="map[number:+491727723996]"
// info msg="the json end"
// debug msg="Accepting call"
// info msg="POST /notify/storno?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /notify/storno?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: e9cbbb30-2a7d-45d2-81d6-96089d516a34\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: x49wT6Y2r7fyGeYbFPbt9MTALOI=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
// info msg="POST /notify/loeschen?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d"
// debug msg="POST /notify/loeschen?bodySHA256=952c14cce02e6601946f59286fa8eab0dfa2c77c843c6f784eb39e6fcff50d6d HTTP/1.0\r\nHost: 127.0.0.1:12000\r\nConnection: close\r\nAccept: */*\r\nConnection: close\r\nContent-Length: 29\r\nContent-Type: application/json\r\nI-Twilio-Idempotency-Token: 56da2d13-5378-4a7a-94bd-210147b1179a\r\nUser-Agent: TwilioProxy/1.1\r\nX-Twilio-Signature: uDdMtHOXZflC8vK31J9XY2/WSzM=\r\n\r\n{\n\"number\": \"+491727723996\"\n}"
