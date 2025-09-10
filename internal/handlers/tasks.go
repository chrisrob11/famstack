package handlers

import (
	"html"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
	"famstack/internal/validation"
)

type TaskHandler struct {
	db *database.DB
}

func NewTaskHandler(db *database.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

type TaskListData struct {
	Tasks       []TaskWithUser
	TasksByUser map[string]UserColumn
	Date        string
}

type UserColumn struct {
	User  models.User
	Tasks []TaskWithUser
}

type TaskWithUser struct {
	models.Task
	AssignedToName string
	CreatedByName  string
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	// Simple data for the TypeScript component page
	// The component will fetch actual task data via JSON API
	data := struct {
		CSRFToken string
	}{
		CSRFToken: generateCSRFToken(), // Simple CSRF token for now
	}

	// Load template from file
	tmplPath := filepath.Join("web", "templates", "tasks.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// generateCSRFToken creates a simple CSRF token
// In production, you'd want a more secure implementation
func generateCSRFToken() string {
	return "csrf-token-placeholder" // TODO: Implement proper CSRF token generation
}

// NewTaskForm returns an HTMX form for creating a new task
func (h *TaskHandler) NewTaskForm(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// Get user name for validation
	var userName string
	if userID == "unassigned" {
		userName = "Unassigned"
	} else {
		row := h.db.QueryRow("SELECT name FROM users WHERE id = ?", userID)
		if err := row.Scan(&userName); err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	}

	// Load template
	tmplPath := filepath.Join("web", "templates", "task-form.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Template data
	data := struct {
		UserID string
	}{
		UserID: userID,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Validate and sanitize title
	title := r.FormValue("title")
	if err := validation.ValidateTitle(title); err != nil {
		http.Error(w, "Invalid title: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate assigned user
	assignedTo := r.FormValue("assigned_to")
	if err := validation.ValidateUserID(assignedTo); err != nil {
		http.Error(w, "Invalid user ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Verify user exists in database (unless unassigned)
	if assignedTo != "" && assignedTo != "unassigned" {
		var exists bool
		err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND family_id = ?)", assignedTo, "fam1").Scan(&exists)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Assigned user not found", http.StatusBadRequest)
			return
		}
	}

	if assignedTo == "unassigned" {
		assignedTo = ""
	}

	// Create the task
	familyID := "fam1"   // For now, hardcoded
	createdBy := "user1" // For now, hardcoded (would come from session)

	query := `
		INSERT INTO tasks (family_id, assigned_to, title, task_type, status, priority, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	var assignedToPtr *string
	if assignedTo != "" {
		assignedToPtr = &assignedTo
	}

	_, err := h.db.Exec(query, familyID, assignedToPtr, title, "todo", "pending", 0, createdBy, time.Now())
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Get the ID of the newly created task
	var newTaskID string
	err = h.db.QueryRow("SELECT id FROM tasks WHERE family_id = ? AND title = ? AND created_by = ? ORDER BY created_at DESC LIMIT 1",
		familyID, title, createdBy).Scan(&newTaskID)
	if err != nil {
		http.Error(w, "Failed to retrieve task ID", http.StatusInternalServerError)
		return
	}

	// Return the new task HTML
	safeTitle := html.EscapeString(title)
	safeTaskID := html.EscapeString(newTaskID)
	taskHTML := `
	<div class="task-item pending todo" data-task-id="` + safeTaskID + `" draggable="true">
		<div class="task-title">` + safeTitle + `</div>
		<div class="task-meta">
			<span class="task-status status-pending">pending</span>
			<span class="task-type type-todo">todo</span>
		</div>
		<div class="task-actions">
			<button class="task-complete-btn" 
					hx-patch="/api/tasks/` + safeTaskID + `/complete"
					hx-target="closest .task-item"
					hx-swap="outerHTML">
				✓ Complete
			</button>
			<details class="task-actions-dropdown">
				<summary class="task-actions-toggle">⋯</summary>
				<div class="actions-menu">
					<button class="task-delete-btn" 
							hx-delete="/api/tasks/` + safeTaskID + `"
							hx-target="closest .task-item"
							hx-swap="outerHTML"
							hx-confirm="⚠️ PERMANENTLY DELETE this task? This cannot be undone!">
						🗑️ Delete Forever
					</button>
				</div>
			</details>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(taskHTML)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// CancelTaskForm returns the original Add Task button
func (h *TaskHandler) CancelTaskForm(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	// Load template
	tmplPath := filepath.Join("web", "templates", "add-task-button.html.tmpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Template data
	data := struct {
		UserID string
	}{
		UserID: userID,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// CompleteTask handles PATCH /api/tasks/{id}/complete
func (h *TaskHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskID := pathParts[3] // /api/tasks/{id}/complete

	// Validate task ID
	if err := validation.ValidateUserID(taskID); err != nil {
		http.Error(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	// Update task status in database
	query := `UPDATE tasks SET status = 'completed', completed_at = ? WHERE id = ? AND family_id = ?`
	result, err := h.db.Exec(query, time.Now(), taskID, "fam1")
	if err != nil {
		http.Error(w, "Failed to complete task", http.StatusInternalServerError)
		return
	}

	// Check if task was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get task details to return updated HTML
	var title, taskType string
	err = h.db.QueryRow("SELECT title, task_type FROM tasks WHERE id = ? AND family_id = ?", taskID, "fam1").Scan(&title, &taskType)
	if err != nil {
		http.Error(w, "Failed to retrieve task details", http.StatusInternalServerError)
		return
	}

	// Return updated task HTML with completed styling
	safeTitle := html.EscapeString(title)
	safeTaskID := html.EscapeString(taskID)
	taskHTML := `
	<div class="task-item completed ` + taskType + `" data-task-id="` + safeTaskID + `" draggable="true">
		<div class="task-title">` + safeTitle + `</div>
		<div class="task-meta">
			<span class="task-status status-completed">completed</span>
			<span class="task-type type-` + taskType + `">` + taskType + `</span>
		</div>
		<div class="task-actions">
			<details class="task-actions-dropdown">
				<summary class="task-actions-toggle">⋯</summary>
				<div class="actions-menu">
					<button class="task-reopen-btn" 
							hx-patch="/api/tasks/` + safeTaskID + `/reopen"
							hx-target="closest .task-item"
							hx-swap="outerHTML">
						↺ Reopen Task
					</button>
					<button class="task-delete-btn" 
							hx-delete="/api/tasks/` + safeTaskID + `"
							hx-target="closest .task-item"
							hx-swap="outerHTML"
							hx-confirm="⚠️ PERMANENTLY DELETE this task? This cannot be undone!">
						🗑️ Delete Forever
					</button>
				</div>
			</details>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(taskHTML)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// ReopenTask handles PATCH /api/tasks/{id}/reopen
func (h *TaskHandler) ReopenTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract task ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskID := pathParts[3] // /api/tasks/{id}/reopen

	// Validate task ID
	if err := validation.ValidateUserID(taskID); err != nil {
		http.Error(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	// Update task status back to pending
	query := `UPDATE tasks SET status = 'pending', completed_at = NULL WHERE id = ? AND family_id = ?`
	result, err := h.db.Exec(query, taskID, "fam1")
	if err != nil {
		http.Error(w, "Failed to reopen task", http.StatusInternalServerError)
		return
	}

	// Check if task was actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get task details to return updated HTML
	var title, taskType string
	err = h.db.QueryRow("SELECT title, task_type FROM tasks WHERE id = ? AND family_id = ?", taskID, "fam1").Scan(&title, &taskType)
	if err != nil {
		http.Error(w, "Failed to retrieve task details", http.StatusInternalServerError)
		return
	}

	// Return updated task HTML with pending styling
	safeTitle := html.EscapeString(title)
	safeTaskID := html.EscapeString(taskID)
	taskHTML := `
	<div class="task-item pending ` + taskType + `" data-task-id="` + safeTaskID + `" draggable="true">
		<div class="task-title">` + safeTitle + `</div>
		<div class="task-meta">
			<span class="task-status status-pending">pending</span>
			<span class="task-type type-` + taskType + `">` + taskType + `</span>
		</div>
		<div class="task-actions">
			<button class="task-complete-btn" 
					hx-patch="/api/tasks/` + safeTaskID + `/complete"
					hx-target="closest .task-item"
					hx-swap="outerHTML">
				✓ Complete
			</button>
			<details class="task-actions-dropdown">
				<summary class="task-actions-toggle">⋯</summary>
				<div class="actions-menu">
					<button class="task-delete-btn" 
							hx-delete="/api/tasks/` + safeTaskID + `"
							hx-target="closest .task-item"
							hx-swap="outerHTML"
							hx-confirm="⚠️ PERMANENTLY DELETE this task? This cannot be undone!">
						🗑️ Delete Forever
					</button>
				</div>
			</details>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(taskHTML)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

// DeleteTask handles DELETE /api/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Note: In production, you'd want to add proper authentication and role-based access control here

	// Extract task ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	taskID := pathParts[3] // /api/tasks/{id}

	// Validate task ID
	if err := validation.ValidateUserID(taskID); err != nil {
		http.Error(w, "Invalid task ID format", http.StatusBadRequest)
		return
	}

	// Delete task from database
	query := `DELETE FROM tasks WHERE id = ? AND family_id = ?`
	result, err := h.db.Exec(query, taskID, "fam1")
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	// Check if task was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Return success response (HTMX will remove the task from the DOM)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Task deleted")); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
