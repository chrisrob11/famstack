package jobsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"famstack/internal/services"

	cron "github.com/robfig/cron/v3"
)

type DBJobSystem struct {
	jobsService    *services.JobsService
	config         *Config
	handlers       map[string]JobHandler
	workers        map[string]*dbWorkerPool
	scheduler      *dbScheduler
	metricsCleanup *time.Ticker
	shutdownCh     chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	running        bool
}

type dbWorkerPool struct {
	queueName   string
	concurrency int
	workers     []*dbWorker
	jobCh       chan *Job
	stopCh      chan struct{}
}

type dbWorker struct {
	id     int
	pool   *dbWorkerPool
	jobSys *DBJobSystem
	stopCh chan struct{}
	doneCh chan struct{}
}

type dbScheduler struct {
	jobSys *DBJobSystem
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewDBJobSystem(config *Config, jobsService *services.JobsService) *DBJobSystem {
	if config == nil {
		config = DefaultConfig()
	}

	return &DBJobSystem{
		jobsService: jobsService,
		config:      config,
		handlers:    make(map[string]JobHandler),
		workers:     make(map[string]*dbWorkerPool),
		shutdownCh:  make(chan struct{}),
	}
}

func (js *DBJobSystem) Enqueue(req *EnqueueRequest) (string, error) {
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

	return js.jobsService.EnqueueJob(
		req.QueueName,
		req.JobType,
		string(payloadBytes),
		req.Priority,
		req.MaxRetries,
		runAt,
		req.IdempotencyKey,
	)
}

func (js *DBJobSystem) Schedule(req *ScheduleRequest) error {
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

	return js.jobsService.ScheduleJob(
		req.Name,
		req.QueueName,
		req.JobType,
		string(payloadBytes),
		req.CronExpr,
		req.Enabled,
		nextRunAt,
	)
}

func (js *DBJobSystem) Register(jobType string, handler JobHandler) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.handlers[jobType] = handler
}

func (js *DBJobSystem) Start(ctx context.Context) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if js.running {
		return fmt.Errorf("job system is already running")
	}

	log.Println("Starting DB job system...")

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

	log.Println("DB job system started successfully")
	return nil
}

func (js *DBJobSystem) Stop() {
	js.mu.Lock()
	defer js.mu.Unlock()

	if !js.running {
		return
	}

	log.Println("Stopping DB job system...")

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

	log.Println("DB job system stopped")
}

func (js *DBJobSystem) GetMetrics(queueName, jobType string) (*REDMetrics, error) {
	timeWindow := 1 * time.Hour
	metrics, err := js.jobsService.GetJobMetrics(queueName, jobType, timeWindow)
	if err != nil {
		return nil, fmt.Errorf("failed to get job metrics: %w", err)
	}

	return &REDMetrics{
		JobsPerSecond:    metrics.JobsPerSecond,
		ErrorRate:        metrics.ErrorRate,
		P50LatencyMs:     metrics.AverageLatencyMs, // Using average as approximation
		P95LatencyMs:     metrics.AverageLatencyMs, // Using average as approximation
		P99LatencyMs:     metrics.AverageLatencyMs, // Using average as approximation
		TotalJobs:        metrics.TotalJobs,
		FailedJobs:       metrics.FailedJobs,
		AverageLatencyMs: metrics.AverageLatencyMs,
	}, nil
}

func (js *DBJobSystem) startWorkerPool(queueName string, concurrency int) {
	pool := &dbWorkerPool{
		queueName:   queueName,
		concurrency: concurrency,
		jobCh:       make(chan *Job, concurrency*2),
		stopCh:      make(chan struct{}),
	}

	js.workers[queueName] = pool

	for i := 0; i < concurrency; i++ {
		worker := &dbWorker{
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

func (js *DBJobSystem) stopWorkerPool(pool *dbWorkerPool) {
	close(pool.stopCh)

	for _, worker := range pool.workers {
		close(worker.stopCh)
		<-worker.doneCh
	}
}

func (js *DBJobSystem) jobPoller(pool *dbWorkerPool) {
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

func (js *DBJobSystem) pollJobs(pool *dbWorkerPool) {
	// Get pending jobs using the service
	jobs, err := js.jobsService.GetPendingJobs(pool.queueName, pool.concurrency*2)
	if err != nil {
		log.Printf("Failed to poll jobs for queue %s: %v", pool.queueName, err)
		return
	}

	// Try to claim each job using optimistic locking
	for _, serviceJob := range jobs {
		// Convert service job to job system job
		job := &Job{
			ID:             serviceJob.ID,
			QueueName:      serviceJob.QueueName,
			JobType:        serviceJob.JobType,
			Payload:        serviceJob.Payload,
			Status:         JobStatus(serviceJob.Status),
			Priority:       serviceJob.Priority,
			MaxRetries:     serviceJob.MaxRetries,
			RetryCount:     serviceJob.RetryCount,
			RunAt:          serviceJob.RunAt,
			StartedAt:      serviceJob.StartedAt,
			CompletedAt:    serviceJob.CompletedAt,
			Error:          &serviceJob.Error,
			IdempotencyKey: serviceJob.IdempotencyKey,
			Version:        int64(serviceJob.Version),
			CreatedAt:      serviceJob.CreatedAt,
			UpdatedAt:      serviceJob.UpdatedAt,
		}

		// Try to claim this job using optimistic locking
		claimed, err := js.jobsService.ClaimJob(job.ID, int(job.Version))
		if err != nil {
			log.Printf("Failed to claim job %s: %v", job.ID, err)
			continue
		}

		if !claimed {
			// Job was already claimed by another worker
			continue
		}

		// Successfully claimed! Update job status and timestamps
		startedAt := time.Now()
		job.Status = JobStatusRunning
		job.StartedAt = &startedAt
		job.Version++ // Reflect the version increment from the claim

		select {
		case pool.jobCh <- job:
			// Job successfully sent to worker
		case <-pool.stopCh:
			return
		default:
			// Channel full, reset job to pending
			if err := js.jobsService.ResetJobToPending(job.ID); err != nil {
				log.Printf("Failed to reset job %s to pending: %v", job.ID, err)
			}
			return
		}
	}
}

func (w *dbWorker) start() {
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

func (w *dbWorker) processJob(job *Job) {
	w.jobSys.mu.RLock()
	handler, exists := w.jobSys.handlers[job.JobType]
	w.jobSys.mu.RUnlock()

	if !exists {
		w.markJobFailed(job, fmt.Sprintf("no handler registered for job type: %s", job.JobType))
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

func (w *dbWorker) markJobCompleted(job *Job) {
	if err := w.jobSys.jobsService.MarkJobCompleted(job.ID); err != nil {
		log.Printf("Failed to mark job %s as completed: %v", job.ID, err)
	}
}

func (w *dbWorker) markJobFailed(job *Job, errorMsg string) {
	if err := w.jobSys.jobsService.MarkJobFailed(job.ID, errorMsg); err != nil {
		log.Printf("Failed to mark job %s as failed: %v", job.ID, err)
	}
}

func (w *dbWorker) scheduleRetry(job *Job, err error) {
	backoff := w.calculateBackoff(job.RetryCount)
	retryAt := time.Now().Add(backoff)

	if dbErr := w.jobSys.jobsService.ScheduleJobRetry(job.ID, retryAt, err.Error()); dbErr != nil {
		log.Printf("Failed to schedule retry for job %s: %v", job.ID, dbErr)
	}
}

func (w *dbWorker) calculateBackoff(retryCount int) time.Duration {
	base := w.jobSys.config.RetryBackoffBase
	backoff := time.Duration(1<<uint(retryCount)) * base

	if backoff > w.jobSys.config.RetryBackoffMax {
		backoff = w.jobSys.config.RetryBackoffMax
	}

	return backoff
}

func (w *dbWorker) recordMetric(job *Job, duration time.Duration, err error) {
	status := JobStatusCompleted
	if err != nil {
		status = JobStatusFailed
	}

	durationMs := duration.Milliseconds()

	if dbErr := w.jobSys.jobsService.RecordJobMetric(job.QueueName, job.JobType, string(status), durationMs); dbErr != nil {
		log.Printf("Failed to record metrics for job %s: %v", job.ID, dbErr)
	}
}

// calculateNextRun calculates the next execution time for a cron expression
func (js *DBJobSystem) calculateNextRun(cronExpr string) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(time.Now()), nil
}

// startScheduler starts the scheduled jobs processor
func (js *DBJobSystem) startScheduler() {
	js.scheduler = &dbScheduler{
		jobSys: js,
		ticker: time.NewTicker(js.config.SchedulerInterval),
		stopCh: make(chan struct{}),
	}

	js.wg.Add(1)
	go js.scheduler.run()

	log.Printf("Started scheduler with interval: %v", js.config.SchedulerInterval)
}

// stopScheduler stops the scheduled jobs processor
func (js *DBJobSystem) stopScheduler() {
	if js.scheduler != nil {
		close(js.scheduler.stopCh)
		js.scheduler.ticker.Stop()
		js.scheduler = nil
	}
}

// startMetricsCleanup starts the periodic metrics cleanup routine
func (js *DBJobSystem) startMetricsCleanup() {
	js.metricsCleanup = time.NewTicker(1 * time.Hour)

	js.wg.Add(1)
	go func() {
		defer js.wg.Done()
		for {
			select {
			case <-js.metricsCleanup.C:
				js.cleanupOldMetrics()
			case <-js.shutdownCh:
				return
			}
		}
	}()

	log.Println("Started metrics cleanup routine")
}

// cleanupOldMetrics removes old job metrics based on retention policy
func (js *DBJobSystem) cleanupOldMetrics() {
	cutoff := time.Now().Add(-js.config.MetricsRetention)
	// Note: This would need a method in JobsService to clean up metrics
	// For now, we'll log that cleanup would happen here
	log.Printf("Would clean up metrics older than: %v", cutoff)
}

// Scheduler methods for DBJobSystem

func (s *dbScheduler) run() {
	defer s.jobSys.wg.Done()

	for {
		select {
		case <-s.ticker.C:
			s.processScheduledJobs()
		case <-s.stopCh:
			return
		}
	}
}

func (s *dbScheduler) processScheduledJobs() {
	// This would need to be implemented using JobsService methods
	// For now, we'll log that scheduled job processing would happen here
	log.Println("Processing scheduled jobs...")
}
