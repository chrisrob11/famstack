-- +goose Up
-- Add recurring event fields to calendar_events table
ALTER TABLE calendar_events ADD COLUMN is_recurring BOOLEAN DEFAULT FALSE;
ALTER TABLE calendar_events ADD COLUMN recurrence_rules TEXT; -- JSON array of RRULE strings
ALTER TABLE calendar_events ADD COLUMN recurring_event_id TEXT; -- Parent recurring event ID
ALTER TABLE calendar_events ADD COLUMN is_recurring_instance BOOLEAN DEFAULT FALSE;

-- Create index for recurring event queries
CREATE INDEX idx_calendar_events_recurring ON calendar_events(recurring_event_id);
CREATE INDEX idx_calendar_events_is_recurring ON calendar_events(is_recurring);

-- +goose Down
-- Remove recurring event indexes
DROP INDEX IF EXISTS idx_calendar_events_recurring;
DROP INDEX IF EXISTS idx_calendar_events_is_recurring;

-- Remove recurring event columns
ALTER TABLE calendar_events DROP COLUMN is_recurring;
ALTER TABLE calendar_events DROP COLUMN recurrence_rules;
ALTER TABLE calendar_events DROP COLUMN recurring_event_id;
ALTER TABLE calendar_events DROP COLUMN is_recurring_instance;