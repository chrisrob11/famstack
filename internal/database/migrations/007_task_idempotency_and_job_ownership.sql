-- +goose Up
-- Comprehensive task generation idempotency and job ownership system

-- 1. Add unique constraint to prevent duplicate tasks for the same schedule on the same date
-- Using expression index since SQLite doesn't support adding computed columns to existing tables
CREATE UNIQUE INDEX idx_tasks_schedule_target_date 
ON tasks(
    schedule_id, 
    CASE 
        WHEN due_date IS NOT NULL THEN DATE(due_date)
        ELSE DATE(created_at)
    END
) 
WHERE schedule_id IS NOT NULL;

-- 2. Add job ownership claims table for optimistic concurrency control
-- This prevents race conditions between concurrent job workers
CREATE TABLE job_claims (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    claim_key TEXT NOT NULL UNIQUE, -- e.g., "schedule:123:date:2025-09-14"
    job_id TEXT NOT NULL,
    job_type TEXT NOT NULL,
    claimed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL, -- Auto-expire claims after some time
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

-- Index for fast lookups and cleanup
CREATE INDEX idx_job_claims_claim_key ON job_claims(claim_key);
CREATE INDEX idx_job_claims_expires_at ON job_claims(expires_at);
CREATE INDEX idx_job_claims_job_id ON job_claims(job_id);

-- +goose Down
-- Remove task idempotency and job ownership features
DROP INDEX idx_tasks_schedule_target_date;
DROP TABLE job_claims;