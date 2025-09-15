-- +goose Up
-- Add last_generated_date column to task_schedules table
ALTER TABLE task_schedules ADD COLUMN last_generated_date DATETIME;

-- +goose Down
ALTER TABLE task_schedules DROP COLUMN last_generated_date;