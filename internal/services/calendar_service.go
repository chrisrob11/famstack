package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
)

// CalendarService handles all calendar and event database operations
type CalendarService struct {
	db *database.Fascade
}

// CalendarEventForSync represents a calendar event for sync operations
type CalendarEventForSync struct {
	ID          string     `json:"id"`
	FamilyID    string     `json:"family_id"`
	CreatedBy   string     `json:"created_by"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Location    string     `json:"location"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	AllDay      bool       `json:"all_day"`
	Attendees   []string   `json:"attendees"`
	SourceType  string     `json:"source_type"`
	SourceID    string     `json:"source_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewCalendarService creates a new calendar service
func NewCalendarService(db *database.Fascade) *CalendarService {
	return &CalendarService{db: db}
}

// GetEvent returns a calendar event by ID
func (s *CalendarService) GetEvent(eventID string) (*models.CalendarEvent, error) {
	query := `
		SELECT id, family_id, title, description, start_time, end_time,
			   location, event_type, assigned_to, created_by, created_at, updated_at
		FROM calendar_events
		WHERE id = ?
	`

	var event models.CalendarEvent
	var description, location, assignedTo sql.NullString

	err := s.db.QueryRow(query, eventID).Scan(
		&event.ID, &event.FamilyID, &event.Title, &description,
		&event.StartTime, &event.EndTime, &location, &event.EventType,
		&assignedTo, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("calendar event not found")
		}
		return nil, fmt.Errorf("failed to get calendar event: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		event.Description = &description.String
	}
	if location.Valid {
		event.Location = &location.String
	}
	if assignedTo.Valid {
		event.AssignedTo = &assignedTo.String
	}

	return &event, nil
}

// ListEvents returns calendar events for a family within a date range
func (s *CalendarService) ListEvents(familyID string, startDate, endDate time.Time) ([]models.CalendarEvent, error) {
	query := `
		SELECT id, family_id, title, description, start_time, end_time,
			   location, event_type, assigned_to, created_by, created_at, updated_at
		FROM calendar_events
		WHERE family_id = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := s.db.Query(query, familyID, startDate, endDate)
	if err != nil {
		return []models.CalendarEvent{}, fmt.Errorf("failed to list calendar events: %w", err)
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		event, scanErr := s.scanCalendarEvent(rows)
		if scanErr != nil {
			return []models.CalendarEvent{}, fmt.Errorf("failed to scan calendar event: %w", scanErr)
		}
		events = append(events, *event)
	}

	if err = rows.Err(); err != nil {
		return []models.CalendarEvent{}, fmt.Errorf("error iterating calendar events: %w", err)
	}

	// Ensure we always return a non-nil slice
	if events == nil {
		events = []models.CalendarEvent{}
	}

	return events, nil
}

// ListEventsByMember returns calendar events assigned to a specific family member
func (s *CalendarService) ListEventsByMember(memberID string, startDate, endDate time.Time) ([]models.CalendarEvent, error) {
	query := `
		SELECT id, family_id, title, description, start_time, end_time,
			   location, event_type, assigned_to, created_by, created_at, updated_at
		FROM calendar_events
		WHERE assigned_to = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := s.db.Query(query, memberID, startDate, endDate)
	if err != nil {
		return []models.CalendarEvent{}, fmt.Errorf("failed to list events by member: %w", err)
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		event, scanErr := s.scanCalendarEvent(rows)
		if scanErr != nil {
			return []models.CalendarEvent{}, fmt.Errorf("failed to scan calendar event: %w", scanErr)
		}
		events = append(events, *event)
	}

	if err = rows.Err(); err != nil {
		return []models.CalendarEvent{}, fmt.Errorf("error iterating calendar events: %w", err)
	}

	// Ensure we always return a non-nil slice
	if events == nil {
		events = []models.CalendarEvent{}
	}

	return events, nil
}

// CreateEvent creates a new calendar event
func (s *CalendarService) CreateEvent(familyID, createdBy string, req *models.CreateCalendarEventRequest) (*models.CalendarEvent, error) {
	eventID := generateEventID()
	now := time.Now()

	query := `
		INSERT INTO calendar_events (id, family_id, title, description, start_time, end_time,
									location, event_type, assigned_to, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		eventID, familyID, req.Title, req.Description, req.StartTime, req.EndTime,
		req.Location, req.EventType, req.AssignedTo, createdBy, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create calendar event: %w", err)
	}

	return s.GetEvent(eventID)
}

// UpdateEvent updates an existing calendar event
func (s *CalendarService) UpdateEvent(eventID string, req *models.UpdateCalendarEventRequest) (*models.CalendarEvent, error) {
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
	if req.StartTime != nil {
		setParts = append(setParts, "start_time = ?")
		args = append(args, *req.StartTime)
	}
	if req.EndTime != nil {
		setParts = append(setParts, "end_time = ?")
		args = append(args, *req.EndTime)
	}
	if req.Location != nil {
		setParts = append(setParts, "location = ?")
		args = append(args, *req.Location)
	}
	if req.EventType != nil {
		setParts = append(setParts, "event_type = ?")
		args = append(args, *req.EventType)
	}
	if req.AssignedTo != nil {
		setParts = append(setParts, "assigned_to = ?")
		args = append(args, *req.AssignedTo)
	}

	if len(setParts) == 1 { // Only updated_at
		return s.GetEvent(eventID) // No changes, return current
	}

	// Add eventID to args for WHERE clause
	args = append(args, eventID)

	query := fmt.Sprintf(`
		UPDATE calendar_events
		SET %s
		WHERE id = ?
	`, strings.Join(setParts, ", "))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update calendar event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("calendar event not found")
	}

	return s.GetEvent(eventID)
}

// DeleteEvent deletes a calendar event
func (s *CalendarService) DeleteEvent(eventID string) error {
	query := `DELETE FROM calendar_events WHERE id = ?`

	result, err := s.db.Exec(query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete calendar event: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("calendar event not found")
	}

	return nil
}

// GetUnifiedCalendarEvents returns unified calendar events (from external integrations)
func (s *CalendarService) GetUnifiedCalendarEvents(familyID string, startDate, endDate time.Time) ([]models.UnifiedCalendarEvent, error) {
	query := `
		SELECT id, family_id, title, description, start_time, end_time, location,
			   all_day, event_type, color, created_by, priority, status, created_at, updated_at
		FROM unified_calendar_events
		WHERE family_id = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := s.db.Query(query, familyID, startDate, endDate)
	if err != nil {
		return []models.UnifiedCalendarEvent{}, fmt.Errorf("failed to list unified calendar events: %w", err)
	}
	defer rows.Close()

	var events []models.UnifiedCalendarEvent
	for rows.Next() {
		event, scanErr := s.scanUnifiedCalendarEvent(rows)
		if scanErr != nil {
			return []models.UnifiedCalendarEvent{}, fmt.Errorf("failed to scan unified calendar event: %w", scanErr)
		}
		events = append(events, *event)
	}

	if err = rows.Err(); err != nil {
		return []models.UnifiedCalendarEvent{}, fmt.Errorf("error iterating unified calendar events: %w", err)
	}

	// Ensure we always return a non-nil slice
	if len(events) == 0 {
		return []models.UnifiedCalendarEvent{}, nil
	}

	// Step 2: Collect all event IDs
	eventIDs := make([]string, len(events))
	for i, event := range events {
		eventIDs[i] = event.ID
	}

	// Step 3: Fetch all attendees for these events in a single query
	attendeeQuery := `
		SELECT event_id, user_id
		FROM unified_calendar_event_attendees
		WHERE event_id IN (?` + strings.Repeat(",?", len(eventIDs)-1) + `)
	`
	args := make([]interface{}, len(eventIDs))
	for i, id := range eventIDs {
		args[i] = id
	}

	attendeeRows, err := s.db.Query(attendeeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query for attendees: %w", err)
	}
	defer attendeeRows.Close()

	// Step 4: Map attendees to their event ID
	attendeeMap := make(map[string][]string)
	for attendeeRows.Next() {
		var eventID, userID string
		if err = attendeeRows.Scan(&eventID, &userID); err != nil {
			return nil, fmt.Errorf("failed to scan attendee: %w", err)
		}
		attendeeMap[eventID] = append(attendeeMap[eventID], userID)
	}
	if err = attendeeRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attendee rows: %w", err)
	}

	// Step 5: Attach attendees to the events
	for i, event := range events {
		if attendees, ok := attendeeMap[event.ID]; ok {
			events[i].Attendees = attendees
		} else {
			events[i].Attendees = []string{} // Ensure it's an empty slice, not nil
		}
	}

	return events, nil
}

// CreateUnifiedCalendarEvent creates a unified calendar event (from external integration)
func (s *CalendarService) CreateUnifiedCalendarEvent(req *models.CreateUnifiedCalendarEventRequest) (*models.UnifiedCalendarEvent, error) {
	eventID := generateUnifiedEventID()
	now := time.Now()

	query := `
		INSERT INTO unified_calendar_events (id, family_id, integration_id, external_event_id,
											title, description, start_time, end_time, location,
											organizer, attendees, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		eventID, req.FamilyID, req.IntegrationID, req.ExternalEventID,
		req.Title, req.Description, req.StartTime, req.EndTime, req.Location,
		req.Organizer, req.Attendees, now, now,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create unified calendar event: %w", err)
	}

	return s.GetUnifiedCalendarEvent(eventID)
}

// GetUnifiedCalendarEvent returns a unified calendar event by ID
func (s *CalendarService) GetUnifiedCalendarEvent(eventID string) (*models.UnifiedCalendarEvent, error) {
	query := `
		SELECT id, family_id, title, description, start_time, end_time, location,
			   all_day, event_type, color, created_by, priority, status, created_at, updated_at
		FROM unified_calendar_events
		WHERE id = ?
	`

	var event models.UnifiedCalendarEvent
	var description, location, createdBy sql.NullString

	err := s.db.QueryRow(query, eventID).Scan(
		&event.ID, &event.FamilyID, &event.Title, &description,
		&event.StartTime, &event.EndTime, &location, &event.AllDay,
		&event.EventType, &event.Color, &createdBy, &event.Priority,
		&event.Status, &event.CreatedAt, &event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("unified calendar event not found")
		}
		return nil, fmt.Errorf("failed to get unified calendar event: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		event.Description = &description.String
	}
	if location.Valid {
		event.Location = &location.String
	}
	if createdBy.Valid {
		event.CreatedBy = &createdBy.String
	}

	return &event, nil
}

// UpsertCalendarEvent inserts or updates a calendar event from external sync
func (s *CalendarService) UpsertCalendarEvent(event *CalendarEventForSync) error {
	query := `
		INSERT OR REPLACE INTO calendar_events
		(id, family_id, created_by, title, description, location, start_time, end_time,
		 all_day, attendees, source_type, source_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	attendeesJSON := "[]"
	if len(event.Attendees) > 0 {
		// Simple JSON encoding for attendees
		attendeesJSON = `["` + strings.Join(event.Attendees, `","`) + `"]`
	}

	_, err := s.db.Exec(query,
		event.ID, event.FamilyID, event.CreatedBy, event.Title, event.Description,
		event.Location, event.StartTime, event.EndTime, event.AllDay,
		attendeesJSON, event.SourceType, event.SourceID,
		event.CreatedAt, event.UpdatedAt,
	)

	return err
}

// GetSyncSettings retrieves sync settings for a user
func (s *CalendarService) GetSyncSettings(userID string) (*SyncSettings, error) {
	query := `
		SELECT sync_frequency_minutes, sync_range_days
		FROM calendar_sync_settings
		WHERE created_by = ?
	`

	var settings SyncSettings
	err := s.db.QueryRow(query, userID).Scan(&settings.SyncFrequencyMinutes, &settings.SyncRangeDays)
	if err != nil {
		// Return default settings if not found
		return &SyncSettings{
			SyncFrequencyMinutes: 15,
			SyncRangeDays:        30,
		}, nil
	}

	return &settings, nil
}

// UpdateSyncStatus updates the sync status for a user
func (s *CalendarService) UpdateSyncStatus(userID, status, errorMsg string, eventsSynced int) error {
	query := `
		INSERT OR REPLACE INTO calendar_sync_settings
		(created_by, last_sync_at, sync_status, sync_error, events_synced, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, userID, time.Now(), status, errorMsg, eventsSynced, time.Now())
	return err
}

// SyncSettings represents calendar sync configuration
type SyncSettings struct {
	SyncFrequencyMinutes int `json:"sync_frequency_minutes"`
	SyncRangeDays        int `json:"sync_range_days"`
}

// Helper functions

func (s *CalendarService) scanCalendarEvent(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.CalendarEvent, error) {
	var event models.CalendarEvent
	var description, location, assignedTo sql.NullString

	err := scanner.Scan(
		&event.ID, &event.FamilyID, &event.Title, &description,
		&event.StartTime, &event.EndTime, &location, &event.EventType,
		&assignedTo, &event.CreatedBy, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if description.Valid {
		event.Description = &description.String
	}
	if location.Valid {
		event.Location = &location.String
	}
	if assignedTo.Valid {
		event.AssignedTo = &assignedTo.String
	}

	return &event, nil
}

func (s *CalendarService) scanUnifiedCalendarEvent(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.UnifiedCalendarEvent, error) {
	var event models.UnifiedCalendarEvent
	var description, location, createdBy sql.NullString

	err := scanner.Scan(
		&event.ID, &event.FamilyID, &event.Title, &description,
		&event.StartTime, &event.EndTime, &location, &event.AllDay,
		&event.EventType, &event.Color, &createdBy, &event.Priority,
		&event.Status, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if description.Valid {
		event.Description = &description.String
	}
	if location.Valid {
		event.Location = &location.String
	}
	if createdBy.Valid {
		event.CreatedBy = &createdBy.String
	}

	return &event, nil
}

func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

func generateUnifiedEventID() string {
	return fmt.Sprintf("unified_event_%d", time.Now().UnixNano())
}
