-- +goose Up
-- Create families table
CREATE TABLE IF NOT EXISTS families (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create family_members table (combines users and family member info)
CREATE TABLE IF NOT EXISTS family_members (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    member_type TEXT NOT NULL DEFAULT 'child', -- 'adult', 'child', 'pet'
    avatar_url TEXT,

    -- Auth fields (optional - only for members who can login)
    email TEXT UNIQUE,
    password_hash TEXT,
    role TEXT CHECK (role IN ('shared', 'user', 'admin')),
    email_verified BOOLEAN DEFAULT FALSE,
    last_login_at DATETIME,

    -- Display and sorting
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,

    -- Metadata
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE
);

-- Create indexes for family_members
CREATE INDEX idx_family_members_family_id ON family_members(family_id);
CREATE UNIQUE INDEX idx_family_members_email ON family_members(email) WHERE email IS NOT NULL;

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    assigned_to TEXT,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    task_type TEXT NOT NULL CHECK (task_type IN ('todo', 'chore', 'appointment')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed')),
    priority INTEGER DEFAULT 0,
    due_date DATETIME,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    schedule_id TEXT REFERENCES task_schedules(id) ON DELETE SET NULL,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES family_members(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_tasks_schedule_target_date 
ON tasks(
    schedule_id, 
    CASE 
        WHEN due_date IS NOT NULL THEN DATE(due_date)
        ELSE DATE(created_at)
    END
) 
WHERE schedule_id IS NOT NULL;

-- Create task_schedules table
CREATE TABLE IF NOT EXISTS task_schedules (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    created_by TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    task_type TEXT NOT NULL CHECK (task_type IN ('todo', 'chore', 'appointment')),
    assigned_to TEXT,
    days_of_week TEXT NOT NULL, -- JSON array: ["tuesday", "thursday"] 
    time_of_day TEXT, -- HH:MM format, optional specific time
    priority INTEGER DEFAULT 0,
    points INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT true,
    last_generated_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES family_members(id) ON DELETE SET NULL
);

-- Create calendar_events table (raw events from external systems)
CREATE TABLE IF NOT EXISTS calendar_events (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    external_id TEXT UNIQUE, -- ID from external system (Google Calendar, Outlook, etc.)
    external_source TEXT NOT NULL, -- 'google', 'outlook', 'apple', 'manual', etc.
    family_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    location TEXT DEFAULT '',
    all_day BOOLEAN DEFAULT FALSE,
    status TEXT DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'tentative', 'cancelled')),
    visibility TEXT DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    raw_data TEXT, -- JSON blob of original event data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_rules TEXT, -- JSON array of RRULE strings
    recurring_event_id TEXT, -- Parent recurring event ID
    is_recurring_instance BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE
);

-- Create unified_calendar_events table (processed events for UI)
CREATE TABLE IF NOT EXISTS unified_calendar_events (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    location TEXT DEFAULT '',
    all_day BOOLEAN DEFAULT FALSE,
    event_type TEXT DEFAULT 'event' CHECK (event_type IN ('event', 'task', 'appointment', 'reminder')),
    color TEXT DEFAULT '#3b82f6', -- Hex color for UI display
    created_by TEXT,
    priority INTEGER DEFAULT 0,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'completed')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE SET NULL
);

-- Create relationship table linking unified events to raw calendar events
CREATE TABLE IF NOT EXISTS unified_calendar_to_calendar_event_rel (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    unified_event_id TEXT NOT NULL,
    calendar_event_id TEXT NOT NULL,
    relationship_type TEXT DEFAULT 'source' CHECK (relationship_type IN ('source', 'merged', 'duplicate')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (unified_event_id) REFERENCES unified_calendar_events(id) ON DELETE CASCADE,
    FOREIGN KEY (calendar_event_id) REFERENCES calendar_events(id) ON DELETE CASCADE,
    UNIQUE(unified_event_id, calendar_event_id)
);

-- Create attendee relationship table for better querying 
CREATE TABLE IF NOT EXISTS unified_calendar_event_attendees (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    response_status TEXT DEFAULT 'needsAction' CHECK (response_status IN ('needsAction', 'accepted', 'declined', 'tentative')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES unified_calendar_events(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES family_members(id) ON DELETE CASCADE,
    UNIQUE(event_id, user_id)
);

-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    queue_name TEXT NOT NULL DEFAULT 'default',
    job_type TEXT NOT NULL,
    payload TEXT NOT NULL, -- JSON payload
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    priority INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    retry_count INTEGER NOT NULL DEFAULT 0,
    run_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    queued_at DATETIME,
    started_at DATETIME,
    completed_at DATETIME,
    error TEXT,
    idempotency_key TEXT,
    version INTEGER NOT NULL DEFAULT 1
);

-- Create scheduled_jobs table
CREATE TABLE IF NOT EXISTS scheduled_jobs (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name TEXT NOT NULL UNIQUE,
    queue_name TEXT NOT NULL DEFAULT 'default',
    job_type TEXT NOT NULL,
    payload TEXT NOT NULL, -- JSON payload
    cron_expr TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    next_run_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_run_at DATETIME
);

-- Create job_metrics table for RED metrics
CREATE TABLE IF NOT EXISTS job_metrics (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    queue_name TEXT NOT NULL,
    job_type TEXT NOT NULL,
    status TEXT NOT NULL,
    duration_ms INTEGER, -- Duration in milliseconds
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create OAuth tokens table for storing third-party integrations
CREATE TABLE oauth_tokens (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    provider TEXT NOT NULL, -- 'google', 'microsoft', 'apple'
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type TEXT DEFAULT 'Bearer',
    expires_at DATETIME NOT NULL,
    scope TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE,

    -- Ensure one token per user per provider
    UNIQUE(created_by, provider)
);

-- Create calendar sync settings table
CREATE TABLE calendar_sync_settings (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    sync_frequency_minutes INTEGER DEFAULT 15, -- How often to sync (minutes)
    sync_range_days INTEGER DEFAULT 30, -- How many days ahead to sync
    last_sync_at DATETIME,
    sync_status TEXT DEFAULT 'pending', -- 'pending', 'syncing', 'success', 'error'
    sync_error TEXT,
    events_synced INTEGER DEFAULT 0,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE,

    -- One setting per user
    UNIQUE(created_by)
);

-- Create OAuth states table for temporary state storage during OAuth flow
CREATE TABLE oauth_states (
    state TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE
);

-- Master integrations table
CREATE TABLE integrations (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    family_id TEXT NOT NULL,
    integration_type TEXT NOT NULL, -- 'calendar', 'storage', 'communication', 'automation', 'smart_home', 'finance'
    provider TEXT NOT NULL,         -- 'google', 'microsoft', 'apple', 'slack', 'dropbox', etc
    auth_method TEXT NOT NULL,      -- 'oauth2', 'api_key', 'basic_auth', 'webhook'
    status TEXT NOT NULL DEFAULT 'pending', -- 'connected', 'disconnected', 'error', 'pending'
    display_name TEXT NOT NULL,     -- "John's Google Calendar", "Family Dropbox"
    description TEXT,               -- Optional description
    settings TEXT,                  -- JSON configuration specific to integration
    last_sync_at DATETIME,
    last_error TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES family_members(id) ON DELETE CASCADE
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


-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_task_schedules_family_active ON task_schedules(family_id, active);
CREATE INDEX IF NOT EXISTS idx_tasks_schedule_id ON tasks(schedule_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_family_start ON calendar_events(family_id, start_time);
CREATE INDEX IF NOT EXISTS idx_calendar_events_external ON calendar_events(external_source, external_id);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_events_family_start ON unified_calendar_events(family_id, start_time);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_events_date_range ON unified_calendar_events(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_event_attendees_event ON unified_calendar_event_attendees(event_id);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_event_attendees_user ON unified_calendar_event_attendees(user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status_run_at ON jobs(status, run_at);
CREATE INDEX IF NOT EXISTS idx_jobs_queue_status ON jobs(queue_name, status);
CREATE INDEX IF NOT EXISTS idx_jobs_job_type_status ON jobs(job_type, status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_queue_status_run_at_priority ON jobs(queue_name, status, run_at, priority);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_enabled_next_run ON scheduled_jobs(enabled, next_run_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_name ON scheduled_jobs(name);
CREATE INDEX IF NOT EXISTS idx_job_metrics_queue_type ON job_metrics(queue_name, job_type, recorded_at);
CREATE UNIQUE INDEX idx_jobs_idempotency_key ON jobs(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_jobs_id_version ON jobs(id, version);
CREATE INDEX idx_oauth_tokens_user_provider ON oauth_tokens(created_by, provider);
CREATE INDEX idx_oauth_tokens_family ON oauth_tokens(family_id);
CREATE INDEX idx_calendar_sync_settings_user ON calendar_sync_settings(created_by);
CREATE INDEX idx_calendar_sync_settings_family ON calendar_sync_settings(family_id);
CREATE INDEX idx_oauth_states_expires ON oauth_states(expires_at);
CREATE INDEX idx_calendar_events_recurring ON calendar_events(recurring_event_id);
CREATE INDEX idx_calendar_events_is_recurring ON calendar_events(is_recurring);
CREATE INDEX idx_integrations_family ON integrations(family_id);
CREATE INDEX idx_integrations_user ON integrations(created_by);
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
DROP TABLE IF EXISTS families;
DROP TABLE IF EXISTS family_members;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS task_schedules;
DROP TABLE IF EXISTS calendar_events;
DROP TABLE IF EXISTS unified_calendar_events;
DROP TABLE IF EXISTS unified_calendar_to_calendar_event_rel;
DROP TABLE IF EXISTS unified_calendar_event_attendees;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS scheduled_jobs;
DROP TABLE IF EXISTS job_metrics;
DROP TABLE IF EXISTS oauth_tokens;
DROP TABLE IF EXISTS calendar_sync_settings;
DROP TABLE IF EXISTS oauth_states;
DROP TABLE IF EXISTS integrations;
DROP TABLE IF EXISTS integration_oauth_credentials;
DROP TABLE IF EXISTS integration_api_credentials;
DROP TABLE IF EXISTS integration_sync_history;
