package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"famstack/internal/jobsystem"
	"famstack/internal/models"
	"famstack/internal/services"
)

type MonthlyTaskGenerationPayload struct {
	ScheduleID string `json:"schedule_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type ScheduleDeletionPayload struct {
	ScheduleID string `json:"schedule_id"`
}

func NewMonthlyTaskGenerationHandler(serviceRegistry *services.Registry) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		var payload MonthlyTaskGenerationPayload

		payloadBytes, err := json.Marshal(job.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal job payload: %w", err)
		}

		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal monthly task generation payload: %w", err)
		}

		return generateMonthlyTasks(serviceRegistry, payload.ScheduleID, payload.StartDate, payload.EndDate)
	}
}

func convertScheduleToLegacyFormat(schedule *models.TaskSchedule) *TaskSchedule {
	var daysOfWeek []string
	if schedule.DaysOfWeek != nil {
		// Parse JSON days of week
		if err := json.Unmarshal([]byte(*schedule.DaysOfWeek), &daysOfWeek); err != nil {
			// Log error but continue with empty days of week
			log.Printf("Failed to unmarshal days of week for schedule %s: %v", schedule.ID, err)
		}
	}

	return &TaskSchedule{
		ID:        schedule.ID,
		FamilyID:  schedule.FamilyID,
		CreatedBy: schedule.CreatedBy,
		Title:     schedule.Title,
		Description: func() string {
			if schedule.Description != nil {
				return *schedule.Description
			} else {
				return ""
			}
		}(),
		TaskType:   schedule.TaskType,
		AssignedTo: schedule.AssignedTo,
		DaysOfWeek: daysOfWeek,
		TimeOfDay:  schedule.TimeOfDay,
		Priority:   schedule.Priority,
		Points:     schedule.Points,
	}
}

type TaskSchedule struct {
	ID          string
	FamilyID    string
	CreatedBy   string
	Title       string
	Description string
	TaskType    string
	AssignedTo  *string
	DaysOfWeek  []string
	TimeOfDay   *string
	Priority    int
	Points      int
}

func generateMonthlyTasks(serviceRegistry *services.Registry, scheduleID, startDateStr, endDateStr string) error {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	scheduleModel, err := serviceRegistry.Schedules.GetSchedule(scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	schedule := convertScheduleToLegacyFormat(scheduleModel)

	// Find existing tasks in the date range to avoid duplicates
	existingTasks, err := serviceRegistry.Tasks.GetExistingTasksInRange(scheduleID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get existing tasks: %w", err)
	}

	existingDates := make(map[string]bool)
	for _, taskDate := range existingTasks {
		existingDates[taskDate] = true
	}

	// Generate all tasks for the month that don't already exist
	today := time.Now().Truncate(24 * time.Hour)
	var tasksToCreate []services.BulkTaskRequest
	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		// Only generate tasks for today and future dates
		currentTruncated := current.Truncate(24 * time.Hour)
		if currentTruncated.Before(today) {
			continue
		}

		weekday := strings.ToLower(current.Weekday().String())
		dayMatches := false
		for _, day := range schedule.DaysOfWeek {
			if weekday == strings.ToLower(day) {
				dayMatches = true
				break
			}
		}

		if !dayMatches {
			continue
		}

		dateStr := current.Format("2006-01-02")
		if existingDates[dateStr] {
			log.Printf("Task already exists for schedule %s on %s, skipping", scheduleID, dateStr)
			continue
		}

		var dueDate *time.Time
		if schedule.TimeOfDay != nil {
			timeStr := *schedule.TimeOfDay
			if timePart, parseErr := time.Parse("15:04", timeStr); parseErr == nil {
				dueDateWithTime := time.Date(
					current.Year(), current.Month(), current.Day(),
					timePart.Hour(), timePart.Minute(), 0, 0, current.Location(),
				)
				dueDate = &dueDateWithTime
			}
		}

		task := services.BulkTaskRequest{
			Title:       schedule.Title,
			Description: schedule.Description,
			TaskType:    schedule.TaskType,
			AssignedTo:  schedule.AssignedTo,
			Priority:    schedule.Priority,
			Points:      schedule.Points,
			DueDate:     dueDate,
			ScheduleID:  schedule.ID,
		}
		tasksToCreate = append(tasksToCreate, task)
	}

	if len(tasksToCreate) == 0 {
		log.Printf("No new tasks to create for schedule %s in range %s to %s", scheduleID, startDateStr, endDateStr)
		return nil
	}

	// Bulk create tasks
	err = serviceRegistry.Tasks.BulkCreateTasks(schedule.FamilyID, schedule.CreatedBy, tasksToCreate)
	if err != nil {
		return fmt.Errorf("failed to bulk create tasks: %w", err)
	}

	// Update last_generated_date if this range extends it
	err = serviceRegistry.Schedules.UpdateLastGeneratedDate(scheduleID, endDate)
	if err != nil {
		return fmt.Errorf("failed to update last generated date: %w", err)
	}

	log.Printf("Created %d tasks for schedule %s from %s to %s", len(tasksToCreate), scheduleID, startDateStr, endDateStr)
	return nil
}

func NewScheduleDeletionHandler(serviceRegistry *services.Registry) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		var payload ScheduleDeletionPayload

		payloadBytes, err := json.Marshal(job.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal job payload: %w", err)
		}

		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal schedule deletion payload: %w", err)
		}

		return deleteScheduleAndTasks(serviceRegistry, payload.ScheduleID)
	}
}

func deleteScheduleAndTasks(serviceRegistry *services.Registry, scheduleID string) error {
	log.Printf("Starting deletion of schedule %s and all its tasks", scheduleID)

	// Use the schedule service to delete schedule and all its tasks
	err := serviceRegistry.Schedules.DeleteScheduleWithTasks(scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule and tasks: %w", err)
	}

	log.Printf("Successfully deleted schedule %s and all associated tasks", scheduleID)
	return nil
}
