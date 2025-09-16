-- +goose Up
-- Create OAuth tokens table for storing third-party integrations
CREATE TABLE oauth_tokens (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL, -- 'google', 'microsoft', 'apple'
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type TEXT DEFAULT 'Bearer',
    expires_at DATETIME NOT NULL,
    scope TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES family_members(id) ON DELETE CASCADE,

    -- Ensure one token per user per provider
    UNIQUE(user_id, provider)
);

-- Create index for efficient lookups
CREATE INDEX idx_oauth_tokens_user_provider ON oauth_tokens(user_id, provider);
CREATE INDEX idx_oauth_tokens_family ON oauth_tokens(family_id);

-- Create calendar sync settings table
CREATE TABLE calendar_sync_settings (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    sync_frequency_minutes INTEGER DEFAULT 15, -- How often to sync (minutes)
    sync_range_days INTEGER DEFAULT 30, -- How many days ahead to sync
    last_sync_at DATETIME,
    sync_status TEXT DEFAULT 'pending', -- 'pending', 'syncing', 'success', 'error'
    sync_error TEXT,
    events_synced INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES family_members(id) ON DELETE CASCADE,

    -- One setting per user
    UNIQUE(user_id)
);

-- Create index for sync settings
CREATE INDEX idx_calendar_sync_settings_user ON calendar_sync_settings(user_id);
CREATE INDEX idx_calendar_sync_settings_family ON calendar_sync_settings(family_id);

-- +goose Down
DROP TABLE IF EXISTS calendar_sync_settings;
DROP TABLE IF EXISTS oauth_tokens;