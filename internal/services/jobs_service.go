package services

import (
	"fmt"
	"strings"
	"time"

	"famstack/internal/database"
)

// JobsService handles job system database operations
type JobsService struct {
	db *database.Fascade
}

// NewJobsService creates a new jobs service
func NewJobsService(db *database.Fascade) *JobsService {
	return &JobsService{db: db}
}

// Job represents a job in the system
type Job struct {
	ID             string                 `json:"id"`
	QueueName      string                 `json:"queue_name"`
	JobType        string                 `json:"job_type"`
	Payload        map[string]interface{} `json:"payload"`
	Status         string                 `json:"status"`
	Priority       int                    `json:"priority"`
	MaxRetries     int                    `json:"max_retries"`
	RetryCount     int                    `json:"retry_count"`
	RunAt          time.Time              `json:"run_at"`
	StartedAt      *time.Time             `json:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at"`
	Error          string                 `json:"error"`
	IdempotencyKey *string                `json:"idempotency_key"`
	Version        int                    `json:"version"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ScheduledJob represents a scheduled job
type ScheduledJob struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	QueueName string    `json:"queue_name"`
	JobType   string    `json:"job_type"`
	Payload   string    `json:"payload"`
	CronExpr  string    `json:"cron_expr"`
	Enabled   bool      `json:"enabled"`
	NextRunAt time.Time `json:"next_run_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JobMetric represents job execution metrics
type JobMetric struct {
	ID         string    `json:"id"`
	QueueName  string    `json:"queue_name"`
	JobType    string    `json:"job_type"`
	Status     string    `json:"status"`
	DurationMs int64     `json:"duration_ms"`
	RecordedAt time.Time `json:"recorded_at"`
}

// EnqueueJob creates a new job
func (s *JobsService) EnqueueJob(queueName, jobType string, payload string, priority, maxRetries int, runAt time.Time, idempotencyKey *string) (string, error) {
	query := `
		INSERT INTO jobs (queue_name, job_type, payload, priority, max_retries, run_at, queued_at, idempotency_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	var jobID string
	err := s.db.QueryRow(query,
		queueName,
		jobType,
		payload,
		priority,
		maxRetries,
		runAt.Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		idempotencyKey,
	).Scan(&jobID)

	if err != nil {
		// Check if this is a duplicate idempotency key
		if strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(err.Error(), "idempotency_key") {
			// Job with this idempotency key already exists - find and return existing job ID
			if idempotencyKey != nil {
				existingQuery := `SELECT id FROM jobs WHERE idempotency_key = ?`
				var existingJobID string
				if existingErr := s.db.QueryRow(existingQuery, *idempotencyKey).Scan(&existingJobID); existingErr == nil {
					return existingJobID, nil // Return existing job ID instead of error
				}
			}
		}
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	return jobID, nil
}

// ScheduleJob creates a scheduled job
func (s *JobsService) ScheduleJob(name, queueName, jobType, payload, cronExpr string, enabled bool, nextRunAt time.Time) error {
	query := `
		INSERT OR REPLACE INTO scheduled_jobs (name, queue_name, job_type, payload, cron_expr, enabled, next_run_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		name,
		queueName,
		jobType,
		payload,
		cronExpr,
		enabled,
		nextRunAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	return nil
}

// GetPendingJobs retrieves pending jobs for a queue
func (s *JobsService) GetPendingJobs(queueName string, limit int) ([]Job, error) {
	query := `
		SELECT id, queue_name, job_type, payload, status, priority, max_retries, retry_count, run_at, idempotency_key, version
		FROM jobs
		WHERE queue_name = ? AND status = 'pending' AND run_at <= datetime('now')
		ORDER BY priority DESC, run_at ASC
		LIMIT ?
	`

	rows, err := s.db.Query(query, queueName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		job := Job{}
		var runAtStr string
		var payloadStr string

		if scanErr := rows.Scan(
			&job.ID, &job.QueueName, &job.JobType, &payloadStr,
			&job.Status, &job.Priority, &job.MaxRetries, &job.RetryCount, &runAtStr, &job.IdempotencyKey, &job.Version,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan job: %w", scanErr)
		}

		// Parse payload JSON - simplified for now
		job.Payload = map[string]interface{}{"raw": payloadStr}

		// Parse run_at time
		if job.RunAt, err = time.Parse("2006-01-02 15:04:05", runAtStr); err != nil {
			if job.RunAt, err = time.Parse(time.RFC3339, runAtStr); err != nil {
				return nil, fmt.Errorf("failed to parse run_at time: %w", err)
			}
		}

		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job rows: %w", err)
	}

	return jobs, nil
}

// ClaimJob attempts to claim a job for processing using optimistic locking
func (s *JobsService) ClaimJob(jobID string, expectedVersion int) (bool, error) {
	startedAt := time.Now()
	query := `
		UPDATE jobs
		SET status = 'running', started_at = ?, updated_at = ?, version = version + 1
		WHERE id = ? AND version = ? AND status = 'pending'
	`
	result, err := s.db.Exec(query,
		startedAt.Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		jobID,
		expectedVersion,
	)
	if err != nil {
		return false, fmt.Errorf("failed to claim job %s: %w", jobID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check rows affected for job %s: %w", jobID, err)
	}

	return rowsAffected > 0, nil
}

// MarkJobCompleted marks a job as completed
func (s *JobsService) MarkJobCompleted(jobID string) error {
	_, err := s.db.Exec(
		"UPDATE jobs SET status = 'completed', completed_at = datetime('now'), version = version + 1, updated_at = datetime('now') WHERE id = ?",
		jobID,
	)
	return err
}

// MarkJobFailed marks a job as failed
func (s *JobsService) MarkJobFailed(jobID, errorMsg string) error {
	_, err := s.db.Exec(
		"UPDATE jobs SET status = 'failed', error = ?, completed_at = datetime('now'), version = version + 1, updated_at = datetime('now') WHERE id = ?",
		errorMsg, jobID,
	)
	return err
}

// ScheduleJobRetry schedules a job for retry
func (s *JobsService) ScheduleJobRetry(jobID string, retryAt time.Time, errorMsg string) error {
	_, err := s.db.Exec(
		"UPDATE jobs SET status = 'pending', retry_count = retry_count + 1, run_at = ?, error = ?, version = version + 1, updated_at = datetime('now') WHERE id = ?",
		retryAt.Format("2006-01-02 15:04:05"), errorMsg, jobID,
	)
	return err
}

// ResetJobToPending resets a job back to pending status
func (s *JobsService) ResetJobToPending(jobID string) error {
	query := `UPDATE jobs SET status = 'pending', started_at = NULL, version = version + 1, updated_at = datetime('now') WHERE id = ?`
	_, err := s.db.Exec(query, jobID)
	return err
}

// RecordJobMetric records job execution metrics
func (s *JobsService) RecordJobMetric(queueName, jobType, status string, durationMs int64) error {
	_, err := s.db.Exec(`
		INSERT INTO job_metrics (queue_name, job_type, status, duration_ms, recorded_at)
		VALUES (?, ?, ?, ?, datetime('now'))
	`, queueName, jobType, status, durationMs)

	return err
}

// GetJobMetrics retrieves job metrics for analysis
func (s *JobsService) GetJobMetrics(queueName, jobType string, timeWindow time.Duration) (*JobMetricsResult, error) {
	cutoff := time.Now().Add(-timeWindow)

	query := `
		SELECT
			COUNT(*) as total_jobs,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_jobs,
			AVG(CASE WHEN duration_ms IS NOT NULL THEN duration_ms ELSE 0 END) as avg_latency,
			COUNT(*) / ? as jobs_per_second
		FROM job_metrics
		WHERE recorded_at >= ?
	`

	args := []interface{}{timeWindow.Seconds(), cutoff.Format("2006-01-02 15:04:05")}

	if queueName != "" {
		query += " AND queue_name = ?"
		args = append(args, queueName)
	}

	if jobType != "" {
		query += " AND job_type = ?"
		args = append(args, jobType)
	}

	var totalJobs, failedJobs int64
	var avgLatency, jobsPerSecond float64

	err := s.db.QueryRow(query, args...).Scan(&totalJobs, &failedJobs, &avgLatency, &jobsPerSecond)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic metrics: %w", err)
	}

	errorRate := float64(0)
	if totalJobs > 0 {
		errorRate = float64(failedJobs) / float64(totalJobs) * 100
	}

	return &JobMetricsResult{
		TotalJobs:        totalJobs,
		FailedJobs:       failedJobs,
		ErrorRate:        errorRate,
		AverageLatencyMs: avgLatency,
		JobsPerSecond:    jobsPerSecond,
	}, nil
}

// JobMetricsResult represents aggregated job metrics
type JobMetricsResult struct {
	TotalJobs        int64   `json:"total_jobs"`
	FailedJobs       int64   `json:"failed_jobs"`
	ErrorRate        float64 `json:"error_rate"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	JobsPerSecond    float64 `json:"jobs_per_second"`
}
