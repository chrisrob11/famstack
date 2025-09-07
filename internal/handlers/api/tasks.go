package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
	"famstack/internal/validation"
)

// TaskAPIHandler handles JSON API requests for tasks
type TaskAPIHandler struct {
	db *database.DB
}

// NewTaskAPIHandler creates a new task API handler
func NewTaskAPIHandler(db *database.DB) *TaskAPIHandler {
	return &TaskAPIHandler{db: db}
}

// TaskColumn represents a column of tasks for a user
type TaskColumn struct {
	User  User          `json:"user"`
	Tasks []models.Task `json:"tasks"`
}

// User represents a user in the API
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// TasksResponse represents the response for listing tasks
type TasksResponse struct {
	TasksByUser map[string]TaskColumn `json:"tasks_by_user"`
	Date        string                `json:"date"`
}

// ListTasks returns all tasks as JSON
func (h *TaskAPIHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Single JOIN query to get all users and their tasks in one go
	query := `
		SELECT 
			u.id as user_id, u.name as user_name, u.role as user_role,
			t.id as task_id, t.title, t.description, t.status, t.task_type, 
			t.assigned_to, t.family_id, t.due_date, t.created_at, t.completed_at
		FROM users u
		LEFT JOIN tasks t ON (u.id = t.assigned_to AND t.family_id = ? AND DATE(t.created_at) = DATE('now'))
		WHERE u.family_id = ?
		ORDER BY u.name, t.created_at DESC
	`

	rows, queryErr := h.db.Query(query, "fam1", "fam1")
	if queryErr != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", queryErr), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasksByUser := make(map[string]TaskColumn)

	for rows.Next() {
		var userID, userName, userRole string
		var taskID, title, description, status, taskType, familyID *string
		var assignedTo *string
		var dueDate, createdAt, completedAt *time.Time

		err := rows.Scan(
			&userID, &userName, &userRole,
			&taskID, &title, &description, &status, &taskType,
			&assignedTo, &familyID, &dueDate, &createdAt, &completedAt,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}

		userKey := fmt.Sprintf("user_%s", userID)

		// Initialize user column if not exists
		if _, exists := tasksByUser[userKey]; !exists {
			tasksByUser[userKey] = TaskColumn{
				User: User{
					ID:   userID,
					Name: userName,
					Role: userRole,
				},
				Tasks: []models.Task{},
			}
		}

		// Add task if it exists (LEFT JOIN may return NULL task fields)
		if taskID != nil && *taskID != "" {
			task := models.Task{
				ID:          *taskID,
				Title:       *title,
				Description: *description,
				Status:      *status,
				TaskType:    *taskType,
				AssignedTo:  assignedTo,
				FamilyID:    *familyID,
				DueDate:     dueDate,
				CreatedAt:   *createdAt,
				CompletedAt: completedAt,
			}

			column := tasksByUser[userKey]
			column.Tasks = append(column.Tasks, task)
			tasksByUser[userKey] = column
		}
	}

	// Get unassigned tasks separately (tasks with NULL or empty assigned_to)
	unassignedQuery := `
		SELECT 
			id, title, description, status, task_type, assigned_to, family_id,
			due_date, created_at, completed_at
		FROM tasks 
		WHERE family_id = ? AND (assigned_to IS NULL OR assigned_to = '') AND DATE(created_at) = DATE('now')
		ORDER BY created_at DESC
	`

	unassignedRows, err := h.db.Query(unassignedQuery, "fam1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Unassigned tasks query error: %v", err), http.StatusInternalServerError)
		return
	}
	defer unassignedRows.Close()

	var unassignedTasks []models.Task
	for unassignedRows.Next() {
		var task models.Task
		var assignedTo *string
		var dueDate, completedAt *time.Time

		err := unassignedRows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status, &task.TaskType,
			&assignedTo, &task.FamilyID, &dueDate, &task.CreatedAt, &completedAt,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unassigned task scan error: %v", err), http.StatusInternalServerError)
			return
		}

		task.AssignedTo = assignedTo
		task.DueDate = dueDate
		task.CompletedAt = completedAt
		unassignedTasks = append(unassignedTasks, task)
	}

	// Add unassigned column
	tasksByUser["unassigned"] = TaskColumn{
		User: User{
			ID:   "unassigned",
			Name: "Unassigned",
			Role: "system",
		},
		Tasks: unassignedTasks,
	}

	response := TasksResponse{
		TasksByUser: tasksByUser,
		Date:        time.Now().Format("Monday, January 2"),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateTask creates a new task
func (h *TaskAPIHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set created by (would come from auth in real app)
	task.CreatedBy = "user1" // Default to user1 for now

	// Prepare for creation (sanitize, set defaults, validate)
	if err := task.PrepareForCreate(); err != nil {
		if validationErrs, ok := err.(validation.ValidationErrors); ok {
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]any{
				"error":   "Validation failed",
				"details": validationErrs,
			}); encErr != nil {
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
				return
			}
			return
		}
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO tasks (title, description, status, task_type, assigned_to, family_id, created_by, priority, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := h.db.Exec(query, task.Title, task.Description, task.Status, task.TaskType, task.AssignedTo, task.FamilyID, task.CreatedBy, task.Priority, task.CreatedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get task ID", http.StatusInternalServerError)
		return
	}
	task.ID = fmt.Sprintf("%d", id)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
}

// UpdateTask updates a task (complete/reopen)
func (h *TaskAPIHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Extract task ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskID := pathParts[4]
	if taskID == "" {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var updateData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Build dynamic UPDATE query for all fields at once
	var setParts []string
	var args []any

	// Handle status updates
	if status, exists := updateData["status"]; exists {
		switch status {
		case "completed":
			setParts = append(setParts, "status = ?", "completed_at = ?")
			args = append(args, "completed", time.Now())
		case "pending":
			setParts = append(setParts, "status = ?", "completed_at = NULL")
			args = append(args, "pending")
		default:
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
	}

	// Handle assignment updates
	if assignedTo, exists := updateData["assigned_to"]; exists {
		if assignedTo == nil || assignedTo == "" {
			setParts = append(setParts, "assigned_to = NULL")
		} else {
			setParts = append(setParts, "assigned_to = ?")
			args = append(args, assignedTo)
		}
	}

	// Handle title updates
	if title, exists := updateData["title"]; exists {
		titleStr, ok := title.(string)
		if !ok {
			http.Error(w, "Invalid title format", http.StatusBadRequest)
			return
		}
		if titleStr == "" {
			http.Error(w, "Title cannot be empty", http.StatusBadRequest)
			return
		}
		setParts = append(setParts, "title = ?")
		args = append(args, titleStr)
	}

	// If no valid updates, return error
	if len(setParts) == 0 {
		http.Error(w, "No valid updates provided", http.StatusBadRequest)
		return
	}

	// Build and execute single UPDATE query
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = ? AND family_id = ?", strings.Join(setParts, ", "))
	args = append(args, taskID, "fam1")

	result, err := h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check affected rows", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Return updated task using SELECT with RETURNING-like behavior
	query = `
		SELECT 
			id, title, description, status, task_type, 
			assigned_to, family_id, due_date, created_at, completed_at
		FROM tasks 
		WHERE id = ? AND family_id = ?
	`

	var task models.Task
	var dueDate, completedAt *time.Time

	err = h.db.QueryRow(query, taskID, "fam1").Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.TaskType,
		&task.AssignedTo, &task.FamilyID, &dueDate, &task.CreatedAt, &completedAt,
	)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	task.DueDate = dueDate
	task.CompletedAt = completedAt

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
}

// DeleteTask deletes a task
func (h *TaskAPIHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	taskID := pathParts[4]
	if taskID == "" {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM tasks WHERE id = ? AND family_id = ?`
	result, err := h.db.Exec(query, taskID, "fam1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check affected rows", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTask retrieves a single task
func (h *TaskAPIHandler) GetTask(w http.ResponseWriter, r *http.Request, taskID string) {
	query := `
		SELECT 
			t.id, t.title, t.description, t.status, t.task_type, 
			t.assigned_to, t.family_id, t.due_date, t.created_at, t.completed_at
		FROM tasks t 
		WHERE t.id = ? AND t.family_id = ?
	`

	var task models.Task
	var dueDate, completedAt *time.Time

	err := h.db.QueryRow(query, taskID, "fam1").Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.TaskType,
		&task.AssignedTo, &task.FamilyID, &dueDate, &task.CreatedAt, &completedAt,
	)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Assign dates directly (they're already *time.Time)
	task.DueDate = dueDate
	task.CompletedAt = completedAt

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
}
