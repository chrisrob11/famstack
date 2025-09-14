-- +goose Up
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
    error TEXT
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

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_jobs_status_run_at ON jobs(status, run_at);
CREATE INDEX IF NOT EXISTS idx_jobs_queue_status ON jobs(queue_name, status);
CREATE INDEX IF NOT EXISTS idx_jobs_job_type_status ON jobs(job_type, status);
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
-- Optimal index for worker polling queries
CREATE INDEX IF NOT EXISTS idx_jobs_queue_status_run_at_priority ON jobs(queue_name, status, run_at, priority);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_enabled_next_run ON scheduled_jobs(enabled, next_run_at);
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_name ON scheduled_jobs(name);
CREATE INDEX IF NOT EXISTS idx_job_metrics_queue_type ON job_metrics(queue_name, job_type, recorded_at);

-- Note: Triggers for updating updated_at timestamps can be added later if needed

-- +goose Down
DROP INDEX IF EXISTS idx_job_metrics_queue_type;
DROP INDEX IF EXISTS idx_scheduled_jobs_name;
DROP INDEX IF EXISTS idx_scheduled_jobs_enabled_next_run;
DROP INDEX IF EXISTS idx_jobs_queue_status_run_at_priority;
DROP INDEX IF EXISTS idx_jobs_created_at;
DROP INDEX IF EXISTS idx_jobs_job_type_status;
DROP INDEX IF EXISTS idx_jobs_queue_status;
DROP INDEX IF EXISTS idx_jobs_status_run_at;
DROP TABLE IF EXISTS job_metrics;
DROP TABLE IF EXISTS scheduled_jobs;
DROP TABLE IF EXISTS jobs;