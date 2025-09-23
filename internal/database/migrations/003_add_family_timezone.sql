-- +goose Up
-- Add timezone column to families table for proper time handling
-- Validation will be handled in Go application layer using time.LoadLocation()
ALTER TABLE families ADD COLUMN timezone TEXT DEFAULT 'UTC';

-- +goose Down
-- Remove timezone column from families table
ALTER TABLE families DROP COLUMN timezone;