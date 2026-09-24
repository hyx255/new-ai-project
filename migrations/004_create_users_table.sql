-- Migration: 004_create_users_table.sql
-- Purpose: Create users table for authentication and user management
-- Database: MySQL / GreatDB / SQLite compatible

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(64) PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'USER',
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    must_change_password BOOLEAN NOT NULL DEFAULT 1,
    token_version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Default admin user
-- Username: admin
-- Password: Admin123 (must change on first login)
INSERT INTO users (id, username, password_hash, role, status, must_change_password, token_version, created_at, updated_at)
VALUES ('usr_admin_default_001', 'admin', '$2a$10$ooWJib0ehOTtYEm.DRU.9uiG46anImtfLj40YB9cRg5G/LtkxefXK', 'ADMIN', 'ACTIVE', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);