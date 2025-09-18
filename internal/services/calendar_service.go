package services

import (
	"database/sql"
	"fmt"
	"time"

	"famstack/internal/models"
)

// CalendarService handles all calendar and event database operations
type CalendarService struct {
	db *sql.DB
}

// NewCalendarService creates a new calendar service
func NewCalendarService(db *sql.DB) *CalendarService {
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
		return nil, fmt.Errorf("failed to list calendar events: %w", err)
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		event, scanErr := s.scanCalendarEvent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan calendar event: %w", scanErr)
		}
		events = append(events, *event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calendar events: %w", err)
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
		return nil, fmt.Errorf("failed to list events by member: %w", err)
	}
	defer rows.Close()

	var events []models.CalendarEvent
	for rows.Next() {
		event, scanErr := s.scanCalendarEvent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan calendar event: %w", scanErr)
		}
		events = append(events, *event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calendar events: %w", err)
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
	`, joinStrings(setParts, ", "))

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
		SELECT id, family_id, integration_id, external_event_id, title, description,
			   start_time, end_time, location, organizer, attendees, created_at, updated_at
		FROM unified_calendar_events
		WHERE family_id = ? AND start_time >= ? AND start_time <= ?
		ORDER BY start_time ASC
	`

	rows, err := s.db.Query(query, familyID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to list unified calendar events: %w", err)
	}
	defer rows.Close()

	var events []models.UnifiedCalendarEvent
	for rows.Next() {
		event, scanErr := s.scanUnifiedCalendarEvent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan unified calendar event: %w", scanErr)
		}
		events = append(events, *event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating unified calendar events: %w", err)
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
		SELECT id, family_id, integration_id, external_event_id, title, description,
			   start_time, end_time, location, organizer, attendees, created_at, updated_at
		FROM unified_calendar_events
		WHERE id = ?
	`

	var event models.UnifiedCalendarEvent
	var description, location, organizer, attendees sql.NullString

	err := s.db.QueryRow(query, eventID).Scan(
		&event.ID, &event.FamilyID, &event.IntegrationID, &event.ExternalEventID,
		&event.Title, &description, &event.StartTime, &event.EndTime, &location,
		&organizer, &attendees, &event.CreatedAt, &event.UpdatedAt,
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
	if organizer.Valid {
		event.Organizer = &organizer.String
	}
	if attendees.Valid {
		event.Attendees = &attendees.String
	}

	return &event, nil
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
	var description, location, organizer, attendees sql.NullString

	err := scanner.Scan(
		&event.ID, &event.FamilyID, &event.IntegrationID, &event.ExternalEventID,
		&event.Title, &description, &event.StartTime, &event.EndTime, &location,
		&organizer, &attendees, &event.CreatedAt, &event.UpdatedAt,
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
	if organizer.Valid {
		event.Organizer = &organizer.String
	}
	if attendees.Valid {
		event.Attendees = &attendees.String
	}

	return &event, nil
}

func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

func generateUnifiedEventID() string {
	return fmt.Sprintf("unified_event_%d", time.Now().UnixNano())
}
