-- +goose Up
-- Add version column to jobs table for proper optimistic concurrency control
ALTER TABLE jobs ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- Create index for efficient version-based queries
CREATE INDEX idx_jobs_id_version ON jobs(id, version);

-- +goose Down
-- Remove version-based optimistic concurrency control
DROP INDEX idx_jobs_id_version;
ALTER TABLE jobs DROP COLUMN version;