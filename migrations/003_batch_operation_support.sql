-- Migration: 003_batch_operation_support.sql
-- Purpose: Add batch operation counters to operations table
-- Database: MySQL / GreatDB / SQLite compatible

-- Add batch counters to operations table
ALTER TABLE operations ADD COLUMN total_count INT NOT NULL DEFAULT 0;
ALTER TABLE operations ADD COLUMN success_count INT NOT NULL DEFAULT 0;
ALTER TABLE operations ADD COLUMN failed_count INT NOT NULL DEFAULT 0;
ALTER TABLE operations ADD COLUMN skipped_count INT NOT NULL DEFAULT 0;
ALTER TABLE operations ADD COLUMN finished_at TIMESTAMP NULL;

-- Update existing single-device operations to have correct counters
UPDATE operations SET total_count = 1 WHERE total_count = 0;
UPDATE operations SET success_count = 1 WHERE status = 'COMPLETED' AND success_count = 0;
UPDATE operations SET failed_count = 1 WHERE status = 'FAILED' AND failed_count = 0;

-- Index for batch operation queries
CREATE INDEX IF NOT EXISTS idx_operations_finished_at ON operations(finished_at);