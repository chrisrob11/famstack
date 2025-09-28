-- +goose Up
-- Add updated_at column to tasks table (SQLite doesn't support non-constant defaults in ALTER TABLE)
ALTER TABLE tasks ADD COLUMN updated_at DATETIME;

-- Update existing tasks to have the same updated_at as created_at initially
UPDATE tasks SET updated_at = created_at;

-- +goose Down
-- Remove updated_at column from tasks table
ALTER TABLE tasks DROP COLUMN updated_at;