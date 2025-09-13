-- +goose Up
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
    recurring BOOLEAN DEFAULT FALSE,
    recurrence_rule TEXT, -- RRULE format for recurring events
    status TEXT DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'tentative', 'cancelled')),
    visibility TEXT DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    raw_data TEXT, -- JSON blob of original event data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
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
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(event_id, user_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_calendar_events_family_start ON calendar_events(family_id, start_time);
CREATE INDEX IF NOT EXISTS idx_calendar_events_external ON calendar_events(external_source, external_id);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_events_family_start ON unified_calendar_events(family_id, start_time);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_events_date_range ON unified_calendar_events(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_event_attendees_event ON unified_calendar_event_attendees(event_id);
CREATE INDEX IF NOT EXISTS idx_unified_calendar_event_attendees_user ON unified_calendar_event_attendees(user_id);

-- Insert sample calendar events for testing
INSERT OR IGNORE INTO calendar_events (id, external_id, external_source, family_id, title, description, start_time, end_time, location) VALUES
    ('cal1', 'ext_soccer_1', 'manual', 'fam1', 'Soccer Practice', 'Weekly soccer practice for Bobby', '2025-09-12 16:00:00', '2025-09-12 17:30:00', 'City Sports Complex'),
    ('cal2', 'ext_book_1', 'manual', 'fam1', 'Book Club', 'Monthly book club meeting', '2025-09-12 18:00:00', '2025-09-12 19:30:00', 'Community Center'),
    ('cal3', 'ext_grocery_1', 'manual', 'fam1', 'Grocery Shopping', 'Weekly grocery run', '2025-09-12 17:00:00', '2025-09-12 17:45:00', 'Whole Foods'),
    ('cal4', 'ext_piano_1', 'manual', 'fam1', 'Piano Lesson', 'Bobby piano lesson', '2025-09-12 17:30:00', '2025-09-12 18:15:00', 'Music Academy'),
    ('cal5', 'ext_dinner_1', 'manual', 'fam1', 'Family Dinner', 'Family dinner at home', '2025-09-12 19:00:00', '2025-09-12 20:00:00', 'Home');

-- Create unified events from the raw events
INSERT OR IGNORE INTO unified_calendar_events (id, family_id, title, description, start_time, end_time, location, event_type, color, created_by) VALUES
    ('uni1', 'fam1', 'Soccer Practice', 'Weekly soccer practice for Bobby', '2025-09-12 16:00:00', '2025-09-12 17:30:00', 'City Sports Complex', 'event', '#10b981', 'user1'),
    ('uni2', 'fam1', 'Book Club', 'Monthly book club meeting', '2025-09-12 18:00:00', '2025-09-12 19:30:00', 'Community Center', 'event', '#8b5cf6', 'user2'),
    ('uni3', 'fam1', 'Grocery Shopping', 'Weekly grocery run', '2025-09-12 17:00:00', '2025-09-12 17:45:00', 'Whole Foods', 'task', '#f59e0b', 'user1'),
    ('uni4', 'fam1', 'Piano Lesson', 'Bobby piano lesson', '2025-09-12 17:30:00', '2025-09-12 18:15:00', 'Music Academy', 'event', '#10b981', 'user1'),
    ('uni5', 'fam1', 'Family Dinner', 'Family dinner at home', '2025-09-12 19:00:00', '2025-09-12 20:00:00', 'Home', 'event', '#3b82f6', 'user1');

-- Link unified events to raw calendar events
INSERT OR IGNORE INTO unified_calendar_to_calendar_event_rel (unified_event_id, calendar_event_id, relationship_type) VALUES
    ('uni1', 'cal1', 'source'),
    ('uni2', 'cal2', 'source'),
    ('uni3', 'cal3', 'source'),
    ('uni4', 'cal4', 'source'),
    ('uni5', 'cal5', 'source');

-- Add attendee relationships for better querying and user management
INSERT OR IGNORE INTO unified_calendar_event_attendees (event_id, user_id, response_status) VALUES
    ('uni1', 'user3', 'accepted'), -- Bobby for soccer practice
    ('uni2', 'user2', 'accepted'), -- Mom for book club
    ('uni3', 'user2', 'accepted'), -- Mom for grocery shopping
    ('uni4', 'user3', 'accepted'), -- Bobby for piano lesson
    ('uni5', 'user1', 'accepted'), -- Dad for family dinner
    ('uni5', 'user2', 'accepted'), -- Mom for family dinner
    ('uni5', 'user3', 'accepted'); -- Bobby for family dinner

-- +goose Down
DROP TABLE IF EXISTS unified_calendar_to_calendar_event_rel;
DROP TABLE IF EXISTS unified_calendar_events;
DROP TABLE IF EXISTS calendar_events;