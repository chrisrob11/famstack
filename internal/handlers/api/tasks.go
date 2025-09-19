package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"famstack/internal/auth"
	"famstack/internal/models"
	"famstack/internal/services"
)

// TaskAPIHandler handles JSON API requests for tasks
type TaskAPIHandler struct {
	tasksService *services.TasksService
}

// NewTaskAPIHandler creates a new task API handler
func NewTaskAPIHandler(tasksService *services.TasksService) *TaskAPIHandler {
	return &TaskAPIHandler{tasksService: tasksService}
}

// These types are now in services.TasksService, so we use those directly

// ListTasks returns all tasks as JSON
func (h *TaskAPIHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
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

	// Use the service to get tasks by family
	tasksResponse, err := h.tasksService.ListTasksByFamily(user.FamilyID, dateFilter)
	if err != nil {
		http.Error(w, "Failed to load tasks", http.StatusInternalServerError)
		return
	}

	// The service response is already in the correct format for JSON output
	response := map[string]interface{}{
		"tasks_by_member": tasksResponse.TasksByMember,
		"date":            time.Now().Format("Monday, January 2"),
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

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

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

	// Create request object for the service
	createReq := &models.CreateTaskRequest{
		Title:       task.Title,
		Description: task.Description,
		TaskType:    task.TaskType,
		AssignedTo:  task.AssignedTo,
		Priority:    task.Priority,
		DueDate:     task.DueDate,
		Points:      0, // Default value since not provided in this API
	}

	// Use the service to create the task
	createdTask, err := h.tasksService.CreateTask(user.FamilyID, user.ID, createReq)
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

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(createdTask); err != nil {
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

	// Create update request object for the service
	updateReq := &models.UpdateTaskRequest{}

	// Handle status updates
	if status, exists := updateData["status"]; exists {
		statusStr, ok := status.(string)
		if !ok {
			http.Error(w, "Invalid status format", http.StatusBadRequest)
			return
		}
		switch statusStr {
		case "completed", "pending":
			updateReq.Status = &statusStr
		default:
			http.Error(w, "Invalid status", http.StatusBadRequest)
			return
		}
	}

	// Handle assignment updates
	if assignedTo, exists := updateData["assigned_to"]; exists {
		if assignedTo == nil || assignedTo == "" {
			emptyString := ""
			updateReq.AssignedTo = &emptyString
		} else {
			assignedToStr, ok := assignedTo.(string)
			if !ok {
				http.Error(w, "Invalid assigned_to format", http.StatusBadRequest)
				return
			}
			updateReq.AssignedTo = &assignedToStr
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
		updateReq.Title = &titleStr
	}

	// Use the service to update the task
	task, err := h.tasksService.UpdateTask(taskID, updateReq)
	if err != nil {
		if err.Error() == "task not found" {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
		}
		return
	}

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

	// Use the service to delete the task
	err := h.tasksService.DeleteTask(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTask retrieves a single task
func (h *TaskAPIHandler) GetTask(w http.ResponseWriter, r *http.Request, taskID string) {
	// Use the service to get the task
	task, err := h.tasksService.GetTask(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get task", http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, "Failed to encode task", http.StatusInternalServerError)
		return
	}
}
