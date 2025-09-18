package services

import (
	"database/sql"
	"fmt"
	"time"

	"famstack/internal/models"
)

// SchedulesService handles all task schedule database operations
type SchedulesService struct {
	db *sql.DB
}

// NewSchedulesService creates a new schedules service
func NewSchedulesService(db *sql.DB) *SchedulesService {
	return &SchedulesService{db: db}
}

// GetSchedule returns a task schedule by ID
func (s *SchedulesService) GetSchedule(scheduleID string) (*models.TaskSchedule, error) {
	query := `
		SELECT id, family_id, name, description, task_template, assigned_to,
			   schedule_type, cron_expression, interval_minutes, days_of_week,
			   start_date, end_date, next_run_at, is_active, created_by,
			   created_at, updated_at
		FROM task_schedules
		WHERE id = ?
	`

	var schedule models.TaskSchedule
	var description, assignedTo, cronExpression, daysOfWeek, endDate sql.NullString
	var intervalMinutes sql.NullInt64

	err := s.db.QueryRow(query, scheduleID).Scan(
		&schedule.ID, &schedule.FamilyID, &schedule.Name, &description,
		&schedule.TaskTemplate, &assignedTo, &schedule.ScheduleType,
		&cronExpression, &intervalMinutes, &daysOfWeek, &schedule.StartDate,
		&endDate, &schedule.NextRunAt, &schedule.IsActive, &schedule.CreatedBy,
		&schedule.CreatedAt, &schedule.UpdatedAt,
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
	if cronExpression.Valid {
		schedule.CronExpression = &cronExpression.String
	}
	if intervalMinutes.Valid {
		val := int(intervalMinutes.Int64)
		schedule.IntervalMinutes = &val
	}
	if daysOfWeek.Valid {
		schedule.DaysOfWeek = &daysOfWeek.String
	}
	if endDate.Valid {
		if parsed, parseErr := time.Parse(time.RFC3339, endDate.String); parseErr == nil {
			schedule.EndDate = &parsed
		}
	}

	return &schedule, nil
}

// ListSchedules returns all task schedules for a family
func (s *SchedulesService) ListSchedules(familyID string) ([]models.TaskSchedule, error) {
	query := `
		SELECT id, family_id, name, description, task_template, assigned_to,
			   schedule_type, cron_expression, interval_minutes, days_of_week,
			   start_date, end_date, next_run_at, is_active, created_by,
			   created_at, updated_at
		FROM task_schedules
		WHERE family_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []models.TaskSchedule
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
		SELECT id, family_id, name, description, task_template, assigned_to,
			   schedule_type, cron_expression, interval_minutes, days_of_week,
			   start_date, end_date, next_run_at, is_active, created_by,
			   created_at, updated_at
		FROM task_schedules
		WHERE is_active = true AND next_run_at <= CURRENT_TIMESTAMP
		ORDER BY next_run_at ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active schedules: %w", err)
	}
	defer rows.Close()

	var schedules []models.TaskSchedule
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
	now := time.Now()

	query := `
		INSERT INTO task_schedules (id, family_id, name, description, task_template,
								   assigned_to, schedule_type, cron_expression, interval_minutes,
								   days_of_week, start_date, end_date, next_run_at, is_active,
								   created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		scheduleID, familyID, req.Name, req.Description, req.TaskTemplate,
		req.AssignedTo, req.ScheduleType, req.CronExpression, req.IntervalMinutes,
		req.DaysOfWeek, req.StartDate, req.EndDate, req.NextRunAt, true,
		createdBy, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	return s.GetSchedule(scheduleID)
}

// UpdateSchedule updates an existing task schedule
func (s *SchedulesService) UpdateSchedule(scheduleID string, req *models.UpdateTaskScheduleRequest) (*models.TaskSchedule, error) {
	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
	}
	if req.TaskTemplate != nil {
		setParts = append(setParts, "task_template = ?")
		args = append(args, *req.TaskTemplate)
	}
	if req.AssignedTo != nil {
		setParts = append(setParts, "assigned_to = ?")
		args = append(args, *req.AssignedTo)
	}
	if req.ScheduleType != nil {
		setParts = append(setParts, "schedule_type = ?")
		args = append(args, *req.ScheduleType)
	}
	if req.CronExpression != nil {
		setParts = append(setParts, "cron_expression = ?")
		args = append(args, *req.CronExpression)
	}
	if req.IntervalMinutes != nil {
		setParts = append(setParts, "interval_minutes = ?")
		args = append(args, *req.IntervalMinutes)
	}
	if req.DaysOfWeek != nil {
		setParts = append(setParts, "days_of_week = ?")
		args = append(args, *req.DaysOfWeek)
	}
	if req.StartDate != nil {
		setParts = append(setParts, "start_date = ?")
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil {
		setParts = append(setParts, "end_date = ?")
		args = append(args, *req.EndDate)
	}
	if req.NextRunAt != nil {
		setParts = append(setParts, "next_run_at = ?")
		args = append(args, *req.NextRunAt)
	}
	if req.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	if len(setParts) == 1 { // Only updated_at
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
func (s *SchedulesService) UpdateNextRunTime(scheduleID string, nextRunAt time.Time) error {
	query := `UPDATE task_schedules SET next_run_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

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
	query := `UPDATE task_schedules SET is_active = true, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

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
	query := `UPDATE task_schedules SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

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
	var description, assignedTo, cronExpression, daysOfWeek, endDate sql.NullString
	var intervalMinutes sql.NullInt64

	err := scanner.Scan(
		&schedule.ID, &schedule.FamilyID, &schedule.Name, &description,
		&schedule.TaskTemplate, &assignedTo, &schedule.ScheduleType,
		&cronExpression, &intervalMinutes, &daysOfWeek, &schedule.StartDate,
		&endDate, &schedule.NextRunAt, &schedule.IsActive, &schedule.CreatedBy,
		&schedule.CreatedAt, &schedule.UpdatedAt,
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
	if cronExpression.Valid {
		schedule.CronExpression = &cronExpression.String
	}
	if intervalMinutes.Valid {
		val := int(intervalMinutes.Int64)
		schedule.IntervalMinutes = &val
	}
	if daysOfWeek.Valid {
		schedule.DaysOfWeek = &daysOfWeek.String
	}
	if endDate.Valid {
		if parsed, parseErr := time.Parse(time.RFC3339, endDate.String); parseErr == nil {
			schedule.EndDate = &parsed
		}
	}

	return &schedule, nil
}

func generateScheduleID() string {
	return fmt.Sprintf("schedule_%d", time.Now().UnixNano())
}
