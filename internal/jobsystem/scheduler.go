package jobsystem

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
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
	return parseCronExpression(cronExpr)
}

func parseCronExpression(expr string) (time.Time, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return time.Time{}, fmt.Errorf("invalid cron expression: must have 5 fields (minute hour day month weekday)")
	}

	now := time.Now()

	minute, err := parseCronField(fields[0], 0, 59)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid minute field: %w", err)
	}

	hour, err := parseCronField(fields[1], 0, 23)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid hour field: %w", err)
	}

	day, err := parseCronField(fields[2], 1, 31)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day field: %w", err)
	}

	month, err := parseCronField(fields[3], 1, 12)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month field: %w", err)
	}

	weekday, err := parseCronField(fields[4], 0, 6)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid weekday field: %w", err)
	}

	return calculateNextCronTime(now, minute, hour, day, month, weekday)
}

func parseCronField(field string, min, max int) ([]int, error) {
	if field == "*" {
		var values []int
		for i := min; i <= max; i++ {
			values = append(values, i)
		}
		return values, nil
	}

	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		var values []int
		for _, part := range parts {
			val, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			if val < min || val > max {
				return nil, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
			}
			values = append(values, val)
		}
		return values, nil
	}

	if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid step format: %s", field)
		}

		start := min
		if parts[0] != "*" {
			var err error
			start, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start value: %s", parts[0])
			}
		}

		step, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid step value: %s", parts[1])
		}

		if step <= 0 {
			return nil, fmt.Errorf("step must be positive: %d", step)
		}

		var values []int
		for i := start; i <= max; i += step {
			values = append(values, i)
		}
		return values, nil
	}

	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range format: %s", field)
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid range start: %s", parts[0])
		}

		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid range end: %s", parts[1])
		}

		if start > end {
			return nil, fmt.Errorf("range start must be <= end: %d > %d", start, end)
		}

		var values []int
		for i := start; i <= end; i++ {
			values = append(values, i)
		}
		return values, nil
	}

	val, err := strconv.Atoi(field)
	if err != nil {
		return nil, fmt.Errorf("invalid number: %s", field)
	}

	if val < min || val > max {
		return nil, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
	}

	return []int{val}, nil
}

func calculateNextCronTime(now time.Time, minutes, hours, days, months, weekdays []int) (time.Time, error) {
	next := now.Add(time.Minute)
	next = next.Truncate(time.Minute)

	for i := 0; i < 366*24*60; i++ {
		if contains(months, int(next.Month())) &&
			(contains(days, next.Day()) || contains(weekdays, int(next.Weekday()))) &&
			contains(hours, next.Hour()) &&
			contains(minutes, next.Minute()) {
			return next, nil
		}
		next = next.Add(time.Minute)
	}

	return time.Time{}, fmt.Errorf("could not find next execution time within one year")
}

func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
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
