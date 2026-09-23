// Package store handles database interactions and low-level CRUD for rss-reader.
package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Store provides access to the persistent storage layer.
type Store struct {
	db *sql.DB
}

// Open opens a sqlite instance at the provided path and returns a Store instance.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

// Close closes the database connection.
func (store *Store) Close() error {
	return store.db.Close()
}
