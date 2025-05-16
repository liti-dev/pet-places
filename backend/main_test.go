package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var dbWrapper *DB

func setupTestServer(t *testing.T) (*server, *http.ServeMux) {
	mux := http.NewServeMux()
	// & returns a pointer to the server struct
	srv := &server{db: dbWrapper, router: mux}
	srv.routes(logMiddleware(slog.Default()))
	return srv, mux
}

// integration testing with actual database.
func setupTestDB(t *testing.T) *sql.DB {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
}
	testConnStr := os.Getenv("TEST_DATABASE_URL")
	
	if testConnStr == "" {
		t.Fatalf("TEST_DATABASE_URL is not set")
	}
	fmt.Println("Connecting to test database:", testConnStr)
	db, err := sql.Open("postgres", testConnStr)
	

	
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.Ping()
	if err!=nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	dbWrapper := &DB{db}

	// Clean existing database
	_, err = db.Exec("DROP TABLE IF EXISTS places")
    if err != nil {
        t.Fatalf("Failed to clean up test database: %v", err)
    }

	err = dbWrapper.createPlaceTable()
    if err != nil {
        t.Fatalf("Failed to create table: %v", err)
    }


    insertDataQuery := `
    INSERT INTO places (name, address, description) VALUES
    ('Place1', 'Address1', 'Description1'),
    ('Place2', 'Address2', 'Description2');`
    _, err = db.Exec(insertDataQuery)
    if err != nil {
        t.Fatalf("Failed to insert test data: %v", err)
    }

	return db
}

func TestGetPlacesAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Setup HTTP Server
	_, mux := setupTestServer(t)

	// Create a request to the /places endpoint
	req, err := http.NewRequest(http.MethodGet, "/places", nil)
	if err != nil {
		t.Fatalf("Failed to get places: %v", err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestGetPlaceAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test place into the database
	testPlace := Place{
		Name:        "Test Place",
		Address:     "123 Avenue",
		Description: "abc",
	}
	var err error
	testPlace.ID, err = dbWrapper.createPlaceDB(testPlace)
	if err != nil {
		t.Fatalf("Failed to insert place: %v", err)
	}

	// Setup HTTP Server
	_, mux := setupTestServer(t)

	// Create a request to the /places/{id} endpoint
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/places/%d", testPlace.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestCreatePlaceAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Setup HTTP Server
	_, mux := setupTestServer(t)

	// Create a new place
	newPlace := Place{
		Name:        "New Place",
		Address:     "456 Street",
		Description: "xyz",
	}
	body, err := json.Marshal(newPlace)
	if err != nil {
		t.Fatalf("Failed to marshal place: %v", err)
	}

	// Create a request to the /places endpoint
	req, err := http.NewRequest(http.MethodPost, "/places", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}
}

func TestUpdatePlaceAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test place into the database
	testPlace := Place{
		Name:        "Test Place",
		Address:     "123 Avenue",
		Description: "abc",
	}
	var err error
	testPlace.ID, err = dbWrapper.createPlaceDB(testPlace)
	if err != nil {
		t.Fatalf("Failed to insert place: %v", err)
	}

	// Setup HTTP Server
	_, mux := setupTestServer(t)

	// Update the place
	updatedPlace := Place{
		Name:        "Updated Place",
		Address:     "789 Boulevard",
		Description: "updated description",
	}
	body, err := json.Marshal(updatedPlace)
	if err != nil {
		t.Fatalf("Failed to marshal place: %v", err)
	}

	// Create a request to the /places/{id} endpoint
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/places/%d", testPlace.ID), bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}

func TestDeletePlaceAPI(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert a test place into db
	testPlace := Place{
		Name:        "Test Place",
		Address:     "123 Avenue",
		Description: "abc",
	}
	var err error
	testPlace.ID, err = dbWrapper.createPlaceDB(testPlace)
	if err != nil {
		t.Fatalf("Failed to insert place: %v", err)
	}

	// Setup HTTP Server
	_, mux := setupTestServer(t)

	// Create a request to the /places/{id} endpoint
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/places/%d", testPlace.ID), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Record the response
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}
}
