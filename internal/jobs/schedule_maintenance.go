package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"famstack/internal/database"
	"famstack/internal/jobsystem"
)

type ScheduleMaintenancePayload struct {
	// Empty payload - this job scans all schedules
}

func NewScheduleMaintenanceHandler(db *database.DB, jobSystem *jobsystem.SQLiteJobSystem) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		log.Println("Running schedule maintenance job")

		// Find schedules that need more task generation
		schedules, err := getSchedulesNeedingGeneration(db)
		if err != nil {
			return fmt.Errorf("failed to get schedules needing generation: %w", err)
		}

		if len(schedules) == 0 {
			log.Println("No schedules need task generation")
			return nil
		}

		// Enqueue 3 monthly generation jobs for each schedule that needs it
		for _, scheduleID := range schedules {
			err := enqueueMonthlyGenerationJobs(jobSystem, scheduleID)
			if err != nil {
				log.Printf("Failed to enqueue generation jobs for schedule %s: %v", scheduleID, err)
				continue
			}
			log.Printf("Enqueued 3 monthly generation jobs for schedule %s", scheduleID)
		}

		log.Printf("Schedule maintenance completed - processed %d schedules", len(schedules))
		return nil
	}
}

func getSchedulesNeedingGeneration(db *database.DB) ([]string, error) {
	// Find schedules where last_generated_date is less than 90 days from now
	ninetyDaysFromNow := time.Now().AddDate(0, 0, 90)

	query := `
		SELECT id 
		FROM task_schedules 
		WHERE active = true 
		AND (last_generated_date IS NULL OR last_generated_date < ?)
	`

	rows, err := db.Query(query, ninetyDaysFromNow.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scheduleIDs []string
	for rows.Next() {
		var scheduleID string
		if err := rows.Scan(&scheduleID); err != nil {
			return nil, err
		}
		scheduleIDs = append(scheduleIDs, scheduleID)
	}

	return scheduleIDs, nil
}

func enqueueMonthlyGenerationJobs(jobSystem *jobsystem.SQLiteJobSystem, scheduleID string) error {
	now := time.Now()

	// Enqueue 3 monthly generation jobs starting from next month
	for i := 1; i <= 3; i++ {
		startDate := time.Date(now.Year(), now.Month()+time.Month(i), 1, 0, 0, 0, 0, now.Location())
		endDate := startDate.AddDate(0, 1, -1) // Last day of the month

		payload := MonthlyTaskGenerationPayload{
			ScheduleID: scheduleID,
			StartDate:  startDate.Format("2006-01-02"),
			EndDate:    endDate.Format("2006-01-02"),
		}

		payloadMap := map[string]interface{}{
			"schedule_id": payload.ScheduleID,
			"start_date":  payload.StartDate,
			"end_date":    payload.EndDate,
		}

		req := &jobsystem.EnqueueRequest{
			QueueName:  "task_generation",
			JobType:    "monthly_task_generation",
			Payload:    payloadMap,
			Priority:   1,
			MaxRetries: 3,
		}

		_, err := jobSystem.Enqueue(req)
		if err != nil {
			return fmt.Errorf("failed to enqueue monthly generation job for %s: %w", startDate.Format("2006-01"), err)
		}
	}

	return nil
}
