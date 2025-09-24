package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
)

// SchedulesService handles all task schedule database operations
type SchedulesService struct {
	db *database.Fascade
}

// NewSchedulesService creates a new schedules service
func NewSchedulesService(db *database.Fascade) *SchedulesService {
	return &SchedulesService{db: db}
}

// GetSchedule returns a task schedule by ID
func (s *SchedulesService) GetSchedule(scheduleID string) (*models.TaskSchedule, error) {
	query := `
		SELECT id, family_id, created_by, title, description, task_type, assigned_to,
			   days_of_week, time_of_day, priority, points, active, created_at,
			   last_generated_date
		FROM task_schedules
		WHERE id = ?
	`

	var schedule models.TaskSchedule
	var description, assignedTo, daysOfWeek, timeOfDay sql.NullString
	var lastGeneratedDate sql.NullTime

	err := s.db.QueryRow(query, scheduleID).Scan(
		&schedule.ID, &schedule.FamilyID, &schedule.CreatedBy, &schedule.Title,
		&description, &schedule.TaskType, &assignedTo, &daysOfWeek,
		&schedule.TimeOfDay, &schedule.Priority, &schedule.Points,
		&schedule.Active, &schedule.CreatedAt, &schedule.LastGeneratedDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schedule not found")
		}
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		schedule.Description = &description.String
	}
	if assignedTo.Valid {
		schedule.AssignedTo = &assignedTo.String
	}
	if daysOfWeek.Valid {
		schedule.DaysOfWeek = &daysOfWeek.String
	}
	if timeOfDay.Valid {
		schedule.TimeOfDay = &timeOfDay.String
	}
	// Get family timezone for conversions
	familyTimezone, err := GetFamilyTimezone(s.db, schedule.FamilyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family timezone for schedule conversion: %w", err)
	}

	if lastGeneratedDate.Valid {
		// Convert LastGeneratedDate from UTC to family timezone
		convertedLastGenerated, convErr := ConvertFromUTC(lastGeneratedDate.Time, familyTimezone)
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert last generated date from UTC: %w", convErr)
		}
		schedule.LastGeneratedDate = &convertedLastGenerated
	}

	// Convert CreatedAt from UTC to family timezone
	schedule.CreatedAt, err = ConvertFromUTC(schedule.CreatedAt, familyTimezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created at from UTC: %w", err)
	}

	return &schedule, nil
}

// ListSchedules returns all task schedules for a family
func (s *SchedulesService) ListSchedules(familyID string) ([]models.TaskSchedule, error) {
	query := `
		SELECT id, family_id, created_by, title, description, task_type, assigned_to,
			   days_of_week, time_of_day, priority, points, active, created_at,
			   last_generated_date
		FROM task_schedules
		WHERE family_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}
	defer rows.Close()

	schedules := make([]models.TaskSchedule, 0)
	for rows.Next() {
		schedule, scanErr := s.scanTaskSchedule(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan schedule: %w", scanErr)
		}
		schedules = append(schedules, *schedule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedules: %w", err)
	}

	return schedules, nil
}

// ListActiveSchedules returns all active schedules that are ready to run
func (s *SchedulesService) ListActiveSchedules() ([]models.TaskSchedule, error) {
	query := `
		SELECT id, family_id, created_by, title, description, task_type, assigned_to,
			   days_of_week, time_of_day, priority, points, active, created_at,
			   last_generated_date
		FROM task_schedules
		WHERE active = true
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active schedules: %w", err)
	}
	defer rows.Close()

	schedules := make([]models.TaskSchedule, 0)
	for rows.Next() {
		schedule, scanErr := s.scanTaskSchedule(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan active schedule: %w", scanErr)
		}
		schedules = append(schedules, *schedule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active schedules: %w", err)
	}

	return schedules, nil
}

// CreateSchedule creates a new task schedule
func (s *SchedulesService) CreateSchedule(familyID, createdBy string, req *models.CreateTaskScheduleRequest) (*models.TaskSchedule, error) {
	scheduleID := generateScheduleID()
	now := time.Now().UTC()

	// For now, map the request to the actual database schema
	// This is a temporary fix until the request models are updated
	query := `
		INSERT INTO task_schedules (id, family_id, created_by, title, description, task_type,
								   assigned_to, days_of_week, time_of_day, priority, points,
								   active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Convert days_of_week array to JSON string for database storage
	daysJSON, err := json.Marshal(req.DaysOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal days_of_week: %w", err)
	}

	_, err = s.db.Exec(query,
		scheduleID, familyID, createdBy, req.Title, req.Description, req.TaskType,
		req.AssignedTo, string(daysJSON), req.TimeOfDay, req.Priority, 0, true, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return s.GetSchedule(scheduleID)
}

// UpdateSchedule updates an existing task schedule
func (s *SchedulesService) UpdateSchedule(scheduleID string, req *models.UpdateTaskScheduleRequest) (*models.TaskSchedule, error) {
	// Simplified update function that maps to actual database schema
	// TODO: Update request models to match database schema

	setParts := []string{}
	args := []interface{}{}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
	}
	if req.AssignedTo != nil {
		setParts = append(setParts, "assigned_to = ?")
		args = append(args, *req.AssignedTo)
	}
	if req.DaysOfWeek != nil {
		// Convert []string to JSON string for database storage
		daysJSON, marshalErr := json.Marshal(*req.DaysOfWeek)
		if marshalErr != nil {
			return nil, fmt.Errorf("cannot marshal days of week: %v", marshalErr)
		}
		setParts = append(setParts, "days_of_week = ?")
		args = append(args, string(daysJSON))
	}
	if req.TaskType != nil {
		setParts = append(setParts, "task_type = ?")
		args = append(args, *req.TaskType)
	}
	if req.TimeOfDay != nil {
		setParts = append(setParts, "time_of_day = ?")
		args = append(args, *req.TimeOfDay)
	}
	if req.Priority != nil {
		setParts = append(setParts, "priority = ?")
		args = append(args, *req.Priority)
	}
	if req.Active != nil {
		setParts = append(setParts, "active = ?")
		args = append(args, *req.Active)
	}

	if len(setParts) == 0 {
		return s.GetSchedule(scheduleID) // No changes, return current
	}

	// Add scheduleID to args for WHERE clause
	args = append(args, scheduleID)

	query := fmt.Sprintf(`
		UPDATE task_schedules
		SET %s
		WHERE id = ?
	`, joinStrings(setParts, ", "))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("schedule not found")
	}

	return s.GetSchedule(scheduleID)
}

// DeleteSchedule deletes a task schedule
func (s *SchedulesService) DeleteSchedule(scheduleID string) error {
	query := `DELETE FROM task_schedules WHERE id = ?`

	result, err := s.db.Exec(query, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found")
	}

	return nil
}

// UpdateNextRunTime updates the next run time for a schedule
// Note: This field doesn't exist in current schema - using last_generated_date as workaround
func (s *SchedulesService) UpdateNextRunTime(scheduleID string, nextRunAt time.Time) error {
	query := `UPDATE task_schedules SET last_generated_date = ? WHERE id = ?`

	result, err := s.db.Exec(query, nextRunAt, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to update next run time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found")
	}

	return nil
}

// ActivateSchedule activates a schedule
func (s *SchedulesService) ActivateSchedule(scheduleID string) error {
	query := `UPDATE task_schedules SET active = true WHERE id = ?`

	result, err := s.db.Exec(query, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to activate schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found")
	}

	return nil
}

// DeactivateSchedule deactivates a schedule
func (s *SchedulesService) DeactivateSchedule(scheduleID string) error {
	query := `UPDATE task_schedules SET active = false WHERE id = ?`

	result, err := s.db.Exec(query, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to deactivate schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("schedule not found")
	}

	return nil
}

// Helper functions

func (s *SchedulesService) scanTaskSchedule(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.TaskSchedule, error) {
	var schedule models.TaskSchedule
	var description, assignedTo, daysOfWeek, timeOfDay sql.NullString
	var lastGeneratedDate sql.NullTime

	err := scanner.Scan(
		&schedule.ID, &schedule.FamilyID, &schedule.CreatedBy, &schedule.Title,
		&description, &schedule.TaskType, &assignedTo, &daysOfWeek,
		&timeOfDay, &schedule.Priority, &schedule.Points, &schedule.Active,
		&schedule.CreatedAt, &lastGeneratedDate,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if description.Valid {
		schedule.Description = &description.String
	}
	if assignedTo.Valid {
		schedule.AssignedTo = &assignedTo.String
	}
	if daysOfWeek.Valid {
		schedule.DaysOfWeek = &daysOfWeek.String
	}
	if timeOfDay.Valid {
		schedule.TimeOfDay = &timeOfDay.String
	}
	// Get family timezone for conversions
	familyTimezone, err := GetFamilyTimezone(s.db, schedule.FamilyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family timezone for schedule conversion: %w", err)
	}

	if lastGeneratedDate.Valid {
		// Convert LastGeneratedDate from UTC to family timezone
		convertedLastGenerated, convErr := ConvertFromUTC(lastGeneratedDate.Time, familyTimezone)
		if convErr != nil {
			return nil, fmt.Errorf("failed to convert last generated date from UTC: %w", convErr)
		}
		schedule.LastGeneratedDate = &convertedLastGenerated
	}

	// Convert CreatedAt from UTC to family timezone
	schedule.CreatedAt, err = ConvertFromUTC(schedule.CreatedAt, familyTimezone)
	if err != nil {
		return nil, fmt.Errorf("failed to convert created at from UTC: %w", err)
	}

	return &schedule, nil
}

// UpdateLastGeneratedDate updates the last generated date for a schedule
func (s *SchedulesService) UpdateLastGeneratedDate(scheduleID string, endDate time.Time) error {
	query := `
		UPDATE task_schedules
		SET last_generated_date = ?
		WHERE id = ? AND (last_generated_date IS NULL OR last_generated_date < ?)
	`

	dateStr := endDate.Format("2006-01-02 15:04:05")
	_, err := s.db.Exec(query, dateStr, scheduleID, dateStr)
	if err != nil {
		return fmt.Errorf("failed to update last generated date: %w", err)
	}

	return nil
}

// DeleteScheduleWithTasks deletes a schedule and all its tasks in a transaction
func (s *SchedulesService) DeleteScheduleWithTasks(scheduleID string) error {
	return s.db.BeginCommit(func(tx *sql.Tx) error {
		defer func() {
			_ = tx.Rollback() // nolint:errcheck
		}()

		// First, delete all tasks associated with this schedule
		deleteTasksQuery := `DELETE FROM tasks WHERE schedule_id = ?`
		_, err := tx.Exec(deleteTasksQuery, scheduleID)
		if err != nil {
			return fmt.Errorf("failed to delete tasks for schedule %s: %w", scheduleID, err)
		}

		// Then, delete the schedule itself
		deleteScheduleQuery := `DELETE FROM task_schedules WHERE id = ?`
		result, err := tx.Exec(deleteScheduleQuery, scheduleID)
		if err != nil {
			return fmt.Errorf("failed to delete schedule %s: %w", scheduleID, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get schedule rows affected: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("schedule %s not found", scheduleID)
		}

		return tx.Commit()
	})
}

// GetSchedulesNeedingGeneration returns schedules that need task generation
func (s *SchedulesService) GetSchedulesNeedingGeneration() ([]models.TaskSchedule, error) {
	query := `
		SELECT id, family_id, created_by, title, description, task_type, assigned_to,
			   days_of_week, time_of_day, priority, points, active, created_at,
			   last_generated_date
		FROM task_schedules
		WHERE active = true
		AND (
			last_generated_date IS NULL OR
			last_generated_date < date('now', '+1 month')
		)
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules needing generation: %w", err)
	}
	defer rows.Close()

	schedules := make([]models.TaskSchedule, 0)
	for rows.Next() {
		schedule, scanErr := s.scanTaskSchedule(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan schedule needing generation: %w", scanErr)
		}
		schedules = append(schedules, *schedule)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedules needing generation: %w", err)
	}

	return schedules, nil
}

func generateScheduleID() string {
	return fmt.Sprintf("schedule_%d", time.Now().UTC().UnixNano())
}
