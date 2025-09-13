package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"famstack/internal/database"
)

// CalendarAPIHandler handles calendar-related API requests
type CalendarAPIHandler struct {
	db *database.DB
}

// NewCalendarAPIHandler creates a new calendar API handler
func NewCalendarAPIHandler(db *database.DB) *CalendarAPIHandler {
	return &CalendarAPIHandler{
		db: db,
	}
}

// UnifiedCalendarEvent represents a unified calendar event for the API
type UnifiedCalendarEvent struct {
	ID          string     `json:"id"`
	FamilyID    string     `json:"family_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Location    string     `json:"location"`
	AllDay      bool       `json:"all_day"`
	EventType   string     `json:"event_type"`
	Color       string     `json:"color"`
	CreatedBy   *string    `json:"created_by"`
	Priority    int        `json:"priority"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Attendees   []string   `json:"attendees"` // This will be populated from the attendee table
}

// GetEvents retrieves unified calendar events for a specific date or date range
func (h *CalendarAPIHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	date := r.URL.Query().Get("date")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	familyID := r.URL.Query().Get("family_id")

	// Default to current family if not specified
	if familyID == "" {
		familyID = "fam1" // Default family for now
	}

	var query string
	var args []any

	if date != "" {
		// Single date query
		query = `
			SELECT id, family_id, title, description, start_time, end_time, 
				   location, all_day, event_type, color, created_by, 
				   priority, status, created_at, updated_at
			FROM unified_calendar_events 
			WHERE family_id = ? AND status = 'active'
			AND (
				DATE(start_time) = DATE(?) OR 
				(end_time IS NOT NULL AND DATE(start_time) <= DATE(?) AND DATE(end_time) >= DATE(?))
			)
			ORDER BY start_time ASC
		`
		args = []any{familyID, date, date, date}
	} else if startDate != "" && endDate != "" {
		// Date range query
		query = `
			SELECT id, family_id, title, description, start_time, end_time, 
				   location, all_day, event_type, color, created_by, 
				   priority, status, created_at, updated_at
			FROM unified_calendar_events 
			WHERE family_id = ? AND status = 'active'
			AND start_time >= ? AND start_time <= ?
			ORDER BY start_time ASC
		`
		args = []any{familyID, startDate, endDate}
	} else {
		// Default to today's events
		today := time.Now().Format("2006-01-02")
		query = `
			SELECT id, family_id, title, description, start_time, end_time, 
				   location, all_day, event_type, color, created_by, 
				   priority, status, created_at, updated_at
			FROM unified_calendar_events 
			WHERE family_id = ? AND status = 'active'
			AND (
				DATE(start_time) = DATE(?) OR 
				(end_time IS NOT NULL AND DATE(start_time) <= DATE(?) AND DATE(end_time) >= DATE(?))
			)
			ORDER BY start_time ASC
		`
		args = []any{familyID, today, today, today}
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := make([]UnifiedCalendarEvent, 0)
	for rows.Next() {
		var event UnifiedCalendarEvent
		var endTime sql.NullTime
		var createdBy sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.FamilyID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&endTime,
			&event.Location,
			&event.AllDay,
			&event.EventType,
			&event.Color,
			&createdBy,
			&event.Priority,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to scan event: %v", err), http.StatusInternalServerError)
			return
		}

		// Handle nullable fields
		if endTime.Valid {
			event.EndTime = &endTime.Time
		}
		if createdBy.Valid {
			event.CreatedBy = &createdBy.String
		}

		// Load attendees from the attendee table
		attendees, err := h.getEventAttendees(event.ID)
		if err != nil {
			// Don't fail the whole request, just set empty attendees
			attendees = []string{}
		}
		event.Attendees = attendees

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getEventAttendees retrieves attendee user IDs for a given event
func (h *CalendarAPIHandler) getEventAttendees(eventID string) ([]string, error) {
	query := `SELECT user_id FROM unified_calendar_event_attendees WHERE event_id = ?`
	rows, err := h.db.Query(query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attendees := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		attendees = append(attendees, userID)
	}

	return attendees, rows.Err()
}

// CreateEvent creates a new unified calendar event
func (h *CalendarAPIHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event UnifiedCalendarEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Set defaults
	if event.FamilyID == "" {
		event.FamilyID = "fam1" // Default family
	}
	if event.EventType == "" {
		event.EventType = "event"
	}
	if event.Color == "" {
		event.Color = "#3b82f6"
	}
	if event.Status == "" {
		event.Status = "active"
	}

	// Insert into database
	query := `
		INSERT INTO unified_calendar_events 
		(family_id, title, description, start_time, end_time, location, all_day, 
		 event_type, color, created_by, priority, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`

	var newID string
	var createdAt, updatedAt time.Time
	err := h.db.QueryRow(query,
		event.FamilyID,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.Location,
		event.AllDay,
		event.EventType,
		event.Color,
		event.CreatedBy,
		event.Priority,
		event.Status,
	).Scan(&newID, &createdAt, &updatedAt)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create event: %v", err), http.StatusInternalServerError)
		return
	}

	// Insert attendees if provided
	if len(event.Attendees) > 0 {
		for _, userID := range event.Attendees {
			attendeeQuery := `
				INSERT INTO unified_calendar_event_attendees (event_id, user_id, response_status)
				VALUES (?, ?, 'accepted')
			`
			_, err := h.db.Exec(attendeeQuery, newID, userID)
			if err != nil {
				// Log error but don't fail the whole request
				fmt.Printf("Warning: Failed to add attendee %s to event %s: %v\n", userID, newID, err)
			}
		}
	}

	// Set the response data
	event.ID = newID
	event.CreatedAt = createdAt
	event.UpdatedAt = updatedAt

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateEvent updates a unified calendar event
func (h *CalendarAPIHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract event ID from URL path
	eventID := path.Base(r.URL.Path)
	if eventID == "" || eventID == "/" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Parse JSON data
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []any{}

	for field, value := range updates {
		switch field {
		case "title":
			if title, ok := value.(string); ok {
				setParts = append(setParts, "title = ?")
				args = append(args, title)
			}
		case "description":
			if description, ok := value.(string); ok {
				setParts = append(setParts, "description = ?")
				args = append(args, description)
			}
		case "start_time":
			if startTime, ok := value.(string); ok {
				setParts = append(setParts, "start_time = ?")
				args = append(args, startTime)
			}
		case "end_time":
			setParts = append(setParts, "end_time = ?")
			args = append(args, value) // Can be null
		case "location":
			if location, ok := value.(string); ok {
				setParts = append(setParts, "location = ?")
				args = append(args, location)
			}
		case "event_type":
			if eventType, ok := value.(string); ok {
				setParts = append(setParts, "event_type = ?")
				args = append(args, eventType)
			}
		case "color":
			if color, ok := value.(string); ok {
				setParts = append(setParts, "color = ?")
				args = append(args, color)
			}
		case "status":
			if status, ok := value.(string); ok {
				setParts = append(setParts, "status = ?")
				args = append(args, status)
			}
		case "priority":
			if priority, ok := value.(float64); ok {
				setParts = append(setParts, "priority = ?")
				args = append(args, int(priority))
			}
		case "attendees":
			if attendees, ok := value.([]any); ok {
				// Convert to JSON
				if data, err := json.Marshal(attendees); err == nil {
					setParts = append(setParts, "attendees = ?")
					args = append(args, string(data))
				}
			}
		}
	}

	if len(setParts) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// Add event ID to args
	args = append(args, eventID)

	// Execute update
	query := "UPDATE unified_calendar_events SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	result, err := h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update event: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check operation result", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Fetch and return the updated event
	h.GetEvent(w, r)
}

// GetEvent retrieves a specific unified calendar event
func (h *CalendarAPIHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract event ID from URL path
	eventID := path.Base(r.URL.Path)
	if eventID == "" || eventID == "/" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, family_id, title, description, start_time, end_time, 
			   location, all_day, event_type, color, attendees, created_by, 
			   priority, status, created_at, updated_at
		FROM unified_calendar_events WHERE id = ?
	`

	var event UnifiedCalendarEvent
	var endTime sql.NullTime
	var attendeesJSON sql.NullString
	var createdBy sql.NullString

	err := h.db.QueryRow(query, eventID).Scan(
		&event.ID,
		&event.FamilyID,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&endTime,
		&event.Location,
		&event.AllDay,
		&event.EventType,
		&event.Color,
		&attendeesJSON,
		&createdBy,
		&event.Priority,
		&event.Status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Event not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query event", http.StatusInternalServerError)
		}
		return
	}

	// Handle nullable fields
	if endTime.Valid {
		event.EndTime = &endTime.Time
	}
	if createdBy.Valid {
		event.CreatedBy = &createdBy.String
	}

	// Parse attendees JSON
	if attendeesJSON.Valid && attendeesJSON.String != "" {
		if err := json.Unmarshal([]byte(attendeesJSON.String), &event.Attendees); err != nil {
			event.Attendees = []string{}
		}
	} else {
		event.Attendees = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteEvent deletes a unified calendar event
func (h *CalendarAPIHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract event ID from URL path
	eventID := path.Base(r.URL.Path)
	if eventID == "" || eventID == "/" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Soft delete by setting status = 'cancelled'
	query := "UPDATE unified_calendar_events SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = ?"
	result, err := h.db.Exec(query, eventID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete event: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check operation result", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
