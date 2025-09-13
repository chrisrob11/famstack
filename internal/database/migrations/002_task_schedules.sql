-- +goose Up
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
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL
);

-- Create task_queue table for reliable task generation
CREATE TABLE IF NOT EXISTS task_queue (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    queue_type TEXT NOT NULL, -- 'generate_scheduled_tasks'
    payload TEXT NOT NULL, -- JSON with schedule_id, target_date, etc.
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    scheduled_for DATETIME NOT NULL, -- when to process this queue item
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME,
    error_message TEXT
);

-- Add schedule_id to existing tasks table to link generated tasks back to their schedule
ALTER TABLE tasks ADD COLUMN schedule_id TEXT REFERENCES task_schedules(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_task_schedules_family_active ON task_schedules(family_id, active);
CREATE INDEX IF NOT EXISTS idx_task_queue_status_scheduled ON task_queue(status, scheduled_for);
CREATE INDEX IF NOT EXISTS idx_tasks_schedule_id ON tasks(schedule_id);

-- +goose Down
DROP INDEX IF EXISTS idx_tasks_schedule_id;
DROP INDEX IF EXISTS idx_task_queue_status_scheduled;
DROP INDEX IF NOT EXISTS idx_task_schedules_family_active;
ALTER TABLE tasks DROP COLUMN schedule_id;
DROP TABLE IF EXISTS task_queue;
DROP TABLE IF EXISTS task_schedules;