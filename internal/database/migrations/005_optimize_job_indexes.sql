-- +goose Up
-- Add optimized index for worker polling queries
-- This index covers the exact query pattern used by workers:
-- WHERE queue_name = ? AND status = 'pending' AND run_at <= datetime('now') ORDER BY priority DESC, run_at ASC
CREATE INDEX IF NOT EXISTS idx_jobs_queue_status_run_at_priority ON jobs(queue_name, status, run_at, priority);

-- +goose Down
DROP INDEX IF EXISTS idx_jobs_queue_status_run_at_priority;