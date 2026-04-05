// Package database provides database connection and initialization.
// Uses modernc.org/sqlite for a pure Go SQLite implementation (no CGO required).
package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // SQLite driver
)

// NewSQLiteDB creates and initializes a new SQLite database connection.
// It also runs the schema migration to ensure tables exist.
// dbPath: path to the SQLite database file (e.g., "./users.db")
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	// Open database connection
	// sqlite is the driver name registered by modernc.org/sqlite
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations to create tables
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations creates the necessary tables if they don't exist.
// In production, you might want to use a migration tool like golang-migrate.
func runMigrations(db *sql.DB) error {
	// Create users table
	// Stores OAuth user profiles from Google
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		picture TEXT,
		created_at DATETIME NOT NULL
	);

	-- Create index on email for faster lookups
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}
