package jobsystem

import (
	"encoding/json"
	"log"
	"time"

	cron "github.com/robfig/cron/v3"
)

func (js *SQLiteJobSystem) startScheduler() {
	js.scheduler = &scheduler{
		jobSys: js,
		ticker: time.NewTicker(js.config.SchedulerInterval),
		stopCh: make(chan struct{}),
	}

	js.wg.Add(1)
	go js.scheduler.run()

	log.Printf("Started scheduler with interval: %v", js.config.SchedulerInterval)
}

func (js *SQLiteJobSystem) stopScheduler() {
	if js.scheduler != nil {
		close(js.scheduler.stopCh)
		js.scheduler.ticker.Stop()
		js.scheduler = nil
	}
}

func (s *scheduler) run() {
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

func (s *scheduler) processScheduledJobs() {
	query := `
		SELECT id, name, queue_name, job_type, payload, cron_expr, next_run_at
		FROM scheduled_jobs
		WHERE enabled = true AND next_run_at <= datetime('now')
		ORDER BY next_run_at ASC
		LIMIT 100
	`

	rows, err := s.jobSys.db.Query(query)
	if err != nil {
		log.Printf("Failed to query scheduled jobs: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var scheduledJob ScheduledJob
		var payloadStr, nextRunAtStr string

		scanErr := rows.Scan(
			&scheduledJob.ID,
			&scheduledJob.Name,
			&scheduledJob.QueueName,
			&scheduledJob.JobType,
			&payloadStr,
			&scheduledJob.CronExpr,
			&nextRunAtStr,
		)
		if scanErr != nil {
			log.Printf("Failed to scan scheduled job: %v", scanErr)
			continue
		}

		if unMarshalErr := json.Unmarshal([]byte(payloadStr), &scheduledJob.Payload); unMarshalErr != nil {
			log.Printf("Failed to unmarshal scheduled job payload: %v", unMarshalErr)
			continue
		}

		var timeParseErr error
		if scheduledJob.NextRunAt, timeParseErr = time.Parse("2006-01-02 15:04:05", nextRunAtStr); timeParseErr != nil {
			log.Printf("Failed to parse next_run_at: %v", timeParseErr)
			continue
		}

		s.executeScheduledJob(&scheduledJob)
	}
}

func (s *scheduler) executeScheduledJob(scheduledJob *ScheduledJob) {
	log.Printf("Executing scheduled job: %s", scheduledJob.Name)

	jobID, err := s.jobSys.Enqueue(&EnqueueRequest{
		QueueName:  scheduledJob.QueueName,
		JobType:    scheduledJob.JobType,
		Payload:    scheduledJob.Payload,
		Priority:   0,
		MaxRetries: s.jobSys.config.DefaultMaxRetries,
	})

	if err != nil {
		log.Printf("Failed to enqueue scheduled job %s: %v", scheduledJob.Name, err)
		return
	}

	nextRun, err := s.jobSys.calculateNextRun(scheduledJob.CronExpr)
	if err != nil {
		log.Printf("Failed to calculate next run for scheduled job %s: %v", scheduledJob.Name, err)
		return
	}

	_, err = s.jobSys.db.Exec(`
		UPDATE scheduled_jobs 
		SET next_run_at = ?, last_run_at = datetime('now'), updated_at = datetime('now')
		WHERE id = ?
	`, nextRun.Format("2006-01-02 15:04:05"), scheduledJob.ID)

	if err != nil {
		log.Printf("Failed to update scheduled job %s: %v", scheduledJob.Name, err)
		return
	}

	log.Printf("Successfully executed scheduled job %s, enqueued as job %s, next run: %v",
		scheduledJob.Name, jobID, nextRun)
}

func (js *SQLiteJobSystem) calculateNextRun(cronExpr string) (time.Time, error) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	return schedule.Next(time.Now()), nil
}

func (js *SQLiteJobSystem) startMetricsCleanup() {
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

func (js *SQLiteJobSystem) cleanupOldMetrics() {
	cutoff := time.Now().Add(-js.config.MetricsRetention)

	result, err := js.db.Exec(
		"DELETE FROM job_metrics WHERE recorded_at < ?",
		cutoff.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		log.Printf("Failed to cleanup old metrics: %v", err)
		return
	}

	if count, err := result.RowsAffected(); err == nil && count > 0 {
		log.Printf("Cleaned up %d old metric records", count)
	}
}
