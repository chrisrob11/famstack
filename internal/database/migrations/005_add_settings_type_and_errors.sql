-- +goose Up
-- Migration 005: Add settings_type to integrations and background errors table

-- Add settings_type to integrations table for Go struct unmarshalling
ALTER TABLE integrations ADD COLUMN settings_type TEXT;

-- Add enabled column to integrations for admin control
ALTER TABLE integrations ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT true;

-- Add sync tracking columns
ALTER TABLE integrations ADD COLUMN last_sync_token TEXT; -- For incremental sync

-- Background errors table for tracking integration and job errors
CREATE TABLE background_errors (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    error_type TEXT NOT NULL, -- 'job_error', 'integration_error', 'sync_error', etc.
    integration_id TEXT, -- NULL if not integration-specific
    job_id TEXT, -- NULL if not job-specific
    error_message TEXT NOT NULL,
    error_details TEXT, -- JSON with additional context
    occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,

    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE
);

-- Indexes for efficient error querying and cleanup
CREATE INDEX idx_background_errors_occurred_at ON background_errors(occurred_at);
CREATE INDEX idx_background_errors_integration_id ON background_errors(integration_id);
CREATE INDEX idx_background_errors_error_type ON background_errors(error_type);
CREATE INDEX idx_background_errors_resolved ON background_errors(resolved_at);
CREATE INDEX idx_background_errors_job_id ON background_errors(job_id);

-- Add indexes for new integration fields
CREATE INDEX idx_integrations_enabled ON integrations(enabled);
CREATE INDEX idx_integrations_settings_type ON integrations(settings_type);

-- +goose Down
-- Remove background errors table
DROP TABLE IF EXISTS background_errors;

-- Remove new integration columns
ALTER TABLE integrations DROP COLUMN last_sync_token;
ALTER TABLE integrations DROP COLUMN enabled;
ALTER TABLE integrations DROP COLUMN settings_type;

-- Remove indexes
DROP INDEX IF EXISTS idx_integrations_enabled;
DROP INDEX IF EXISTS idx_integrations_settings_type;