package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/jobsystem"
)

type ScheduleHandler struct {
	db        *database.DB
	jobSystem *jobsystem.SQLiteJobSystem
}

func NewScheduleHandler(db *database.DB) *ScheduleHandler {
	return &ScheduleHandler{
		db: db,
	}
}

func NewScheduleHandlerWithJobSystem(db *database.DB, jobSystem *jobsystem.SQLiteJobSystem) *ScheduleHandler {
	return &ScheduleHandler{
		db:        db,
		jobSystem: jobSystem,
	}
}

type TaskSchedule struct {
	ID          string   `json:"id"`
	FamilyID    string   `json:"family_id"`
	CreatedBy   string   `json:"created_by"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TaskType    string   `json:"task_type"`
	AssignedTo  *string  `json:"assigned_to"`
	DaysOfWeek  []string `json:"days_of_week"`
	TimeOfDay   *string  `json:"time_of_day"`
	Priority    int      `json:"priority"`
	Points      int      `json:"points"`
	Active      bool     `json:"active"`
	CreatedAt   string   `json:"created_at"`
}

type CreateScheduleRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TaskType    string   `json:"task_type"`
	AssignedTo  *string  `json:"assigned_to"`
	DaysOfWeek  []string `json:"days_of_week"`
	TimeOfDay   *string  `json:"time_of_day"`
	Priority    int      `json:"priority"`
	Points      int      `json:"points"`
	FamilyID    string   `json:"family_id"`
}

type QueueItem struct {
	ID           string  `json:"id"`
	QueueType    string  `json:"queue_type"`
	Payload      string  `json:"payload"`
	Status       string  `json:"status"`
	Attempts     int     `json:"attempts"`
	MaxAttempts  int     `json:"max_attempts"`
	ScheduledFor string  `json:"scheduled_for"`
	CreatedAt    string  `json:"created_at"`
	ProcessedAt  *string `json:"processed_at"`
	ErrorMessage *string `json:"error_message"`
}

// extractIDFromPath extracts the ID from the URL path
func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remaining := strings.TrimPrefix(path, prefix)
	if remaining == "" {
		return ""
	}
	// Remove leading slash
	remaining = strings.TrimPrefix(remaining, "/")
	// Get first segment (ID)
	parts := strings.Split(remaining, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	familyID := r.URL.Query().Get("family_id")
	if familyID == "" {
		familyID = "fam1" // Default family for now
	}

	query := `
		SELECT id, family_id, created_by, title, description, task_type, 
			   assigned_to, days_of_week, time_of_day, priority, points, active, created_at
		FROM task_schedules 
		WHERE family_id = ? AND active = true
		ORDER BY created_at DESC
	`

	rows, err := h.db.Query(query, familyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query schedules: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	schedules := make([]TaskSchedule, 0)
	for rows.Next() {
		var schedule TaskSchedule
		var assignedTo sql.NullString
		var timeOfDay sql.NullString
		var daysOfWeekJSON string

		scanErr := rows.Scan(
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
			&schedule.Active,
			&schedule.CreatedAt,
		)
		if scanErr != nil {
			http.Error(w, fmt.Sprintf("Failed to scan schedule: %v", scanErr), http.StatusInternalServerError)
			return
		}

		// Handle nullable fields
		if assignedTo.Valid {
			schedule.AssignedTo = &assignedTo.String
		}
		if timeOfDay.Valid {
			schedule.TimeOfDay = &timeOfDay.String
		}

		// Parse days of week JSON
		err = json.Unmarshal([]byte(daysOfWeekJSON), &schedule.DaysOfWeek)
		if err != nil {
			log.Printf("Failed to parse days_of_week for schedule %s: %v", schedule.ID, err)
			schedule.DaysOfWeek = []string{}
		}

		schedules = append(schedules, schedule)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedules); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if req.TaskType == "" {
		http.Error(w, "Task type is required", http.StatusBadRequest)
		return
	}
	if len(req.DaysOfWeek) == 0 {
		http.Error(w, "At least one day of the week is required", http.StatusBadRequest)
		return
	}
	if req.FamilyID == "" {
		req.FamilyID = "fam1" // Default family
	}

	// Validate days of week
	validDays := map[string]bool{
		"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
		"friday": true, "saturday": true, "sunday": true,
	}
	for _, day := range req.DaysOfWeek {
		if !validDays[strings.ToLower(day)] {
			http.Error(w, fmt.Sprintf("Invalid day of week: %s", day), http.StatusBadRequest)
			return
		}
	}

	// Convert days of week to JSON
	daysOfWeekJSON, err := json.Marshal(req.DaysOfWeek)
	if err != nil {
		http.Error(w, "Failed to encode days of week", http.StatusInternalServerError)
		return
	}

	// Handle assigned_to value
	var assignedToValue interface{}
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		assignedToValue = *req.AssignedTo
	} else {
		assignedToValue = nil
	}

	// Handle time_of_day value
	var timeOfDayValue interface{}
	if req.TimeOfDay != nil && *req.TimeOfDay != "" {
		timeOfDayValue = *req.TimeOfDay
	} else {
		timeOfDayValue = nil
	}

	// Insert into database
	query := `
		INSERT INTO task_schedules (family_id, created_by, title, description, task_type,
								   assigned_to, days_of_week, time_of_day, priority, points)
		VALUES (?, 'user1', ?, ?, ?, ?, ?, ?, ?, ?) 
		RETURNING id, created_at
	`

	var newID, createdAt string
	err = h.db.QueryRow(query,
		req.FamilyID,
		req.Title,
		req.Description,
		req.TaskType,
		assignedToValue,
		string(daysOfWeekJSON),
		timeOfDayValue,
		req.Priority,
		req.Points,
	).Scan(&newID, &createdAt)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create schedule: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the response schedule object
	schedule := TaskSchedule{
		ID:          newID,
		FamilyID:    req.FamilyID,
		CreatedBy:   "user1", // TODO: Get from auth context
		Title:       req.Title,
		Description: req.Description,
		TaskType:    req.TaskType,
		AssignedTo:  req.AssignedTo,
		DaysOfWeek:  req.DaysOfWeek,
		TimeOfDay:   req.TimeOfDay,
		Priority:    req.Priority,
		Points:      req.Points,
		Active:      true,
		CreatedAt:   createdAt,
	}

	// Queue initial task generation for the next 14 days
	if h.jobSystem != nil {
		err = h.queueTaskGeneration(newID)
		if err != nil {
			log.Printf("Failed to queue task generation for schedule %s: %v", newID, err)
			// Don't fail the request, just log the error
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(schedule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := extractIDFromPath(r.URL.Path, "/api/v1/schedules")

	query := `
		SELECT id, family_id, created_by, title, description, task_type,
			   assigned_to, days_of_week, time_of_day, priority, points, active, created_at
		FROM task_schedules WHERE id = ?
	`

	var schedule TaskSchedule
	var assignedTo sql.NullString
	var timeOfDay sql.NullString
	var daysOfWeekJSON string

	err := h.db.QueryRow(query, scheduleID).Scan(
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
		&schedule.Active,
		&schedule.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get schedule: %v", err), http.StatusInternalServerError)
		return
	}

	// Handle nullable fields
	if assignedTo.Valid {
		schedule.AssignedTo = &assignedTo.String
	}
	if timeOfDay.Valid {
		schedule.TimeOfDay = &timeOfDay.String
	}

	// Parse days of week JSON
	err = json.Unmarshal([]byte(daysOfWeekJSON), &schedule.DaysOfWeek)
	if err != nil {
		log.Printf("Failed to parse days_of_week for schedule %s: %v", schedule.ID, err)
		schedule.DaysOfWeek = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := extractIDFromPath(r.URL.Path, "/api/v1/schedules")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}

	for field, value := range updates {
		switch field {
		case "title", "description", "task_type", "priority", "points":
			setParts = append(setParts, fmt.Sprintf("%s = ?", field))
			args = append(args, value)
		case "assigned_to":
			setParts = append(setParts, "assigned_to = ?")
			if value == nil || value == "" {
				args = append(args, nil)
			} else {
				args = append(args, value)
			}
		case "time_of_day":
			setParts = append(setParts, "time_of_day = ?")
			if value == nil || value == "" {
				args = append(args, nil)
			} else {
				args = append(args, value)
			}
		case "days_of_week":
			daysJSON, err := json.Marshal(value)
			if err != nil {
				http.Error(w, "Invalid days_of_week format", http.StatusBadRequest)
				return
			}
			setParts = append(setParts, "days_of_week = ?")
			args = append(args, string(daysJSON))
		case "active":
			setParts = append(setParts, "active = ?")
			args = append(args, value)
		}
	}

	if len(setParts) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	args = append(args, scheduleID)
	query := fmt.Sprintf("UPDATE task_schedules SET %s WHERE id = ?", strings.Join(setParts, ", "))

	result, err := h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update schedule: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check operation result", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	// If the schedule was updated, re-queue task generation
	if h.jobSystem != nil {
		err = h.queueTaskGeneration(scheduleID)
		if err != nil {
			log.Printf("Failed to queue task generation after update for schedule %s: %v", scheduleID, err)
		}
	}

	// Fetch and return the updated schedule
	query = `
		SELECT id, family_id, created_by, title, description, task_type,
			   assigned_to, days_of_week, time_of_day, priority, points, active, created_at
		FROM task_schedules WHERE id = ?
	`

	var schedule TaskSchedule
	var assignedTo sql.NullString
	var timeOfDay sql.NullString
	var daysOfWeekJSON string

	err = h.db.QueryRow(query, scheduleID).Scan(
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
		&schedule.Active,
		&schedule.CreatedAt,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch updated schedule: %v", err), http.StatusInternalServerError)
		return
	}

	// Handle nullable fields
	if assignedTo.Valid {
		schedule.AssignedTo = &assignedTo.String
	}
	if timeOfDay.Valid {
		schedule.TimeOfDay = &timeOfDay.String
	}

	// Parse days of week JSON
	err = json.Unmarshal([]byte(daysOfWeekJSON), &schedule.DaysOfWeek)
	if err != nil {
		log.Printf("Failed to parse days_of_week for schedule %s: %v", schedule.ID, err)
		schedule.DaysOfWeek = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	scheduleID := extractIDFromPath(r.URL.Path, "/api/v1/schedules")

	// Soft delete by setting active = false
	query := "UPDATE task_schedules SET active = false WHERE id = ?"
	result, err := h.db.Exec(query, scheduleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete schedule: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check operation result", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	// Note: With the job system, pending jobs will naturally not execute
	// if the schedule is deleted since the handler checks if the schedule exists

	w.WriteHeader(http.StatusOK)
}

// queueTaskGeneration queues task generation for a schedule over the next 14 days
func (h *ScheduleHandler) queueTaskGeneration(scheduleID string) error {
	now := time.Now()

	// Generate jobs for the next 14 days
	for i := range 14 {
		targetDate := now.AddDate(0, 0, i)
		payload := map[string]any{
			"schedule_id": scheduleID,
			"target_date": targetDate.Format("2006-01-02"),
		}

		// Schedule processing for midnight of the target date minus 1 day
		scheduledFor := targetDate.AddDate(0, 0, -1).Truncate(24 * time.Hour)
		if scheduledFor.Before(now) {
			scheduledFor = now // Process immediately if it's in the past
		}

		_, err := h.jobSystem.Enqueue(&jobsystem.EnqueueRequest{
			QueueName:  "task_generation",
			JobType:    "generate_scheduled_task",
			Payload:    payload,
			Priority:   0,
			MaxRetries: 3,
			RunAt:      &scheduledFor,
		})
		if err != nil {
			return fmt.Errorf("failed to enqueue task generation job: %w", err)
		}
	}

	return nil
}
