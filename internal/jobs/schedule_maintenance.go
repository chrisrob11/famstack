package jobs

import (
	"context"
	"fmt"
	"log"
	"time"

	"famstack/internal/jobsystem"
	"famstack/internal/services"
)

// JobEnqueuer interface for job systems that can enqueue jobs
type JobEnqueuer interface {
	Enqueue(req *jobsystem.EnqueueRequest) (string, error)
}

type ScheduleMaintenancePayload struct {
	// Empty payload - this job scans all schedules
}

func NewScheduleMaintenanceHandler(serviceRegistry *services.Registry, jobSystem JobEnqueuer) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		log.Println("Running schedule maintenance job")

		// Find schedules that need more task generation
		schedules, err := serviceRegistry.Schedules.GetSchedulesNeedingGeneration()
		if err != nil {
			return fmt.Errorf("failed to get schedules needing generation: %w", err)
		}

		if len(schedules) == 0 {
			log.Println("No schedules need task generation")
			return nil
		}

		// Enqueue 3 monthly generation jobs for each schedule that needs it
		for _, schedule := range schedules {
			err := enqueueMonthlyGenerationJobs(jobSystem, schedule.ID)
			if err != nil {
				log.Printf("Failed to enqueue generation jobs for schedule %s: %v", schedule.ID, err)
				continue
			}
			log.Printf("Enqueued 3 monthly generation jobs for schedule %s", schedule.ID)
		}

		log.Printf("Schedule maintenance completed - processed %d schedules", len(schedules))
		return nil
	}
}

func enqueueMonthlyGenerationJobs(jobSystem JobEnqueuer, scheduleID string) error {
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

		idempotencyKey := fmt.Sprintf("schedule:%s:month:%s:%s", scheduleID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		req := &jobsystem.EnqueueRequest{
			QueueName:      "task_generation",
			JobType:        "monthly_task_generation",
			Payload:        payloadMap,
			Priority:       1,
			MaxRetries:     3,
			IdempotencyKey: &idempotencyKey,
		}

		_, err := jobSystem.Enqueue(req)
		if err != nil {
			return fmt.Errorf("failed to enqueue monthly generation job for %s: %w", startDate.Format("2006-01"), err)
		}
	}

	return nil
}
