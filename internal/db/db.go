// Package db provides SQLite database management for offline Kaiten data.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database connection.
type DB struct {
	SQL *sql.DB
}

// Open opens (or creates) a SQLite database at path, creates parent directories
// as needed, enables sane pragmas, and runs migrations idempotently.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	db := &DB{SQL: sqlDB}
	if err := db.enablePragmas(path); err != nil {
		db.Close()
		return nil, err
	}
	if err := db.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// OpenInMemory opens an in-memory SQLite database for testing.
func OpenInMemory() (*DB, error) {
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("opening in-memory sqlite: %w", err)
	}

	db := &DB{SQL: sqlDB}
	if err := db.enablePragmas(":memory:"); err != nil {
		db.Close()
		return nil, err
	}
	if err := db.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.SQL.Close()
}

func (db *DB) enablePragmas(path string) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
	}
	// Only enable WAL for file-based databases.
	if path != ":memory:" {
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL")
	}
	for _, p := range pragmas {
		if _, err := db.SQL.Exec(p); err != nil {
			return fmt.Errorf("pragmas: %w", err)
		}
	}
	return nil
}
