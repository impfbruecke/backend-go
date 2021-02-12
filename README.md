# Impfbruecke


[![tests](https://github.com/impfbruecke/backend-go/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/impfbruecke/backend-go/branch/main/graph/badge.svg)](https://codecov.io/gh/impfbruecke/backend-go)


## Endpoints

### GET /
Redirect to `/call` for convenience

#### Parameters:
none

### GET /static
Serves static files like css and images from the `/static` directory. Place any file there you want make publicly avialable

#### Parameters:
none

### GET /call
Show `call.html` page, which allows to create new calls

#### Parameters:
none

### POST /call
Create a new call

#### Parameters:
TODO

### GET /call/{id}
Information about call with ID `{id}`

#### Parameters:
none

### GET /active
Show `active.html`, which displays active calls

#### Parameters:
none

### GET /add
Show `add.html`, which allows to add a single person to the database

#### Parameters:
none

### POST /add
Add person to database

#### Parameters:
TODO

### POST /upload
Upload `.csv` for bulk import of persons

#### Parameters:
TODO

### POST /api/ja
Listen for incoming webhook to accept a appointment

#### Parameters:
TODO

### POST /api/loeschen
Listen for incoming webhook to delete a single user from the database

#### Parameters:
TODO

### POST /api/storno
Listen for incoming webhook to cancel an appointment

#### Parameters:
TODO



## Twilio
Twilio Nummer:  +49 1573 5984785


Hab Twilio aufgesetzt:

Eingehende SMS werden folgt verarbeitet und triggern dann API Calls bei der Impfbruecke :

https://www.twilio.com/docs/studio/widget-library/http-request


| SMS Inhalt| JA| STORNO | LOESCHEN LÖSCHEN|
| ------ | ------ |  ------ | ------ | 
| API Call| /ja| /storno| /loeschen|

Der endpunkt ist über Base Auth möglich also "https://user:password@api.impfbruecke.de/api/ja"
Die payload ist überall die gleiche und beinhaltet die Absendernummer im [e.164](https://www.twilio.com/docs/glossary/what-e164) format 


```json
{
"from": "+49xxxxxxxxx"
}
```


Der Weg um SMS zu senden läuft auch über eine REST API

https://www.twilio.com/docs/studio/rest-api

Endpoint ist ein POST auf "https://studio.twilio.com/v2/Flows/FWa50a95a66d8639edd72ca466b3dfba32/Executions" mit den Parametern "To" und "From"  und Parameters=json

folgende ausgehende SMS sind möglich:

| Name | Anfrage | Absage | Zusage | Loeschbestaetigung |
|----|----|----|----|----|
|JSON | A| B | B | B |

folgende JSONs gibt es


A: 
Dieser JSON schicke ein Send&Wait.SMS inhalt in "message" packen und mit waitSeconds die Wartezeit definieren.
```json

{
"type": "anfrage",
"message":"string",
"waitSeconds":3600
}
```
waitSeconds: definiert den Wert wie lange auf eine Antwort gewartet wird. Wenn 'waitTime > waitSeconds' wird ein STORNO ausgeführt. Wenn eine Antwort kommt, wird diese wieder geparsed und bei JA kommt eine Bestätigung.




Dieser JSON kann einfach nur stupide SMS verschicken. SMS inhalt in "message" packen und spaß haben
B:
```json
{
"type": "nachricht",
"message":"string"
}

```
