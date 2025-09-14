package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"famstack/internal/database"
	"famstack/internal/jobsystem"
)

type TaskGenerationPayload struct {
	ScheduleID string `json:"schedule_id"`
	TargetDate string `json:"target_date"`
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

func NewTaskGenerationHandler(db *database.DB) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		var payload TaskGenerationPayload

		payloadBytes, err := json.Marshal(job.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal job payload: %w", err)
		}

		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal task generation payload: %w", err)
		}

		return generateScheduledTask(db, payload.ScheduleID, payload.TargetDate)
	}
}

func generateScheduledTask(db *database.DB, scheduleID, targetDateStr string) error {
	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		return fmt.Errorf("invalid target date format: %w", err)
	}

	schedule, err := getTaskSchedule(db, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	weekday := targetDate.Weekday().String()
	dayMatches := false
	for _, day := range schedule.DaysOfWeek {
		if weekday == day {
			dayMatches = true
			break
		}
	}

	if !dayMatches {
		log.Printf("Target date %s (%s) doesn't match schedule days: %v", targetDateStr, weekday, schedule.DaysOfWeek)
		return nil
	}

	existingTask, err := checkExistingTask(db, scheduleID, targetDateStr)
	if err != nil {
		return fmt.Errorf("failed to check existing task: %w", err)
	}
	if existingTask {
		log.Printf("Task already exists for schedule %s on %s", scheduleID, targetDateStr)
		return nil
	}

	err = createTaskFromSchedule(db, schedule, targetDate)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("Created task for schedule %s on %s", scheduleID, targetDateStr)
	return nil
}

func getTaskSchedule(db *database.DB, scheduleID string) (*TaskSchedule, error) {
	query := `
		SELECT id, family_id, created_by, title, description, task_type,
			   assigned_to, days_of_week, time_of_day, priority, points
		FROM task_schedules 
		WHERE id = ? AND active = true
	`

	var schedule TaskSchedule
	var assignedTo sql.NullString
	var timeOfDay sql.NullString
	var daysOfWeekJSON string

	err := db.QueryRow(query, scheduleID).Scan(
		&schedule.ID,
		&schedule.FamilyID,
		&schedule.CreatedBy,
		&schedule.Title,
		&schedule.Description,
		&schedule.TaskType,
		&assignedTo,
		&daysOfWeekJSON,
		&timeOfDay,
		&schedule.Priority,
		&schedule.Points,
	)

	if err != nil {
		return nil, err
	}

	if assignedTo.Valid {
		schedule.AssignedTo = &assignedTo.String
	}
	if timeOfDay.Valid {
		schedule.TimeOfDay = &timeOfDay.String
	}

	err = json.Unmarshal([]byte(daysOfWeekJSON), &schedule.DaysOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to parse days_of_week: %w", err)
	}

	return &schedule, nil
}

func checkExistingTask(db *database.DB, scheduleID, targetDate string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM tasks 
		WHERE schedule_id = ? 
		AND DATE(created_at) = ?
	`

	var count int
	err := db.QueryRow(query, scheduleID, targetDate).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func createTaskFromSchedule(db *database.DB, schedule *TaskSchedule, targetDate time.Time) error {
	var assignedToValue interface{}
	if schedule.AssignedTo != nil {
		assignedToValue = *schedule.AssignedTo
	} else {
		assignedToValue = nil
	}

	var dueDate *time.Time
	if schedule.TimeOfDay != nil {
		timeStr := *schedule.TimeOfDay
		if timePart, err := time.Parse("15:04", timeStr); err == nil {
			dueDateWithTime := time.Date(
				targetDate.Year(), targetDate.Month(), targetDate.Day(),
				timePart.Hour(), timePart.Minute(), 0, 0, targetDate.Location(),
			)
			dueDate = &dueDateWithTime
		}
	}

	var dueDateValue interface{}
	if dueDate != nil {
		dueDateValue = dueDate.Format("2006-01-02 15:04:05")
	} else {
		dueDateValue = nil
	}

	query := `
		INSERT INTO tasks (family_id, assigned_to, title, description, task_type,
						  status, priority, points, due_date, created_by, schedule_id)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?)
		RETURNING id
	`

	var newID string
	err := db.QueryRow(query,
		schedule.FamilyID,
		assignedToValue,
		schedule.Title,
		schedule.Description,
		schedule.TaskType,
		schedule.Priority,
		schedule.Points,
		dueDateValue,
		schedule.CreatedBy,
		schedule.ID,
	).Scan(&newID)

	return err
}
