-- +goose Up
-- Create comprehensive integrations system

-- Master integrations table
CREATE TABLE integrations (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    integration_type TEXT NOT NULL, -- 'calendar', 'storage', 'communication', 'automation', 'smart_home', 'finance'
    provider TEXT NOT NULL,         -- 'google', 'microsoft', 'apple', 'slack', 'dropbox', etc
    auth_method TEXT NOT NULL,      -- 'oauth2', 'api_key', 'basic_auth', 'webhook'
    status TEXT NOT NULL DEFAULT 'pending', -- 'connected', 'disconnected', 'error', 'pending'
    display_name TEXT NOT NULL,     -- "John's Google Calendar", "Family Dropbox"
    description TEXT,               -- Optional description
    settings TEXT,                  -- JSON configuration specific to integration
    last_sync_at DATETIME,
    last_error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES family_members(id) ON DELETE CASCADE
);

-- OAuth credentials (encrypted, separate from main table)
CREATE TABLE integration_oauth_credentials (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    integration_id TEXT NOT NULL,
    access_token TEXT NOT NULL,     -- encrypted
    refresh_token TEXT,             -- encrypted
    token_type TEXT DEFAULT 'Bearer',
    expires_at DATETIME,
    scope TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE,
    UNIQUE(integration_id) -- One OAuth credential per integration
);

-- API keys and other auth methods (encrypted, separate from main table)
CREATE TABLE integration_api_credentials (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    integration_id TEXT NOT NULL,
    credential_type TEXT NOT NULL,  -- 'api_key', 'webhook_secret', 'basic_auth', 'token'
    credential_name TEXT,           -- 'api_key', 'username', 'password', etc
    credential_value TEXT NOT NULL, -- encrypted
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE
);

-- Integration sync history
CREATE TABLE integration_sync_history (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    integration_id TEXT NOT NULL,
    sync_type TEXT NOT NULL,        -- 'manual', 'scheduled', 'webhook'
    status TEXT NOT NULL,           -- 'success', 'error', 'partial'
    items_synced INTEGER DEFAULT 0,
    error_message TEXT,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE
);

-- Create indexes for efficient queries
CREATE INDEX idx_integrations_family ON integrations(family_id);
CREATE INDEX idx_integrations_user ON integrations(user_id);
CREATE INDEX idx_integrations_type ON integrations(integration_type);
CREATE INDEX idx_integrations_provider ON integrations(provider);
CREATE INDEX idx_integrations_status ON integrations(status);
CREATE INDEX idx_integrations_auth_method ON integrations(auth_method);

CREATE INDEX idx_oauth_creds_integration ON integration_oauth_credentials(integration_id);
CREATE INDEX idx_api_creds_integration ON integration_api_credentials(integration_id);
CREATE INDEX idx_api_creds_type ON integration_api_credentials(credential_type);

CREATE INDEX idx_sync_history_integration ON integration_sync_history(integration_id);
CREATE INDEX idx_sync_history_status ON integration_sync_history(status);
CREATE INDEX idx_sync_history_started ON integration_sync_history(started_at);

-- +goose Down
-- Remove indexes
DROP INDEX IF EXISTS idx_sync_history_started;
DROP INDEX IF EXISTS idx_sync_history_status;
DROP INDEX IF EXISTS idx_sync_history_integration;
DROP INDEX IF EXISTS idx_api_creds_type;
DROP INDEX IF EXISTS idx_api_creds_integration;
DROP INDEX IF EXISTS idx_oauth_creds_integration;
DROP INDEX IF EXISTS idx_integrations_auth_method;
DROP INDEX IF EXISTS idx_integrations_status;
DROP INDEX IF EXISTS idx_integrations_provider;
DROP INDEX IF EXISTS idx_integrations_type;
DROP INDEX IF EXISTS idx_integrations_user;
DROP INDEX IF EXISTS idx_integrations_family;

-- Remove tables
DROP TABLE IF EXISTS integration_sync_history;
DROP TABLE IF EXISTS integration_api_credentials;
DROP TABLE IF EXISTS integration_oauth_credentials;
DROP TABLE IF EXISTS integrations;