-- Migration: 002_create_operation_tables.sql
-- Purpose: Create operations and executions tables for Device Operation MVP
-- Database: MySQL / GreatDB / SQLite compatible

-- Operations table
CREATE TABLE IF NOT EXISTS operations (
    id VARCHAR(64) PRIMARY KEY,
    device_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    parameters JSON,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_operations_device_id ON operations(device_id);
CREATE INDEX IF NOT EXISTS idx_operations_status ON operations(status);
CREATE INDEX IF NOT EXISTS idx_operations_created_at ON operations(created_at);

-- Executions table
CREATE TABLE IF NOT EXISTS executions (
    id VARCHAR(64) PRIMARY KEY,
    operation_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 0,
    request JSON,
    result JSON,
    error TEXT,
    started_at TIMESTAMP NULL,
    finished_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_executions_operation_id ON executions(operation_id);
CREATE INDEX IF NOT EXISTS idx_executions_device_id ON executions(device_id);
CREATE INDEX IF NOT EXISTS idx_executions_status ON executions(status);