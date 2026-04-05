-- OAuth 2.0 Demo - Database Schema
-- SQLite database for user management

-- ============================================
-- Users Table
-- ============================================
-- Stores user profiles fetched from Google OAuth
-- Users are automatically created on first login

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    picture TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index on email for faster lookups during authentication
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ============================================
-- Optional: Refresh Tokens Table (for refresh token support)
-- ============================================
-- Uncomment if you want to implement refresh tokens

-- CREATE TABLE IF NOT EXISTS refresh_tokens (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     user_id INTEGER NOT NULL,
--     token_hash TEXT UNIQUE NOT NULL, -- Store hashed token, not plain
--     expires_at DATETIME NOT NULL,
--     created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
-- );

-- CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token_hash);
-- CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);

-- ============================================
-- Sample Query Examples
-- ============================================

-- Find user by email (used during OAuth callback)
-- SELECT id, email, name, picture, created_at FROM users WHERE email = ?;

-- Create new user (used when user doesn't exist)
-- INSERT INTO users (email, name, picture, created_at) VALUES (?, ?, ?, ?);

-- Find user by ID (used by /api/me endpoint)
-- SELECT id, email, name, picture, created_at FROM users WHERE id = ?;
