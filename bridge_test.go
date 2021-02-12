package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	testfixtures "github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db       *sql.DB
	fixtures *testfixtures.Loader
	sender   *TwillioSender
)

func TestMain(m *testing.M) {
	var err error

	// exist. Exit application on errors, we can't continue without database
	db, err := sqlx.Connect("sqlite3", "./test.db")

	db.SetMaxOpenConns(1)
	if err != nil {
		panic(err)
	}

	fmt.Println("creating sender")
	sender = NewTwillioSender("test", "test", "test", "test")

	fmt.Println("creating bridge")
	bridge = &Bridge{
		db:     db,
		sender: sender,
	}

	// Open connection to the test database.
	// Do NOT import fixtures in a production database!
	// Existing data would be deleted.

	fmt.Println("creating fixtures")
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

	os.Exit(m.Run())
}

func prepareTestDatabase() {
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
			name: "Retrieve persons from DB",
			want: []Person{
				{Phone: "1230", CenterID: 0, Group: 1, Status: false},
				{Phone: "1231", CenterID: 0, Group: 1, Status: false},
				{Phone: "1232", CenterID: 0, Group: 1, Status: false},
			},
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
				{Phone: "1230", Group: 1},
				{Phone: "1231", Group: 1},
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
