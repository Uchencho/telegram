package testutils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	// Sqlite3
	_ "github.com/mattn/go-sqlite3"
)

const (
	basicToken = "6cf457aafeb3128c99fd3d0d8267a9a9462cecfe58d80460be67aa059c9cdb9b"
)

// CreateDB returns an sqlite db for testing
func CreateDB() *sql.DB {
	_, err := os.Create("test_database.db")
	if err != nil {
		log.Println("Error in creating test db")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "test_database.db")
	if err != nil {
		log.Fatal("Failed to open DB with err", err)
	}
	log.Println("DB Open")
	return db
}

// GetTestDriver returns a mysql testdriver for migration
func GetTestDriver(db *sql.DB) database.Driver {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal("Could not get sqlite2 driver for test", err)
	}
	return driver
}

// DropDB returns the functionality to drop the DB
func DropDB() {
	log.Println("Called dropDB")
	err := os.Remove("test_database.db")
	if err != nil {
		log.Println("File does not exist")
	}
}

// FileToStruct unmarshals a json file into s truct
func FileToStruct(filepath string, s interface{}) io.Reader {
	bb, _ := ioutil.ReadFile(filepath)
	json.Unmarshal(bb, s)
	return bytes.NewReader(bb)
}

// NewTestServer creates a test server for testing
func NewTestServer(h http.HandlerFunc) (string, func()) {
	ts := httptest.NewServer(h)
	return ts.URL, func() { ts.Close() }
}

// SetTestStandardHeaders sets standard headers for testing
func SetTestStandardHeaders(r *http.Request) {
	h := r.Header
	h["Authorization"] = []string{fmt.Sprintf("Bearer %s", basicToken)}
	err := os.Setenv("BASIC_TOKEN", "6cf457aafeb3128c99fd3d0d8267a9a9462cecfe58d80460be67aa059c9cdb9b")
	if err != nil {
		log.Fatal(err)
	}
}

// GetResponseBody returns the unmarshalled response body
func GetResponseBody(res *http.Response, responseBody interface{}) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(body, responseBody)
}
