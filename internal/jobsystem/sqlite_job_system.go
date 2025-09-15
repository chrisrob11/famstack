package jobsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"famstack/internal/database"
)

type SQLiteJobSystem struct {
	db             *database.DB
	config         *Config
	handlers       map[string]JobHandler
	workers        map[string]*workerPool
	scheduler      *scheduler
	metricsCleanup *time.Ticker
	shutdownCh     chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	running        bool
}

type workerPool struct {
	queueName   string
	concurrency int
	workers     []*worker
	jobCh       chan *Job
	stopCh      chan struct{}
}

type worker struct {
	id     int
	pool   *workerPool
	jobSys *SQLiteJobSystem
	stopCh chan struct{}
	doneCh chan struct{}
}

type scheduler struct {
	jobSys *SQLiteJobSystem
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewSQLiteJobSystem(config *Config, db *database.DB) *SQLiteJobSystem {
	if config == nil {
		config = DefaultConfig()
	}

	return &SQLiteJobSystem{
		db:         db,
		config:     config,
		handlers:   make(map[string]JobHandler),
		workers:    make(map[string]*workerPool),
		shutdownCh: make(chan struct{}),
	}
}

func (js *SQLiteJobSystem) Enqueue(req *EnqueueRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("enqueue request cannot be nil")
	}

	if req.QueueName == "" {
		req.QueueName = "default"
	}

	if req.MaxRetries == 0 {
		req.MaxRetries = js.config.DefaultMaxRetries
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	runAt := time.Now()
	if req.RunAt != nil {
		runAt = *req.RunAt
	} else if req.RunIn > 0 {
		runAt = time.Now().Add(req.RunIn)
	}

	query := `
		INSERT INTO jobs (queue_name, job_type, payload, priority, max_retries, run_at, queued_at, idempotency_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	var jobID string
	err = js.db.QueryRow(query,
		req.QueueName,
		req.JobType,
		string(payloadBytes),
		req.Priority,
		req.MaxRetries,
		runAt.Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		req.IdempotencyKey,
	).Scan(&jobID)

	if err != nil {
		// Check if this is a duplicate idempotency key
		if strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(err.Error(), "idempotency_key") {
			// Job with this idempotency key already exists - find and return existing job ID
			if req.IdempotencyKey != nil {
				existingQuery := `SELECT id FROM jobs WHERE idempotency_key = ?`
				var existingJobID string
				if existingErr := js.db.QueryRow(existingQuery, *req.IdempotencyKey).Scan(&existingJobID); existingErr == nil {
					return existingJobID, nil // Return existing job ID instead of error
				}
			}
		}
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	return jobID, nil
}

func (js *SQLiteJobSystem) Schedule(req *ScheduleRequest) error {
	if req == nil {
		return fmt.Errorf("schedule request cannot be nil")
	}

	if req.QueueName == "" {
		req.QueueName = "default"
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	nextRunAt, err := js.calculateNextRun(req.CronExpr)
	if err != nil {
		return fmt.Errorf("failed to calculate next run time: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO scheduled_jobs (name, queue_name, job_type, payload, cron_expr, enabled, next_run_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = js.db.Exec(query,
		req.Name,
		req.QueueName,
		req.JobType,
		string(payloadBytes),
		req.CronExpr,
		req.Enabled,
		nextRunAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return fmt.Errorf("failed to schedule job: %w", err)
	}

	return nil
}

func (js *SQLiteJobSystem) Register(jobType string, handler JobHandler) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.handlers[jobType] = handler
}

func (js *SQLiteJobSystem) Start(ctx context.Context) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if js.running {
		return fmt.Errorf("job system is already running")
	}

	log.Println("Starting SQLite job system...")

	for queueName, concurrency := range js.config.WorkerConcurrency {
		js.startWorkerPool(queueName, concurrency)
	}

	if js.config.SchedulerEnabled {
		js.startScheduler()
	}

	js.startMetricsCleanup()

	js.running = true

	go func() {
		<-ctx.Done()
		js.Stop()
	}()

	log.Println("SQLite job system started successfully")
	return nil
}

func (js *SQLiteJobSystem) Stop() {
	js.mu.Lock()
	defer js.mu.Unlock()

	if !js.running {
		return
	}

	log.Println("Stopping SQLite job system...")

	close(js.shutdownCh)

	for _, pool := range js.workers {
		js.stopWorkerPool(pool)
	}

	if js.scheduler != nil {
		js.stopScheduler()
	}

	if js.metricsCleanup != nil {
		js.metricsCleanup.Stop()
	}

	js.wg.Wait()
	js.running = false

	log.Println("SQLite job system stopped")
}

func (js *SQLiteJobSystem) GetMetrics(queueName, jobType string) (*REDMetrics, error) {
	timeWindow := 1 * time.Hour
	cutoff := time.Now().Add(-timeWindow)

	query := `
		SELECT 
			COUNT(*) as total_jobs,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_jobs,
			AVG(CASE WHEN duration_ms IS NOT NULL THEN duration_ms ELSE 0 END) as avg_latency,
			COUNT(*) / 3600.0 as jobs_per_second
		FROM job_metrics
		WHERE recorded_at >= ?
	`

	args := []interface{}{cutoff.Format("2006-01-02 15:04:05")}

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

	err := js.db.QueryRow(query, args...).Scan(&totalJobs, &failedJobs, &avgLatency, &jobsPerSecond)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic metrics: %w", err)
	}

	errorRate := float64(0)
	if totalJobs > 0 {
		errorRate = float64(failedJobs) / float64(totalJobs) * 100
	}

	percentileQuery := `
		SELECT duration_ms
		FROM job_metrics
		WHERE recorded_at >= ? AND duration_ms IS NOT NULL
	`
	percentileArgs := []interface{}{cutoff.Format("2006-01-02 15:04:05")}

	if queueName != "" {
		percentileQuery += " AND queue_name = ?"
		percentileArgs = append(percentileArgs, queueName)
	}

	if jobType != "" {
		percentileQuery += " AND job_type = ?"
		percentileArgs = append(percentileArgs, jobType)
	}

	percentileQuery += " ORDER BY duration_ms"

	rows, err := js.db.Query(percentileQuery, percentileArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get percentile data: %w", err)
	}
	defer rows.Close()

	var durations []int64
	for rows.Next() {
		var duration int64
		if err := rows.Scan(&duration); err != nil {
			continue
		}
		durations = append(durations, duration)
	}

	p50, p95, p99 := calculatePercentiles(durations)

	return &REDMetrics{
		JobsPerSecond:    jobsPerSecond,
		ErrorRate:        errorRate,
		P50LatencyMs:     p50,
		P95LatencyMs:     p95,
		P99LatencyMs:     p99,
		TotalJobs:        totalJobs,
		FailedJobs:       failedJobs,
		AverageLatencyMs: avgLatency,
	}, nil
}

func (js *SQLiteJobSystem) startWorkerPool(queueName string, concurrency int) {
	pool := &workerPool{
		queueName:   queueName,
		concurrency: concurrency,
		jobCh:       make(chan *Job, concurrency*2),
		stopCh:      make(chan struct{}),
	}

	js.workers[queueName] = pool

	for i := 0; i < concurrency; i++ {
		worker := &worker{
			id:     i,
			pool:   pool,
			jobSys: js,
			stopCh: make(chan struct{}),
			doneCh: make(chan struct{}),
		}
		pool.workers = append(pool.workers, worker)

		js.wg.Add(1)
		go worker.start()
	}

	js.wg.Add(1)
	go js.jobPoller(pool)

	log.Printf("Started worker pool for queue '%s' with %d workers", queueName, concurrency)
}

func (js *SQLiteJobSystem) stopWorkerPool(pool *workerPool) {
	close(pool.stopCh)

	for _, worker := range pool.workers {
		close(worker.stopCh)
		<-worker.doneCh
	}
}

func (js *SQLiteJobSystem) jobPoller(pool *workerPool) {
	defer js.wg.Done()

	ticker := time.NewTicker(js.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			js.pollJobs(pool)
		case <-pool.stopCh:
			close(pool.jobCh)
			return
		}
	}
}

func (js *SQLiteJobSystem) pollJobs(pool *workerPool) {
	// First, get available jobs with their current versions
	selectQuery := `
		SELECT id, queue_name, job_type, payload, status, priority, max_retries, retry_count, run_at, idempotency_key, version
		FROM jobs
		WHERE queue_name = ? AND status = 'pending' AND run_at <= datetime('now')
		ORDER BY priority DESC, run_at ASC
		LIMIT ?
	`

	rows, err := js.db.Query(selectQuery, pool.queueName, pool.concurrency*2)
	if err != nil {
		log.Printf("Failed to poll jobs for queue %s: %v", pool.queueName, err)
		return
	}
	defer rows.Close()

	// Try to claim each job using optimistic locking
	for rows.Next() {
		job := &Job{}
		var runAtStr string
		var payloadStr string

		err := rows.Scan(
			&job.ID, &job.QueueName, &job.JobType, &payloadStr,
			&job.Status, &job.Priority, &job.MaxRetries, &job.RetryCount, &runAtStr, &job.IdempotencyKey, &job.Version,
		)
		if err != nil {
			log.Printf("Failed to scan job: %v", err)
			continue
		}

		// Try to claim this job using version-based optimistic concurrency control
		startedAt := time.Now()
		expectedVersion := job.Version
		claimQuery := `
			UPDATE jobs
			SET status = 'running', started_at = ?, updated_at = ?, version = version + 1
			WHERE id = ? AND version = ? AND status = 'pending'
		`
		result, err := js.db.Exec(claimQuery,
			startedAt.Format("2006-01-02 15:04:05"),
			time.Now().Format("2006-01-02 15:04:05"),
			job.ID,
			expectedVersion,
		)
		if err != nil {
			log.Printf("Failed to claim job %s: %v", job.ID, err)
			continue
		}

		// Check if we successfully claimed the job
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Printf("Failed to check rows affected for job %s: %v", job.ID, err)
			continue
		}
		if rowsAffected == 0 {
			// Either another worker claimed this job OR version mismatch (job was updated)
			// Either way, this job is no longer available to us - move to next
			continue
		}

		// Successfully claimed with version-based locking! Update job version and parse
		if unMarshalErr := json.Unmarshal([]byte(payloadStr), &job.Payload); unMarshalErr != nil {
			log.Printf("Failed to unmarshal job payload: %v", unMarshalErr)
			// Mark job as failed since we can't process it
			js.markJobFailedByID(job.ID, fmt.Sprintf("Failed to unmarshal payload: %v", unMarshalErr))
			continue
		}

		// Try parsing with multiple time formats
		if job.RunAt, err = time.Parse("2006-01-02 15:04:05", runAtStr); err != nil {
			if job.RunAt, err = time.Parse(time.RFC3339, runAtStr); err != nil {
				log.Printf("Failed to parse run_at time: %v", err)
				js.markJobFailedByID(job.ID, fmt.Sprintf("Failed to parse run_at time: %v", err))
				continue
			}
		}

		// Update job status, version, and timestamps
		job.Status = JobStatusRunning
		job.StartedAt = &startedAt
		job.Version = expectedVersion + 1 // Reflect the version increment from the UPDATE

		select {
		case pool.jobCh <- job:
			// Job successfully sent to worker
		case <-pool.stopCh:
			return
		default:
			// Channel full, put job back to pending (increment version since we're changing state)
			resetQuery := `UPDATE jobs SET status = 'pending', started_at = NULL, version = version + 1, updated_at = datetime('now') WHERE id = ?`
			_, err := js.db.Exec(resetQuery, job.ID)
			if err != nil {
				log.Printf("Failed to reset job %s to pending: %v", job.ID, err)
			}
			return
		}
	}
}

func (w *worker) start() {
	defer w.jobSys.wg.Done()
	defer close(w.doneCh)

	for {
		select {
		case job := <-w.pool.jobCh:
			if job != nil {
				w.processJob(job)
			}
		case <-w.stopCh:
			return
		}
	}
}

func (w *worker) processJob(job *Job) {
	w.jobSys.mu.RLock()
	handler, exists := w.jobSys.handlers[job.JobType]
	w.jobSys.mu.RUnlock()

	if !exists {
		w.markJobFailed(job, fmt.Sprintf("no handler registered for job type: %s", job.JobType))
		return
	}

	if err := w.markJobRunning(job); err != nil {
		log.Printf("Failed to mark job %s as running: %v", job.ID, err)
		return
	}

	startTime := time.Now()
	ctx := context.Background()
	err := handler(ctx, job)
	duration := time.Since(startTime)

	w.recordMetric(job, duration, err)

	if err != nil {
		if job.RetryCount < job.MaxRetries {
			w.scheduleRetry(job, err)
		} else {
			w.markJobFailed(job, err.Error())
		}
	} else {
		w.markJobCompleted(job)
	}
}

func (w *worker) markJobRunning(job *Job) error {
	_, err := w.jobSys.db.Exec(
		"UPDATE jobs SET status = 'running', started_at = datetime('now'), version = version + 1, updated_at = datetime('now') WHERE id = ?",
		job.ID,
	)
	if err == nil {
		job.Version++ // Update local version to match database
	}
	return err
}

func (w *worker) markJobCompleted(job *Job) {
	_, err := w.jobSys.db.Exec(
		"UPDATE jobs SET status = 'completed', completed_at = datetime('now'), version = version + 1, updated_at = datetime('now') WHERE id = ?",
		job.ID,
	)
	if err != nil {
		log.Printf("Failed to mark job %s as completed: %v", job.ID, err)
	}
}

func (w *worker) markJobFailed(job *Job, errorMsg string) {
	_, err := w.jobSys.db.Exec(
		"UPDATE jobs SET status = 'failed', error = ?, completed_at = datetime('now'), version = version + 1, updated_at = datetime('now') WHERE id = ?",
		errorMsg, job.ID,
	)
	if err != nil {
		log.Printf("Failed to mark job %s as failed: %v", job.ID, err)
	}
}

func (w *worker) scheduleRetry(job *Job, err error) {
	backoff := w.calculateBackoff(job.RetryCount)
	retryAt := time.Now().Add(backoff)

	_, dbErr := w.jobSys.db.Exec(
		"UPDATE jobs SET status = 'pending', retry_count = retry_count + 1, run_at = ?, error = ?, version = version + 1, updated_at = datetime('now') WHERE id = ?",
		retryAt.Format("2006-01-02 15:04:05"), err.Error(), job.ID,
	)
	if dbErr != nil {
		log.Printf("Failed to schedule retry for job %s: %v", job.ID, dbErr)
	}
}

func (w *worker) calculateBackoff(retryCount int) time.Duration {
	base := w.jobSys.config.RetryBackoffBase
	backoff := time.Duration(1<<uint(retryCount)) * base

	if backoff > w.jobSys.config.RetryBackoffMax {
		backoff = w.jobSys.config.RetryBackoffMax
	}

	return backoff
}

func (w *worker) recordMetric(job *Job, duration time.Duration, err error) {
	status := JobStatusCompleted
	if err != nil {
		status = JobStatusFailed
	}

	durationMs := duration.Milliseconds()

	_, dbErr := w.jobSys.db.Exec(`
		INSERT INTO job_metrics (queue_name, job_type, status, duration_ms, recorded_at)
		VALUES (?, ?, ?, ?, datetime('now'))
	`, job.QueueName, job.JobType, status, durationMs)

	if dbErr != nil {
		log.Printf("Failed to record metrics for job %s: %v", job.ID, dbErr)
	}
}

func calculatePercentiles(values []int64) (p50, p95, p99 float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	p50Idx := len(values) * 50 / 100
	if p50Idx >= len(values) {
		p50Idx = len(values) - 1
	}
	p50 = float64(values[p50Idx])

	p95Idx := len(values) * 95 / 100
	if p95Idx >= len(values) {
		p95Idx = len(values) - 1
	}
	p95 = float64(values[p95Idx])

	p99Idx := len(values) * 99 / 100
	if p99Idx >= len(values) {
		p99Idx = len(values) - 1
	}
	p99 = float64(values[p99Idx])

	return p50, p95, p99
}

// markJobFailedByID marks a job as failed by ID - used during job polling when we can't process a job
func (js *SQLiteJobSystem) markJobFailedByID(jobID, errorMsg string) {
	_, err := js.db.Exec(
		"UPDATE jobs SET status = 'failed', error = ?, completed_at = datetime('now'), version = version + 1, updated_at = datetime('now') WHERE id = ?",
		errorMsg, jobID,
	)
	if err != nil {
		log.Printf("Failed to mark job %s as failed: %v", jobID, err)
	}
}
