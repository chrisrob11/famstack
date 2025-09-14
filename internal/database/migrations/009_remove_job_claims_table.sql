-- +goose Up
-- Remove the job_claims table since job system now handles concurrency at dequeue level
DROP TABLE IF EXISTS job_claims;

-- +goose Down
-- Recreate job_claims table (though this should not be needed)
CREATE TABLE job_claims (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    claim_key TEXT NOT NULL UNIQUE,
    job_id TEXT NOT NULL,
    job_type TEXT NOT NULL,
    claimed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

CREATE INDEX idx_job_claims_claim_key ON job_claims(claim_key);
CREATE INDEX idx_job_claims_expires_at ON job_claims(expires_at);
CREATE INDEX idx_job_claims_job_id ON job_claims(job_id);