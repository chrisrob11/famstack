-- +goose Up
-- Add idempotency_key column to jobs table for built-in duplicate prevention
ALTER TABLE jobs ADD COLUMN idempotency_key TEXT;

-- Create unique index to prevent duplicate jobs with same idempotency key
CREATE UNIQUE INDEX idx_jobs_idempotency_key ON jobs(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- +goose Down
-- Remove idempotency key support
DROP INDEX idx_jobs_idempotency_key;
ALTER TABLE jobs DROP COLUMN idempotency_key;