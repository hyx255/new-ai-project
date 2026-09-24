-- Migration: 001_create_device_tables.sql
-- Purpose: Create device_types and devices tables for Device Module MVP
-- Database: MySQL / GreatDB / SQLite compatible

-- DeviceType table
CREATE TABLE IF NOT EXISTS device_types (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    vendor VARCHAR(255) NOT NULL,
    model VARCHAR(255) NOT NULL,
    description TEXT,
    capabilities JSON,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (name, vendor, model)
);

-- Device table
CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    device_type_id VARCHAR(64) NOT NULL,
    address TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'REGISTERED',
    last_online_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Indexes (separate statements for SQLite compatibility)
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
CREATE INDEX IF NOT EXISTS idx_devices_device_type_id ON devices(device_type_id);
CREATE INDEX IF NOT EXISTS idx_devices_name ON devices(name);

-- Note: Foreign key constraint for devices.device_type_id is enforced at application layer
-- SQLite doesn't support ALTER TABLE ADD CONSTRAINT
