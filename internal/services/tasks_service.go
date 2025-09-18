package services

import (
	"database/sql"
	"fmt"
	"time"

	"famstack/internal/models"
)

// TasksService handles all task database operations
type TasksService struct {
	db *sql.DB
}

// NewTasksService creates a new tasks service
func NewTasksService(db *sql.DB) *TasksService {
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
	// Single JOIN query to get all family members and their tasks in one go
	query := `
		SELECT
			fm.id as member_id, fm.name as member_name, fm.member_type as member_type,
			t.id as task_id, t.title, t.description, t.status, t.task_type,
			t.assigned_to, t.family_id, t.due_date, t.created_at, t.completed_at
		FROM family_members fm
		LEFT JOIN tasks t ON (fm.id = t.assigned_to AND t.family_id = ? AND DATE(t.created_at) = ?)
		WHERE fm.family_id = ? AND fm.is_active = TRUE
		ORDER BY fm.name, t.created_at DESC
	`

	rows, err := s.db.Query(query, familyID, dateFilter, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by family: %w", err)
	}
	defer rows.Close()

	tasksByMember := make(map[string]TaskColumn)

	for rows.Next() {
		var memberID, memberName, memberType string
		var taskID, title, description, status, taskType, familyID *string
		var assignedTo *string
		var dueDate, createdAt, completedAt *time.Time

		if scanErr := rows.Scan(
			&memberID, &memberName, &memberType,
			&taskID, &title, &description, &status, &taskType,
			&assignedTo, &familyID, &dueDate, &createdAt, &completedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan task row: %w", scanErr)
		}

		// Initialize member column if not exists
		if _, exists := tasksByMember[memberID]; !exists {
			tasksByMember[memberID] = TaskColumn{
				Member: TaskMember{
					ID:         memberID,
					Name:       memberName,
					MemberType: memberType,
				},
				Tasks: []models.Task{},
			}
		}

		// Add task if it exists (not null from LEFT JOIN)
		if taskID != nil {
			task := models.Task{
				ID:          *taskID,
				FamilyID:    *familyID,
				AssignedTo:  assignedTo,
				Title:       *title,
				Description: *description,
				Status:      *status,
				TaskType:    *taskType,
				DueDate:     dueDate,
				CreatedAt:   *createdAt,
				CompletedAt: completedAt,
			}

			column := tasksByMember[memberID]
			column.Tasks = append(column.Tasks, task)
			tasksByMember[memberID] = column
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return &TasksResponse{
		TasksByMember: tasksByMember,
		Date:          dateFilter,
	}, nil
}

// GetTask returns a specific task by ID
func (s *TasksService) GetTask(taskID string) (*models.Task, error) {
	query := `
		SELECT id, family_id, assigned_to, title, description, task_type, status,
			   priority, due_date, points, created_by, created_at, completed_at
		FROM tasks
		WHERE id = ?
	`

	var task models.Task
	var assignedTo, dueDate, completedAt sql.NullString
	var points sql.NullInt64

	err := s.db.QueryRow(query, taskID).Scan(
		&task.ID, &task.FamilyID, &assignedTo, &task.Title, &task.Description,
		&task.TaskType, &task.Status, &task.Priority, &dueDate, &points,
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
	now := time.Now()

	query := `
		INSERT INTO tasks (id, family_id, assigned_to, title, description, task_type,
						  status, priority, due_date, points, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		taskID, familyID, req.AssignedTo, req.Title, req.Description,
		req.TaskType, "pending", req.Priority, req.DueDate, req.Points,
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
	args := []interface{}{}

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
	`, joinStrings(setParts, ", "))

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
			   priority, due_date, points, created_by, created_at, completed_at
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
			   priority, due_date, points, created_by, created_at, completed_at
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
	Scan(dest ...interface{}) error
}) (*models.Task, error) {
	var task models.Task
	var assignedTo, dueDate, completedAt sql.NullString
	var points sql.NullInt64

	err := scanner.Scan(
		&task.ID, &task.FamilyID, &assignedTo, &task.Title, &task.Description,
		&task.TaskType, &task.Status, &task.Priority, &dueDate, &points,
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

func generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}
