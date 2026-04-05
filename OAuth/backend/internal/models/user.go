// Package models defines the data structures used across the application.
// These models represent the core entities in our OAuth system.
package models

import "time"

// User represents a user in the system.
// This struct maps to the users table in SQLite and contains
// profile information fetched from Google OAuth.
type User struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Picture   string    `json:"picture" db:"picture"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GoogleUserInfo represents the user profile data returned by Google's UserInfo endpoint.
// This is the structure of the JSON response from https://openidconnect.googleapis.com/v1/userinfo
type GoogleUserInfo struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	EmailVerified bool   `json:"email_verified"`
	Sub           string `json:"sub"` // Google's unique identifier for the user
}

// Claims represents the JWT claims structure.
// These claims are embedded in the JWT token and used for authentication.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}
