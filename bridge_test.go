package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	testfixtures "github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Custom type that allows setting the func that our Mock Do func will run
// instead
type MockClient struct {
	MockDo func(req *http.Request) (*http.Response, error) // MockClient is the mock client
}

// Overriding what the Do function should "do" in our MockClient
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

var (
	fixtures *testfixtures.Loader
	sender   *TwillioSender
	loc      = time.FixedZone("+0100", 3600)

	fakeNow time.Time = time.Date(2999, 1, 1, 20, 0, 0, 0, time.UTC)

	fixtureCalls []Call = []Call{
		{
			ID:         1,
			Title:      "Call number 1",
			Capacity:   1,
			TimeStart:  time.Date(2021, time.February, 10, 12, 30, 0, 0, loc),
			TimeEnd:    time.Date(2021, time.February, 10, 12, 35, 0, 0, loc),
			AgeMin:     0,
			AgeMax:     100,
			LocName:    "loc_name1",
			LocStreet:  "loc_street1",
			LocHouseNr: "loc_housenr1",
			LocPLZ:     "loc_plz1",
			LocCity:    "loc_city1",
			LocOpt:     "loc_opt1",
		},
		{
			ID:         2,
			Title:      "Call number 2",
			Capacity:   2,
			TimeStart:  time.Date(2021, time.February, 10, 12, 31, 0, 0, loc),
			TimeEnd:    time.Date(2021, time.February, 10, 12, 36, 0, 0, loc),
			AgeMin:     0,
			AgeMax:     70,
			LocName:    "loc_name2",
			LocStreet:  "loc_street2",
			LocHouseNr: "loc_housenr2",
			LocPLZ:     "loc_plz2",
			LocCity:    "loc_city2",
		},
		{
			ID:         3,
			Title:      "Call number 3",
			Capacity:   3,
			TimeStart:  time.Date(2021, time.January, 1, 12, 30, 0, 0, loc),
			TimeEnd:    time.Date(2021, time.January, 1, 12, 35, 0, 0, loc),
			AgeMin:     70,
			AgeMax:     200,
			LocName:    "loc_name3",
			LocStreet:  "loc_street3",
			LocHouseNr: "loc_housenr3",
			LocPLZ:     "loc_plz3",
			LocCity:    "loc_city3",
			LocOpt:     "loc_opt3",
		},
	}
	fixturePersons []Person = []Person{
		{
			Phone:    "1230",
			CenterID: 0,
			Group:    1,
			Age:      10,
			Status:   false,
		},
		{
			Phone:    "1231",
			CenterID: 0,
			Group:    2,
			Age:      70,
			Status:   false,
		},
		{
			Phone:    "1232",
			CenterID: 0,
			Group:    1,
			Age:      150,
			Status:   true,
		},
	}

	fixtureInvitations []Invitation = []Invitation{
		{
			Phone:  "1230",
			CallID: 1,
			Status: "accepted",
			Time:   time.Date(2021, 2, 10, 12, 36, 0, 0, loc),
		},
		{
			ID:     1,
			Phone:  "1231",
			CallID: 1,
			Status: "accepted",
			Time:   time.Date(2021, 2, 10, 12, 36, 0, 0, loc),
		},
		{
			ID:     2,
			Phone:  "1232",
			CallID: 2,
			Status: "rejected",
			Time:   time.Date(2021, 2, 10, 12, 36, 0, 0, loc),
		},
	}
)

var HTTPResponse string

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string // Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)                             // Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host)) // Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		// request = append(request, "Formdata:")
		request = append(request, r.Form.Encode())
	} // Return the request as a string
	return strings.Join(request, "\n")
}

func TestMain(m *testing.M) {

	Client = &MockClient{
		MockDo: func(req *http.Request) (*http.Response, error) {
			fmt.Printf("Faking request: \n%s\n", formatRequest(req))
			fmt.Printf("Faking response: %s\n", HTTPResponse)
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(HTTPResponse))),
			}, nil
		},
	}

	// Fix current time inside tests to:
	// 2021-01-01 20:00:00 +0000 UTC
	monkey.Patch(time.Now, func() time.Time { return time.Date(2999, 1, 1, 20, 0, 0, 0, time.UTC) })
	fmt.Println("Time is now fixed to:", time.Now())

	os.Exit(m.Run())
}

func prepareTestDatabase(fixtureFiles ...string) {

	os.Setenv("IMPF_DISABLE_SMS", "")
	os.Setenv("IMPF_MODE", "DEVEL")
	os.Setenv("IMPF_SESSION_SECRET", "session_secret")
	os.Setenv("IMPF_TWILIO_API_ENDPOINT", "https://studio.twilio.com/v2/Flows/")
	os.Setenv("IMPF_TWILIO_API_FROM", "twilio_api_from")
	os.Setenv("IMPF_TWILIO_API_PASS", "twilio_api_pass")
	os.Setenv("IMPF_TWILIO_API_USER", "twilio_api_user")
	os.Setenv("IMPF_TWILIO_PASS", "twilio_pass")
	os.Setenv("IMPF_TWILIO_USER", "twilio_user")

	var err error

	if _, err := os.Stat("./test.db"); err == nil {
		// Old DB exists, try to remove it
		fmt.Println("Removing old testDB")
		err = os.Remove("./test.db")
		if err != nil {
			panic(err)
		}
	}

	// exist. Exit application on errors, we can't continue without database
	db, err := sqlx.Connect("sqlite3", "./test.db")

	db.SetMaxOpenConns(1)
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating schemas from scratch")
	db.MustExec(schemaCalls)
	db.MustExec(schemaPersons)
	db.MustExec(schemaUsers)
	db.MustExec(schemaNotifications)

	sender = NewTwillioSender(
		os.Getenv("IMPF_TWILIO_API_ENDPOINT"),
		os.Getenv("IMPF_TWILIO_API_USER"),
		os.Getenv("IMPF_TWILIO_API_PASS"),
		os.Getenv("IMPF_TWILIO_API_FROM"),
	)

	bridge = &Bridge{
		db:     db,
		sender: sender,
	}

	// Open connection to the test database.
	// Do NOT import fixtures in a production database!
	// Existing data would be deleted.

	// Set default fixtures if none where specified
	if len(fixtureFiles) == 0 {
		fixtureFiles = []string{
			"testdata/fixtures/calls.yml",
			"testdata/fixtures/invitations.yml",
			"testdata/fixtures/persons.yml",
		}
	}

	fixtures, err = testfixtures.New(
		testfixtures.Database(db.DB),        // You database connection
		testfixtures.Dialect("sqlite"),      // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Files(fixtureFiles...), // the directory containing the YAML files
	)
	if err != nil {
		panic(err)
	}

	if err := fixtures.Load(); err != nil {
		fmt.Println("Loading fixtures")
		panic(err)
	}

}

func TestBridge_GetPersons(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name    string
		want    []Person
		wantErr bool
	}{
		{
			name:    "Retrieve persons from DB",
			want:    fixturePersons,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetPersons()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetPersons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_GetAcceptedPersons(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name    string
		id      int
		want    []Person
		wantErr bool
	}{
		{
			name: "Call with accepted invitations",
			id:   1,
			want: []Person{
				fixturePersons[0],
				fixturePersons[1],
			},
			wantErr: false,
		},
		{
			name:    "Call without rejected invitations",
			id:      2,
			want:    []Person{},
			wantErr: false,
		},
		{
			name:    "Call without invitations",
			id:      3,
			want:    []Person{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetAcceptedPersons(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetAcceptedPersons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("MakeGatewayInfo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_AddPerson(t *testing.T) {
	prepareTestDatabase()
	tests := []struct {
		name    string
		person  Person
		want    []Person
		wantErr bool
	}{
		{
			name: "Add a valid person",
			person: Person{
				Phone:    "0001",
				CenterID: 0,
				Group:    1,
				Status:   false,
				Age:      80,
			},
			want: []Person{
				fixturePersons[0],
				fixturePersons[1],
				fixturePersons[2],
				{"0001", 0, 1, false, 80},
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := bridge.AddPerson(tt.person); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.AddPerson() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := bridge.GetPersons()
			if err != nil {
				t.Errorf("GetPersons() after AddPersion() failed with error: %v \n", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetPersons() after AddPersion() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

func TestBridge_AddPersons(t *testing.T) {
	prepareTestDatabase()
	tests := []struct {
		name    string
		persons []Person
		want    []Person
		wantErr bool
	}{
		{
			name: "Add two persons",
			persons: []Person{
				{"0001", 0, 1, false, 40},
				{"0002", 0, 1, false, 90},
			},
			want: []Person{

				fixturePersons[0],
				fixturePersons[1],
				fixturePersons[2],
				{"0001", 0, 1, false, 40},
				{"0002", 0, 1, false, 90},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := bridge.AddPersons(tt.persons); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.AddPersons() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := bridge.GetPersons()
			if err != nil {
				t.Errorf("GetPersons() after AddPersons() failed with error: %v \n", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetPersons() after AddPersons() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_GetCallStatus(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name    string
		id      string
		want    CallStatus
		wantErr bool
	}{
		{
			name: "Get a valid callstatus",
			id:   "1",
			want: CallStatus{
				Call: fixtureCalls[0],
				Persons: []Person{
					fixturePersons[0],
					fixturePersons[1],
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetCallStatus(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetCallStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.GetPersons() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_GetActiveCalls(t *testing.T) {

	prepareTestDatabase("testdata/fixtures/TestBridge_GetActiveCalls/calls.yml")

	calls, err := bridge.GetAllCalls()

	if err != nil {
		panic(err)
	}

	tests := []struct {
		name    string
		want    []Call
		wantErr bool
	}{
		{
			name:    "Get two active calls",
			want:    []Call{calls[0], calls[1]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetActiveCalls()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetActiveCalls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.GetActive() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_GetNextPersonsForCall(t *testing.T) {

	var fixtures *testfixtures.Loader
	var err error
	var allPersons []Person

	fixtures, err = testfixtures.New(
		testfixtures.Database(bridge.db.DB),                                             // You database connection
		testfixtures.Dialect("sqlite"),                                                  // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Files("./testdata/fixtures/getNextPersonsForCall/persons.yml"),     // the directory containing the YAML files
		testfixtures.Files("./testdata/fixtures/getNextPersonsForCall/invitations.yml"), // the directory containing the YAML files
		testfixtures.Files("./testdata/fixtures/getNextPersonsForCall/calls.yml"),       // the directory containing the YAML files
	)
	if err != nil {
		panic(err)
	}

	allPersons, err = bridge.GetPersons()
	if err != nil {
		panic(err)
	}

	if err := fixtures.Load(); err != nil {
		fmt.Println("Loading fixtures")
		panic(err)
	}
	tests := []struct {
		name    string
		num     int
		callID  int
		wantErr bool
	}{
		// 11 persons total in fixtures
		// call ID 0: no age restriction
		// call ID 1: withage restriction
		{"Call without age restriction, 5/11 persons", 5, 0, false},
		{"Call without age restriction, 20/11 persons", 20, 0, false},
		{"Call with age restriction, 5/11 persons", 5, 1, false},
		{"Call with age restriction, 20/11 persons", 20, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetNextPersonsForCall(tt.num, tt.callID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetNextPersonsForCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Get the call with the ID specified
			call, err := bridge.GetCallStatus(strconv.Itoa(tt.callID))
			if err != nil {
				panic(err)
			}

			invitations, err := bridge.GetInvitations()
			if err != nil {
				panic(err)
			}

			// Find highest selected group in results
			highestGroup := 0
			for _, v := range got {
				if v.Group > highestGroup {
					highestGroup = v.Group
				}
			}

			lowerPersonNum := 0
			// Find how many persons are in the DB with group number under highestGroup
			for _, v := range allPersons {
				if v.Group > highestGroup {
					lowerPersonNum += 1
				}
			}

			// Lower group numbers should always go first. If the number of
			// persons with group nubmer below the highest returned group is
			// greater or equal of to the number we are looking for, the result
			// is wrong. A call should first be issued to alle the persons of
			// lower groups before trying the next one up
			if lowerPersonNum >= tt.num {
				t.Error("Bridge.GetNextPersonsForCall selected persons with higher group than necessary")
				return
			}

			// if diff := cmp.Diff(tt.want, got); diff != "" {
			// 	t.Errorf("Bridge.GetNextPersonsForCall() mismatch (-want +got):\n%s", diff)
			// }

			// Check there are no duplicates in the persons returned. We place
			// the returned slice in a map, with the phone as key and a
			// arbitrary value (1). For each person, we try to get it by
			// accessing the map, if it fails, we add it (this is good and
			// means the person was not yet in the map). If it succeds we exit
			// with an error, this means we tried to access a key (phonenumber)
			// that already exists in the map, meaning that it is duplicate
			duplicate_frequency := make(map[string]int)
			for _, pers := range got {
				if _, ok := duplicate_frequency[pers.Phone]; ok {
					t.Error("Bridge.GetNextPersonsForCall returned duplicates")
				} else {
					duplicate_frequency[pers.Phone] = 1
				}
			}

			// Check if less than the requested number of persons where
			// returned, even though we had enough to choose from
			if len(got) < tt.num && len(allPersons) > tt.num {
				t.Error("Bridge.GetNextPersonsForCall returned not enough persons, even though available")
				return
			}

			// Check the number of persons returned does not exist the number requested
			if len(got) > tt.num {
				t.Error("Bridge.GetNextPersonsForCall returned too many persons")
				return
			}

			for _, v := range got {
				// check none is already vaccinated
				if v.Status {
					t.Errorf("Bridge.GetNextPersonsForCall returned already vaccinated person Phone: %v\n", v.Phone)
					return
				}

				// check if all persons age criteria
				if v.Age > call.Call.AgeMax || v.Age < call.Call.AgeMin {
					t.Errorf("Bridge.GetNextPersonsForCall returned person outside of allowed age range, phone: %v\n", v.Phone)
					return
				}
			}

			// check not already notified persons are retrieved
			for _, i := range invitations {
				for _, p := range got {
					if i.Phone == p.Phone && i.CallID == tt.callID {
						t.Errorf("Bridge.GetNextPersonsForCall returned phone %s already notified for call: %v\n", p.Phone, tt.callID)
						return
					}
				}
			}

			// TODO check: order within one group should be randomized
		})
	}
}

func TestNewBridge(t *testing.T) {
	tests := []struct {
		name string
		want *Bridge
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBridge(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBridge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBridge_DeleteOldCalls(t *testing.T) {

	prepareTestDatabase("testdata/fixtures/TestBridge_DeleteOldCalls/calls.yml")

	calls, err := bridge.GetAllCalls()
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name string
		want []Call
	}{
		{
			name: "Get all calls after running deletion",
			want: []Call{calls[2]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridge.DeleteOldCalls()
		})

		got, err := bridge.GetAllCalls()
		if err != nil {
			t.Errorf("Bridge.GetAllCalls() error = %v", err)
			return
		}

		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("Bridge.GetActive() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestBridge_SendNotifications(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		sender *TwillioSender
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Bridge{
				db:     tt.fields.db,
				sender: tt.fields.sender,
			}
			b.SendNotifications()
		})
	}
}

func TestBridge_NotifyCall(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name       string
		callID     int
		numPersons int
		wantErr    bool
	}{

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := bridge.NotifyCall(tt.callID, tt.numPersons); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.NotifyCall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridge_CallFull(t *testing.T) {
	prepareTestDatabase()

	tests := []struct {
		name    string
		call    Call
		want    bool
		wantErr bool
	}{
		{"Get full call (ID:1)", fixtureCalls[0], true, false},
		{"Get not full call (ID:2)", fixtureCalls[1], false, false},
		{"Get not full call (ID:3)", fixtureCalls[2], false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.CallFull(tt.call)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.CallFull() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Bridge.CallFull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBridge_LastCallNotified(t *testing.T) {

	prepareTestDatabase(
		"testdata/fixtures/TestBridge_LastCallNotified/invitations.yml",
		"testdata/fixtures/TestBridge_LastCallNotified/calls.yml")

	tests := []struct {
		name    string
		person  Person
		want    Call
		wantErr bool
	}{
		{"Phone 0", Person{Phone: "0"}, fixtureCalls[1], false},
		{"Phone 1", Person{Phone: "1"}, fixtureCalls[2], false},
		{"Phone 2", Person{Phone: "2"}, fixtureCalls[1], false},
		{"Phone 3", Person{Phone: "3"}, fixtureCalls[1], false},
		{"Phone 4", Person{Phone: "4"}, fixtureCalls[1], false},
		{"Phone noexist", Person{Phone: "noexist"}, Call{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.LastCallNotified(tt.person)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.LastCallNotified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.LastCallNotified() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_PersonAcceptLastCall(t *testing.T) {

	prepareTestDatabase(
		"testdata/fixtures/TestBridge_PersonAcceptLastCall/invitations.yml",
		"testdata/fixtures/TestBridge_PersonAcceptLastCall/calls.yml",
	)

	gotBefore, err := bridge.GetInvitations()

	if err != nil {
		panic(err)
	}

	tests := []struct {
		name        string
		phoneNumber string
		want        []Invitation
		wantErr     bool
	}{
		{"Phone 1230", "1230", gotBefore, false},
		{"Phone 1231", "1231", []Invitation{
			{
				ID:     1,
				Phone:  "1231",
				CallID: 1,
				Status: InvitationAccepted,
				Time:   time.Now(),
			},
			gotBefore[0],
			gotBefore[2],
			gotBefore[3],
			gotBefore[4],
		}, false},
		{"Phone 1232", "1232", gotBefore, false},
		{"Phone noexist", "noexist", gotBefore, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Reset database to initial values on each test, since the tests
			// change the contents
			prepareTestDatabase(
				"testdata/fixtures/TestBridge_PersonAcceptLastCall/invitations.yml",
				"testdata/fixtures/TestBridge_PersonAcceptLastCall/calls.yml",
			)

			if err := bridge.PersonAcceptLastCall(tt.phoneNumber); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.PersonAcceptLastCall() error = %v, wantErr %v", err, tt.wantErr)
			}

			gotAfter, err := bridge.GetInvitations()
			if err != nil {
				panic(err)
			}

			if diff := cmp.Diff(tt.want, gotAfter); diff != "" {
				t.Errorf("Bridge.GetInvitations() after PersonAcceptLastCall(%s) mismatch (-want +got):\n%s", tt.phoneNumber, diff)
			}
		})
	}
}

func TestBridge_PersonDelete(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name        string
		phoneNumber string
		want        []Person
		wantErr     bool
	}{
		{"Remove 1230", "1230", []Person{fixturePersons[1], fixturePersons[2]}, false},
		{"Remove 1231", "1231", []Person{fixturePersons[2]}, false},
		{"Remove 1232", "1232", []Person{}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := bridge.PersonDelete(tt.phoneNumber); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.PersonDelete() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := bridge.GetPersons()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetPersons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.GetPersons() after deleting mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

func TestBridge_AddCall(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		sender *TwillioSender
	}
	type args struct {
		call Call
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bridge{
				db:     tt.fields.db,
				sender: tt.fields.sender,
			}
			if err := b.AddCall(tt.args.call); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.AddCall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridge_GetAllCalls(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name    string
		want    []Call
		wantErr bool
	}{
		{
			name:    "Retrieve calls from DB",
			want:    fixtureCalls,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetAllCalls()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetAllCalls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.GetAllCalls() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_GetInvitations(t *testing.T) {

	prepareTestDatabase()

	tests := []struct {
		name    string
		want    []Invitation
		wantErr bool
	}{
		{"Retrieve all invitations from DB", fixtureInvitations, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bridge.GetInvitations()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetInvitations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.GetInvitations() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridge_PersonCancelAllCalls(t *testing.T) {
	prepareTestDatabase("testdata/fixtures/TestBridge_PersonCancelAllCalls/invitations.yml")

	gotBefore, err := bridge.GetInvitations()
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name        string
		phoneNumber string
		want        []Invitation
		wantErr     bool
	}{
		{
			"Cancel Calls of phone 1230",
			"1230",
			[]Invitation{
				{
					ID:     0,
					Phone:  "1230",
					CallID: 1,
					Status: "cancelled",
					Time:   fakeNow,
				},
				{
					ID:     1,
					Phone:  "1230",
					CallID: 2,
					Status: "cancelled",
					Time:   fakeNow,
				},
				gotBefore[2],
				gotBefore[3],
				gotBefore[4],
				gotBefore[5],
				gotBefore[6],
				gotBefore[7],
			},
			false},
		{
			"Cancel Calls of phone 1231",
			"1231",
			[]Invitation{
				{
					ID:     0,
					Phone:  "1230",
					CallID: 1,
					Status: "cancelled",
					Time:   fakeNow,
				},
				{
					ID:     1,
					Phone:  "1230",
					CallID: 2,
					Status: "cancelled",
					Time:   fakeNow,
				},

				{
					ID:     4,
					Phone:  "1231",
					CallID: 1,
					Status: "cancelled",
					Time:   fakeNow,
				},
				{
					ID:     5,
					Phone:  "1231",
					CallID: 2,
					Status: "cancelled",
					Time:   fakeNow,
				},
				gotBefore[2],
				gotBefore[3],
				gotBefore[6],
				gotBefore[7],
			},
			false},
		{"Cancel Calls of phone noexist", "noexist", []Invitation{
			{
				ID:     0,
				Phone:  "1230",
				CallID: 1,
				Status: "cancelled",
				Time:   fakeNow,
			},
			{
				ID:     1,
				Phone:  "1230",
				CallID: 2,
				Status: "cancelled",
				Time:   fakeNow,
			},

			{
				ID:     4,
				Phone:  "1231",
				CallID: 1,
				Status: "cancelled",
				Time:   fakeNow,
			},
			{
				ID:     5,
				Phone:  "1231",
				CallID: 2,
				Status: "cancelled",
				Time:   fakeNow,
			},
			gotBefore[2],
			gotBefore[3],
			gotBefore[6],
			gotBefore[7],
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := bridge.PersonCancelAllCalls(tt.phoneNumber); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.PersonCancelAllCalls() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := bridge.GetInvitations()
			if err != nil {
				panic(err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Bridge.GetInvitations() mismatch (-want +got):\n%s", diff)
			}

		})
	}
}
