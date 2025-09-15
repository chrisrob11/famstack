package jobsystem

import (
	"context"
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents a job to be processed
type Job struct {
	ID             string                 `json:"id" db:"id"`
	QueueName      string                 `json:"queue_name" db:"queue_name"`
	JobType        string                 `json:"job_type" db:"job_type"`
	Payload        map[string]interface{} `json:"payload" db:"payload"`
	Status         JobStatus              `json:"status" db:"status"`
	Priority       int                    `json:"priority" db:"priority"`
	MaxRetries     int                    `json:"max_retries" db:"max_retries"`
	RetryCount     int                    `json:"retry_count" db:"retry_count"`
	RunAt          time.Time              `json:"run_at" db:"run_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
	QueuedAt       *time.Time             `json:"queued_at" db:"queued_at"`
	StartedAt      *time.Time             `json:"started_at" db:"started_at"`
	CompletedAt    *time.Time             `json:"completed_at" db:"completed_at"`
	Error          *string                `json:"error" db:"error"`
	IdempotencyKey *string                `json:"idempotency_key" db:"idempotency_key"`
	Version        int64                  `json:"version" db:"version"`
}

// ScheduledJob represents a recurring job
type ScheduledJob struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	QueueName string                 `json:"queue_name" db:"queue_name"`
	JobType   string                 `json:"job_type" db:"job_type"`
	Payload   map[string]interface{} `json:"payload" db:"payload"`
	CronExpr  string                 `json:"cron_expr" db:"cron_expr"`
	Enabled   bool                   `json:"enabled" db:"enabled"`
	NextRunAt time.Time              `json:"next_run_at" db:"next_run_at"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	LastRunAt *time.Time             `json:"last_run_at" db:"last_run_at"`
}

// EnqueueRequest represents a request to enqueue a job
type EnqueueRequest struct {
	QueueName      string                 `json:"queue_name"`
	JobType        string                 `json:"job_type"`
	Payload        map[string]interface{} `json:"payload"`
	Priority       int                    `json:"priority"`
	MaxRetries     int                    `json:"max_retries"`
	RunAt          *time.Time             `json:"run_at"`
	RunIn          time.Duration          `json:"run_in"`
	IdempotencyKey *string                `json:"idempotency_key"` // Optional key for preventing duplicate jobs
}

// ScheduleRequest represents a request to schedule a recurring job
type ScheduleRequest struct {
	Name      string                 `json:"name"`
	QueueName string                 `json:"queue_name"`
	JobType   string                 `json:"job_type"`
	Payload   map[string]interface{} `json:"payload"`
	CronExpr  string                 `json:"cron_expr"`
	Enabled   bool                   `json:"enabled"`
}

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, job *Job) error

// JobMetric represents metrics for a job execution
type JobMetric struct {
	ID         string    `json:"id" db:"id"`
	QueueName  string    `json:"queue_name" db:"queue_name"`
	JobType    string    `json:"job_type" db:"job_type"`
	Status     JobStatus `json:"status" db:"status"`
	DurationMs *int64    `json:"duration_ms" db:"duration_ms"`
	RecordedAt time.Time `json:"recorded_at" db:"recorded_at"`
}

// REDMetrics represents Rate, Error, and Duration metrics
type REDMetrics struct {
	// Rate - jobs per second
	JobsPerSecond float64 `json:"jobs_per_second"`

	// Error - error rate (percentage)
	ErrorRate float64 `json:"error_rate"`

	// Duration - latency percentiles in milliseconds
	P50LatencyMs float64 `json:"p50_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	P99LatencyMs float64 `json:"p99_latency_ms"`

	// Additional useful metrics
	TotalJobs        int64   `json:"total_jobs"`
	FailedJobs       int64   `json:"failed_jobs"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
}

// Config represents configuration for the job system
type Config struct {
	// Database connection string
	DatabasePath string `json:"database_path"`

	// Worker configuration
	WorkerConcurrency  map[string]int `json:"worker_concurrency"` // queue_name -> concurrency
	DefaultConcurrency int            `json:"default_concurrency"`
	PollInterval       time.Duration  `json:"poll_interval"`

	// Retry configuration
	DefaultMaxRetries int           `json:"default_max_retries"`
	RetryBackoffBase  time.Duration `json:"retry_backoff_base"`
	RetryBackoffMax   time.Duration `json:"retry_backoff_max"`

	// Scheduler configuration
	SchedulerEnabled  bool          `json:"scheduler_enabled"`
	SchedulerInterval time.Duration `json:"scheduler_interval"`

	// Metrics configuration
	MetricsRetention time.Duration `json:"metrics_retention"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		WorkerConcurrency: map[string]int{
			"default":         5,
			"task_generation": 3,
		},
		DefaultConcurrency: 5,
		PollInterval:       5 * time.Second,
		DefaultMaxRetries:  3,
		RetryBackoffBase:   1 * time.Second,
		RetryBackoffMax:    5 * time.Minute,
		SchedulerEnabled:   true,
		SchedulerInterval:  1 * time.Minute,
		MetricsRetention:   24 * time.Hour,
	}
}
