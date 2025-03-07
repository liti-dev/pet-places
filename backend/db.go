package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)
type DB struct {
	*sql.DB
}

// creates a new DB connection
func createDB(connStr string) (*DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) createPlaceTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS places (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			address VARCHAR(100) NOT NULL,
			description TEXT,
			created TIMESTAMP DEFAULT NOW()
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating table: %v\n", err)
		return err
	}
	return nil
}


func (db *DB) getPlacesDB(queryName string) ([]Place, error) {	
	rows, err := db.Query(`SELECT id, name, address, description
FROM places
WHERE LOWER(name) LIKE '%' || $1 || '%'`,queryName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	places := []Place{}
	for rows.Next() {
		var place Place
		if err := rows.Scan(&place.ID, &place.Name, &place.Address, &place.Description); err != nil {
			return nil, err
		}
		places = append(places, place)
	}
	return places, nil
}

func (db *DB) getPlaceDB(id int) (*Place, error) {
	var place Place
	err := db.QueryRow("SELECT id, name, address, description FROM places WHERE id = $1", id).Scan(&place.ID, &place.Name, &place.Address, &place.Description)
	if err != nil {
		return nil, fmt.Errorf("place not found")
	}
	return &place, nil
}


func (db *DB) createPlaceDB(place Place) (int, error) {
	query := `INSERT INTO places (name, address, description)
	VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := db.QueryRow(query, place.Name, place.Address, place.Description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting place: %w", err)
	}
	fmt.Print("Inserted ID=", id)
	return id, nil
}

func (db *DB) updatePlaceDB(place Place, id int) error {
	query := `UPDATE place SET name=$1, address=$2, description=$3 WHERE id=$4`
	_, err := db.Exec(query, place.Name, place.Address, place.Description, id)
	fmt.Print("Updated", place)
	return err
}

func (db *DB) deletePlaceDB(id int) error {
	query := `DELETE FROM places WHERE id=$1`
	_, err := db.Exec(query, id)
	fmt.Print("Deleted ID=", id)

	return err
}

