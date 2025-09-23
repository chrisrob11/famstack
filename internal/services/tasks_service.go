package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
)

// TasksService handles all task database operations
type TasksService struct {
	db *database.Fascade
}

// NewTasksService creates a new tasks service
func NewTasksService(db *database.Fascade) *TasksService {
	return &TasksService{db: db}
}

// TaskColumn represents a column of tasks for a family member
type TaskColumn struct {
	Member TaskMember    `json:"member"`
	Tasks  []models.Task `json:"tasks"`
}

// TaskMember represents a family member in the task context
type TaskMember struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MemberType string `json:"member_type"`
}

// TasksResponse represents the response for listing tasks
type TasksResponse struct {
	TasksByMember map[string]TaskColumn `json:"tasks_by_member"`
	Date          string                `json:"date"`
}

// ListTasksByFamily returns all tasks organized by family member for a specific date
func (s *TasksService) ListTasksByFamily(familyID, dateFilter string) (*TasksResponse, error) {
	// 1. Get all active family members
	members, err := s.getActiveFamilyMembers(familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family members: %w", err)
	}

	// 2. Get all tasks for the family for the given date
	tasks, err := s.getTasksForFamily(familyID, dateFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks for family: %w", err)
	}

	// 3. Organize tasks by member
	tasksByMember := make(map[string]TaskColumn)
	for _, member := range members {
		tasksByMember[member.ID] = TaskColumn{
			Member: member,
			Tasks:  []models.Task{},
		}
	}

	// Create a bucket for unassigned tasks
	tasksByMember["unassigned"] = TaskColumn{
		Member: TaskMember{
			ID:         "unassigned",
			Name:       "Unassigned",
			MemberType: "system",
		},
		Tasks: []models.Task{},
	}

	for _, task := range tasks {
		assigneeID := "unassigned"
		if task.AssignedTo != nil && *task.AssignedTo != "" {
			assigneeID = *task.AssignedTo
		}

		if column, ok := tasksByMember[assigneeID]; ok {
			column.Tasks = append(column.Tasks, task)
			tasksByMember[assigneeID] = column
		}
		// If a task is assigned to a member who is no longer active,
		// it will be ignored. This is the desired behavior.
	}

	return &TasksResponse{
		TasksByMember: tasksByMember,
		Date:          dateFilter,
	}, nil
}

// getActiveFamilyMembers retrieves all active members for a given family
func (s *TasksService) getActiveFamilyMembers(familyID string) ([]TaskMember, error) {
	query := `
		SELECT id, first_name, last_name, member_type
		FROM family_members
		WHERE family_id = ? AND is_active = TRUE
		ORDER BY first_name, last_name
	`
	rows, err := s.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active family members: %w", err)
	}
	defer rows.Close()

	var members []TaskMember
	for rows.Next() {
		var member TaskMember
		var firstName, lastName string
		if err := rows.Scan(&member.ID, &firstName, &lastName, &member.MemberType); err != nil {
			return nil, fmt.Errorf("failed to scan family member: %w", err)
		}
		member.Name = firstName + " " + lastName
		members = append(members, member)
	}

	return members, rows.Err()
}

// getTasksForFamily retrieves all tasks for a family on a specific date
func (s *TasksService) getTasksForFamily(familyID, dateFilter string) ([]models.Task, error) {
	query := `
		SELECT id, family_id, assigned_to, title, description, task_type, status,
			   priority, due_date, created_by, created_at, completed_at
		FROM tasks
		WHERE family_id = ? AND SUBSTR(due_date, 1, 10) = ?
		ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, familyID, dateFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks for family: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task, scanErr := s.scanTask(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan task: %w", scanErr)
		}
		tasks = append(tasks, *task)
	}

	return tasks, rows.Err()
}

// GetTask returns a specific task by ID
func (s *TasksService) GetTask(taskID string) (*models.Task, error) {
	query := `
		SELECT id, family_id, assigned_to, title, description, task_type, status,
			   priority, due_date, created_by, created_at, completed_at
		FROM tasks
		WHERE id = ?
	`

	var task models.Task
	var assignedTo, dueDate, completedAt sql.NullString

	err := s.db.QueryRow(query, taskID).Scan(
		&task.ID, &task.FamilyID, &assignedTo, &task.Title, &task.Description,
		&task.TaskType, &task.Status, &task.Priority, &dueDate,
		&task.CreatedBy, &task.CreatedAt, &completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Handle nullable fields
	if assignedTo.Valid {
		task.AssignedTo = &assignedTo.String
	}
	if dueDate.Valid {
		if parsed, parseErr := time.Parse(time.RFC3339, dueDate.String); parseErr == nil {
			task.DueDate = &parsed
		}
	}
	if completedAt.Valid {
		if parsed, parseErr := time.Parse(time.RFC3339, completedAt.String); parseErr == nil {
			task.CompletedAt = &parsed
		}
	}

	return &task, nil
}

// CreateTask creates a new task
func (s *TasksService) CreateTask(familyID, createdBy string, req *models.CreateTaskRequest) (*models.Task, error) {
	taskID := generateTaskID()
	now := time.Now().UTC()

	query := `
		INSERT INTO tasks (id, family_id, assigned_to, title, description, task_type,
						  status, priority, due_date, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		taskID, familyID, req.AssignedTo, req.Title, req.Description,
		req.TaskType, "pending", req.Priority, req.DueDate,
		createdBy, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return s.GetTask(taskID)
}

// UpdateTask updates an existing task
func (s *TasksService) UpdateTask(taskID string, req *models.UpdateTaskRequest) (*models.Task, error) {
	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []any{}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *req.Status)

		// Set completed_at when marking as completed
		if *req.Status == "completed" {
			setParts = append(setParts, "completed_at = CURRENT_TIMESTAMP")
		} else {
			setParts = append(setParts, "completed_at = NULL")
		}
	}
	if req.AssignedTo != nil {
		setParts = append(setParts, "assigned_to = ?")
		args = append(args, *req.AssignedTo)
	}
	if req.Priority != nil {
		setParts = append(setParts, "priority = ?")
		args = append(args, *req.Priority)
	}
	if req.DueDate != nil {
		setParts = append(setParts, "due_date = ?")
		args = append(args, *req.DueDate)
	}

	if len(setParts) == 1 { // Only updated_at
		return s.GetTask(taskID) // No changes, return current
	}

	// Add taskID to args for WHERE clause
	args = append(args, taskID)

	query := fmt.Sprintf(`
		UPDATE tasks
		SET %s
		WHERE id = ?
	`, strings.Join(setParts, ", "))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("task not found")
	}

	return s.GetTask(taskID)
}

// DeleteTask deletes a task
func (s *TasksService) DeleteTask(taskID string) error {
	query := `DELETE FROM tasks WHERE id = ?`

	result, err := s.db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// ListTasksByMember returns all tasks assigned to a specific family member
func (s *TasksService) ListTasksByMember(memberID string) ([]models.Task, error) {
	query := `
		SELECT id, family_id, assigned_to, title, description, task_type, status,
			   priority, due_date, created_by, created_at, completed_at
		FROM tasks
		WHERE assigned_to = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by member: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task, scanErr := s.scanTask(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan task: %w", scanErr)
		}
		tasks = append(tasks, *task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

// ListTasksForFamily returns all tasks for a family
func (s *TasksService) ListTasksForFamily(familyID string) ([]models.Task, error) {
	query := `
		SELECT id, family_id, assigned_to, title, description, task_type, status,
			   priority, due_date, created_by, created_at, completed_at
		FROM tasks
		WHERE family_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query family tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		task, scanErr := s.scanTask(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan task: %w", scanErr)
		}
		tasks = append(tasks, *task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

// Helper functions

func (s *TasksService) scanTask(scanner interface {
	Scan(dest ...any) error
}) (*models.Task, error) {
	var task models.Task
	var assignedTo, dueDate, completedAt sql.NullString

	err := scanner.Scan(
		&task.ID, &task.FamilyID, &assignedTo, &task.Title, &task.Description,
		&task.TaskType, &task.Status, &task.Priority, &dueDate,
		&task.CreatedBy, &task.CreatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if assignedTo.Valid {
		task.AssignedTo = &assignedTo.String
	}
	if dueDate.Valid {
		if parsed, parseErr := time.Parse(time.RFC3339, dueDate.String); parseErr == nil {
			task.DueDate = &parsed
		}
	}
	if completedAt.Valid {
		if parsed, parseErr := time.Parse(time.RFC3339, completedAt.String); parseErr == nil {
			task.CompletedAt = &parsed
		}
	}

	return &task, nil
}

// GetExistingTasksInRange retrieves existing task dates in a date range for a schedule
func (s *TasksService) GetExistingTasksInRange(scheduleID string, startDate, endDate time.Time) ([]string, error) {
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

	rows, err := s.db.Query(query, scheduleID,
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

// BulkCreateTasks creates multiple tasks in a single transaction
func (s *TasksService) BulkCreateTasks(familyID, createdBy string, tasks []BulkTaskRequest) error {
	if len(tasks) == 0 {
		return nil
	}

	return s.db.BeginCommit(func(tx *sql.Tx) error {
		defer func() {
			_ = tx.Rollback() // nolint:errcheck
		}()

		query := `
			INSERT INTO tasks (id, family_id, assigned_to, title, description, task_type,
							  status, priority, due_date, created_by, schedule_id)
			VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?)
		`

		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, task := range tasks {
			taskID := generateTaskID()
			var assignedToValue any
			if task.AssignedTo != nil {
				assignedToValue = *task.AssignedTo
			} else {
				assignedToValue = nil
			}

			var dueDateValue any
			if task.DueDate != nil {
				dueDateValue = task.DueDate.Format("2006-01-02 15:04:05")
			} else {
				dueDateValue = nil
			}

			_, err = stmt.Exec(
				taskID, familyID, assignedToValue, task.Title, task.Description,
				task.TaskType, task.Priority, dueDateValue,
				createdBy, task.ScheduleID,
			)
			if err != nil {
				if isUniqueConstraintViolation(err) {
					// Task already exists for this schedule on this date - skip and continue
					continue
				}
				return fmt.Errorf("failed to insert task: %w", err)
			}
		}

		return tx.Commit()
	})
}

// DeleteTasksBySchedule deletes all tasks for a given schedule
func (s *TasksService) DeleteTasksBySchedule(scheduleID string) (int64, error) {
	query := `DELETE FROM tasks WHERE schedule_id = ?`
	result, err := s.db.Exec(query, scheduleID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete tasks for schedule %s: %w", scheduleID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// BulkTaskRequest represents a task to be created in bulk
type BulkTaskRequest struct {
	Title       string
	Description string
	TaskType    string
	AssignedTo  *string
	Priority    int
	Points      int
	DueDate     *time.Time
	ScheduleID  string
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

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UTC().UnixNano())
}
