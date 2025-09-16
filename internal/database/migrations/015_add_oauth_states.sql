-- +goose Up
-- Create OAuth states table for temporary state storage during OAuth flow
CREATE TABLE oauth_states (
    state TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES family_members(id) ON DELETE CASCADE
);

-- Create index for cleanup of expired states
CREATE INDEX idx_oauth_states_expires ON oauth_states(expires_at);

-- +goose Down
DROP TABLE IF EXISTS oauth_states;