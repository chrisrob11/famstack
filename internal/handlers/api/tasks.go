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

// TaskColumn represents a column of tasks for a family member
type TaskColumn struct {
	Member Member        `json:"member"`
	Tasks  []models.Task `json:"tasks"`
}

// Member represents a family member in the API
type Member struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	MemberType string `json:"member_type"`
}

// TasksResponse represents the response for listing tasks
type TasksResponse struct {
	TasksByMember map[string]TaskColumn `json:"tasks_by_member"`
	Date          string                `json:"date"`
}

// ListTasks returns all tasks as JSON
func (h *TaskAPIHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get date parameter from query string, default to today
	dateParam := r.URL.Query().Get("date")
	var dateFilter string
	if dateParam != "" {
		// Use provided date (expected in YYYY-MM-DD format)
		dateFilter = dateParam
	} else {
		// Default to today
		dateFilter = time.Now().Format("2006-01-02")
	}

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

	rows, queryErr := h.db.Query(query, "fam1", dateFilter, "fam1")
	if queryErr != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", queryErr), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasksByMember := make(map[string]TaskColumn)

	for rows.Next() {
		var memberID, memberName, memberType string
		var taskID, title, description, status, taskType, familyID *string
		var assignedTo *string
		var dueDate, createdAt, completedAt *time.Time

		err := rows.Scan(
			&memberID, &memberName, &memberType,
			&taskID, &title, &description, &status, &taskType,
			&assignedTo, &familyID, &dueDate, &createdAt, &completedAt,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}

		memberKey := fmt.Sprintf("member_%s", memberID)

		// Initialize member column if not exists
		if _, exists := tasksByMember[memberKey]; !exists {
			tasksByMember[memberKey] = TaskColumn{
				Member: Member{
					ID:         memberID,
					Name:       memberName,
					MemberType: memberType,
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

			column := tasksByMember[memberKey]
			column.Tasks = append(column.Tasks, task)
			tasksByMember[memberKey] = column
		}
	}

	// Get unassigned tasks separately (tasks with NULL or empty assigned_to)
	unassignedQuery := `
		SELECT
			id, title, description, status, task_type, assigned_to, family_id,
			due_date, created_at, completed_at
		FROM tasks
		WHERE family_id = ? AND (assigned_to IS NULL OR assigned_to = '') AND DATE(created_at) = ?
		ORDER BY created_at DESC
	`

	unassignedRows, err := h.db.Query(unassignedQuery, "fam1", dateFilter)
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
	tasksByMember["unassigned"] = TaskColumn{
		Member: Member{
			ID:         "unassigned",
			Name:       "Unassigned",
			MemberType: "system",
		},
		Tasks: unassignedTasks,
	}

	response := TasksResponse{
		TasksByMember: tasksByMember,
		Date:          time.Now().Format("Monday, January 2"),
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

	// Get the authenticated family member from the request context
	// For now, we'll use a valid family member ID from our database
	// In a real implementation, this would come from the auth middleware
	var validMemberID string
	err := h.db.QueryRow("SELECT id FROM family_members WHERE password_hash IS NOT NULL LIMIT 1").Scan(&validMemberID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]any{
			"error":   "Authentication error",
			"details": "Could not determine current user",
		}); encErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}
	task.CreatedBy = validMemberID

	// assigned_to field can now directly use family member IDs - no conversion needed!

	// Validate that due_date is not in the past
	if task.DueDate != nil {
		today := time.Now().Truncate(24 * time.Hour)
		dueDate := task.DueDate.Truncate(24 * time.Hour)
		if dueDate.Before(today) {
			w.WriteHeader(http.StatusBadRequest)
			if encErr := json.NewEncoder(w).Encode(map[string]any{
				"error":   "Validation failed",
				"details": "Cannot create tasks for past dates",
			}); encErr != nil {
				http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
			}
			return
		}
	}

	// Prepare for creation (sanitize, set defaults, validate)
	if prepareErr := task.PrepareForCreate(); prepareErr != nil {
		if validationErrs, ok := prepareErr.(validation.ValidationErrors); ok {
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
		http.Error(w, fmt.Sprintf("Validation failed: %v", prepareErr), http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO tasks (title, description, status, task_type, assigned_to, family_id, created_by, priority, created_at, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	var newID string
	err = h.db.QueryRow(query, task.Title, task.Description, task.Status, task.TaskType, task.AssignedTo, task.FamilyID, task.CreatedBy, task.Priority, task.CreatedAt, task.DueDate).Scan(&newID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if encErr := json.NewEncoder(w).Encode(map[string]any{
			"error":   "Database error",
			"details": fmt.Sprintf("Failed to create task: %v", err),
		}); encErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}
	task.ID = newID

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
