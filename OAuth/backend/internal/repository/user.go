// Package repository provides data access layer for the application.
// This package handles all database operations for user management.
package repository

import (
	"context"
	"database/sql"
	"oauth-demo/internal/models"
	"time"
)

// UserRepository defines the interface for user data operations.
// Using an interface allows for easy testing with mocks.
type UserRepository interface {
	// FindByEmail looks up a user by their email address.
	// Returns sql.ErrNoRows if user doesn't exist.
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// Create inserts a new user into the database.
	// Returns the created user with the generated ID.
	Create(ctx context.Context, user *models.User) (*models.User, error)

	// FindByID looks up a user by their ID.
	FindByID(ctx context.Context, id int64) (*models.User, error)
}

// SQLiteUserRepository implements UserRepository using SQLite.
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository creates a new repository instance.
// Requires an initialized database connection.
func NewSQLiteUserRepository(db *sql.DB) UserRepository {
	return &SQLiteUserRepository{db: db}
}

// FindByEmail implements UserRepository.FindByEmail
func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, name, picture, created_at FROM users WHERE email = ?`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Picture,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Create implements UserRepository.Create
// Automatically sets the CreatedAt timestamp.
func (r *SQLiteUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.CreatedAt = time.Now().UTC()

	query := `INSERT INTO users (email, name, picture, created_at) VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, user.Email, user.Name, user.Picture, user.CreatedAt)
	if err != nil {
		return nil, err
	}

	user.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByID implements UserRepository.FindByID
func (r *SQLiteUserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, name, picture, created_at FROM users WHERE id = ?`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Picture,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
