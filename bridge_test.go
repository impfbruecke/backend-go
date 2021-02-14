package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	testfixtures "github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	fixtures *testfixtures.Loader
	sender   *TwillioSender
	loc      = time.FixedZone("+0100", 3600)

	fixtureCalls []Call = []Call{
		{
			ID:         1,
			Title:      "Call number 1",
			Capacity:   1,
			TimeStart:  time.Date(2021, time.February, 10, 12, 30, 0, 0, loc),
			TimeEnd:    time.Date(2021, time.February, 10, 12, 35, 0, 0, loc),
			YoungOnly:  true,
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
			Status:   false,
		},
		{
			Phone:    "1231",
			CenterID: 0,
			Group:    2,
			Status:   false,
		},
		{
			Phone:    "1232",
			CenterID: 0,
			Group:    1,
			Status:   true,
		},
	}
)

func TestMain(m *testing.M) {

	// Fix current time inside tests to:
	// 2021-01-01 20:00:00 +0000 UTC
	monkey.Patch(time.Now, func() time.Time { return time.Date(2021, 1, 1, 20, 0, 0, 0, time.UTC) })
	fmt.Println("Time is now ", time.Now())

	os.Exit(m.Run())
}

func prepareTestDatabase() {

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

	sender = NewTwillioSender("test", "test", "test", "test")

	bridge = &Bridge{
		db:     db,
		sender: sender,
	}

	// Open connection to the test database.
	// Do NOT import fixtures in a production database!
	// Existing data would be deleted.
	fixtures, err = testfixtures.New(
		testfixtures.Database(db.DB),                              // You database connection
		testfixtures.Dialect("sqlite"),                            // Available: "postgresql", "timescaledb", "mysql", "mariadb", "sqlite" and "sqlserver"
		testfixtures.Files("./testdata/fixtures/persons.yml"),     // the directory containing the YAML files
		testfixtures.Files("./testdata/fixtures/invitations.yml"), // the directory containing the YAML files
		testfixtures.Files("./testdata/fixtures/calls.yml"),       // the directory containing the YAML files
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
			},
			want: []Person{
				fixturePersons[0],
				fixturePersons[1],
				fixturePersons[2],
				{"0001", 0, 1, false},
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
				{"0001", 0, 1, false},
				{"0002", 0, 1, false},
			},
			want: []Person{

				fixturePersons[0],
				fixturePersons[1],
				fixturePersons[2],
				{"0001", 0, 1, false},
				{"0002", 0, 1, false},
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

	prepareTestDatabase()

	tests := []struct {
		name    string
		want    []Call
		wantErr bool
	}{
		{
			name: "Get two active calls",
			want: []Call{
				fixtureCalls[0],
				fixtureCalls[1],
			},
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
	type fields struct {
		db     *sqlx.DB
		sender *TwillioSender
	}
	type args struct {
		num    int
		callID int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Person
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
			got, err := b.GetNextPersonsForCall(tt.args.num, tt.args.callID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bridge.GetNextPersonsForCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bridge.GetNextPersonsForCall() = %v, want %v", got, tt.want)
			}
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
	tests := []struct {
		name string
		want []Call
	}{
		{
			name: "Get all calls after running deletion",
			want: []Call{
				fixtureCalls[0],
				fixtureCalls[1],
			},
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
	prepareTestDatabase()

	tests := []struct {
		name    string
		person  Person
		want    Call
		wantErr bool
	}{
		{"Phone 1230", fixturePersons[0], fixtureCalls[0], false},
		{"Phone 1231", fixturePersons[1], fixtureCalls[0], false},
		{"Phone 1232", fixturePersons[2], fixtureCalls[1], false},
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
	type fields struct {
		db     *sqlx.DB
		sender *TwillioSender
	}
	type args struct {
		phoneNumber string
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
			if err := b.PersonAcceptLastCall(tt.args.phoneNumber); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.PersonAcceptLastCall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridge_PersonCancelCall(t *testing.T) {
	type fields struct {
		db     *sqlx.DB
		sender *TwillioSender
	}
	type args struct {
		phoneNumber string
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
			if err := b.PersonCancelCall(tt.args.phoneNumber); (err != nil) != tt.wantErr {
				t.Errorf("Bridge.PersonCancelCall() error = %v, wantErr %v", err, tt.wantErr)
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
