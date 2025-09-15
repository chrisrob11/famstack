package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/jobsystem"
)

type TaskGenerationPayload struct {
	ScheduleID string `json:"schedule_id"`
	TargetDate string `json:"target_date"`
}

type MonthlyTaskGenerationPayload struct {
	ScheduleID string `json:"schedule_id"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}

type ScheduleDeletionPayload struct {
	ScheduleID string `json:"schedule_id"`
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

func NewMonthlyTaskGenerationHandler(db *database.DB) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		var payload MonthlyTaskGenerationPayload

		payloadBytes, err := json.Marshal(job.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal job payload: %w", err)
		}

		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal monthly task generation payload: %w", err)
		}

		return generateMonthlyTasks(db, payload.ScheduleID, payload.StartDate, payload.EndDate)
	}
}

func generateScheduledTask(db *database.DB, scheduleID, targetDateStr string) error {
	targetDate, err := time.Parse("2006-01-02", targetDateStr)
	if err != nil {
		return fmt.Errorf("invalid target date format: %w", err)
	}

	// Only generate tasks for today and future dates
	today := time.Now().Truncate(24 * time.Hour)
	targetDateTruncated := targetDate.Truncate(24 * time.Hour)
	if targetDateTruncated.Before(today) {
		log.Printf("Skipping task generation for past date %s", targetDateStr)
		return nil
	}

	schedule, err := getTaskSchedule(db, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	weekday := strings.ToLower(targetDate.Weekday().String())
	dayMatches := false
	for _, day := range schedule.DaysOfWeek {
		if weekday == strings.ToLower(day) {
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
	// Check if a task already exists for this schedule on this specific date
	// We should check for tasks that were created to be due on this date,
	// regardless of when they were actually created
	query := `
		SELECT COUNT(*) FROM tasks 
		WHERE schedule_id = ? 
		AND (
			(due_date IS NOT NULL AND DATE(due_date) = ?) OR
			(due_date IS NULL AND DATE(created_at) = ?)
		)
	`

	var count int
	err := db.QueryRow(query, scheduleID, targetDate, targetDate).Scan(&count)
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

	if err != nil && isUniqueConstraintViolation(err) {
		// Task already exists for this schedule on this date - this is expected with concurrent job execution
		log.Printf("Task already exists for schedule %s on %s (concurrent creation detected)", schedule.ID, targetDate.Format("2006-01-02"))
		return nil
	}
	return err
}

// isUniqueConstraintViolation checks if the error is a SQLite unique constraint violation
func isUniqueConstraintViolation(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "UNIQUE constraint failed") &&
		strings.Contains(errMsg, "idx_tasks_schedule_target_date")
}

func generateMonthlyTasks(db *database.DB, scheduleID, startDateStr, endDateStr string) error {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return fmt.Errorf("invalid start date format: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return fmt.Errorf("invalid end date format: %w", err)
	}

	schedule, err := getTaskSchedule(db, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	// Find existing tasks in the date range to avoid duplicates
	existingTasks, err := getExistingTasksInRange(db, scheduleID, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get existing tasks: %w", err)
	}

	existingDates := make(map[string]bool)
	for _, taskDate := range existingTasks {
		existingDates[taskDate] = true
	}

	// Generate all tasks for the month that don't already exist
	today := time.Now().Truncate(24 * time.Hour)
	var tasksToCreate []taskToCreate
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

		task := taskToCreate{
			schedule:   schedule,
			targetDate: current,
		}
		tasksToCreate = append(tasksToCreate, task)
	}

	if len(tasksToCreate) == 0 {
		log.Printf("No new tasks to create for schedule %s in range %s to %s", scheduleID, startDateStr, endDateStr)
		return nil
	}

	// Bulk create tasks
	err = bulkCreateTasks(db, tasksToCreate)
	if err != nil {
		return fmt.Errorf("failed to bulk create tasks: %w", err)
	}

	// Update last_generated_date if this range extends it
	err = updateLastGeneratedDate(db, scheduleID, endDate)
	if err != nil {
		return fmt.Errorf("failed to update last generated date: %w", err)
	}

	log.Printf("Created %d tasks for schedule %s from %s to %s", len(tasksToCreate), scheduleID, startDateStr, endDateStr)
	return nil
}

type taskToCreate struct {
	schedule   *TaskSchedule
	targetDate time.Time
}

func getExistingTasksInRange(db *database.DB, scheduleID string, startDate, endDate time.Time) ([]string, error) {
	// Get existing task dates based on when they were supposed to be due, not when they were created
	query := `
		SELECT DISTINCT 
			CASE 
				WHEN due_date IS NOT NULL THEN DATE(due_date)
				ELSE DATE(created_at)
			END as target_date
		FROM tasks 
		WHERE schedule_id = ? 
		AND (
			(due_date IS NOT NULL AND DATE(due_date) >= ? AND DATE(due_date) <= ?) OR
			(due_date IS NULL AND DATE(created_at) >= ? AND DATE(created_at) <= ?)
		)
	`

	rows, err := db.Query(query, scheduleID,
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date)
	}

	return dates, nil
}

func bulkCreateTasks(db *database.DB, tasks []taskToCreate) error {
	if len(tasks) == 0 {
		return nil
	}

	// Begin transaction for bulk insert
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // nolint:errcheck // Ignore error as we'll commit or it's already rolled back
	}()

	query := `
		INSERT INTO tasks (family_id, assigned_to, title, description, task_type,
						  status, priority, points, due_date, created_by, schedule_id)
		VALUES (?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, task := range tasks {
		var assignedToValue interface{}
		if task.schedule.AssignedTo != nil {
			assignedToValue = *task.schedule.AssignedTo
		} else {
			assignedToValue = nil
		}

		var dueDate *time.Time
		if task.schedule.TimeOfDay != nil {
			timeStr := *task.schedule.TimeOfDay
			if timePart, parseErr := time.Parse("15:04", timeStr); parseErr == nil {
				dueDateWithTime := time.Date(
					task.targetDate.Year(), task.targetDate.Month(), task.targetDate.Day(),
					timePart.Hour(), timePart.Minute(), 0, 0, task.targetDate.Location(),
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

		_, err = stmt.Exec(
			task.schedule.FamilyID,
			assignedToValue,
			task.schedule.Title,
			task.schedule.Description,
			task.schedule.TaskType,
			task.schedule.Priority,
			task.schedule.Points,
			dueDateValue,
			task.schedule.CreatedBy,
			task.schedule.ID,
		)
		if err != nil {
			if isUniqueConstraintViolation(err) {
				// Task already exists for this schedule on this date - skip and continue
				log.Printf("Task already exists for schedule %s on %s (concurrent creation detected)", task.schedule.ID, task.targetDate.Format("2006-01-02"))
				continue
			}
			return fmt.Errorf("failed to insert task for date %s: %w", task.targetDate.Format("2006-01-02"), err)
		}
	}

	return tx.Commit()
}

func updateLastGeneratedDate(db *database.DB, scheduleID string, endDate time.Time) error {
	query := `
		UPDATE task_schedules 
		SET last_generated_date = ?
		WHERE id = ? AND (last_generated_date IS NULL OR last_generated_date < ?)
	`

	dateStr := endDate.Format("2006-01-02 15:04:05")
	_, err := db.Exec(query, dateStr, scheduleID, dateStr)
	if err != nil {
		return fmt.Errorf("failed to update last generated date: %w", err)
	}

	return nil
}

func NewScheduleDeletionHandler(db *database.DB) jobsystem.JobHandler {
	return func(ctx context.Context, job *jobsystem.Job) error {
		var payload ScheduleDeletionPayload

		payloadBytes, err := json.Marshal(job.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal job payload: %w", err)
		}

		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return fmt.Errorf("failed to unmarshal schedule deletion payload: %w", err)
		}

		return deleteScheduleAndTasks(db, payload.ScheduleID)
	}
}

func deleteScheduleAndTasks(db *database.DB, scheduleID string) error {
	log.Printf("Starting deletion of schedule %s and all its tasks", scheduleID)

	// Start transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // nolint:errcheck // Ignore error as we'll commit or it's already rolled back
	}()

	// First, delete all tasks associated with this schedule
	deleteTasksQuery := `DELETE FROM tasks WHERE schedule_id = ?`
	result, err := tx.Exec(deleteTasksQuery, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete tasks for schedule %s: %w", scheduleID, err)
	}

	taskRowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get task rows affected: %w", err)
	}
	log.Printf("Deleted %d tasks for schedule %s", taskRowsAffected, scheduleID)

	// Then, delete the schedule itself
	deleteScheduleQuery := `DELETE FROM task_schedules WHERE id = ?`
	result, err = tx.Exec(deleteScheduleQuery, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule %s: %w", scheduleID, err)
	}

	scheduleRowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get schedule rows affected: %w", err)
	}
	if scheduleRowsAffected == 0 {
		return fmt.Errorf("schedule %s not found", scheduleID)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit deletion transaction: %w", err)
	}

	log.Printf("Successfully deleted schedule %s and %d associated tasks", scheduleID, taskRowsAffected)
	return nil
}
