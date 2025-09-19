package models

import "time"

// CalendarEvent represents a calendar event
type CalendarEvent struct {
	ID          string    `json:"id" db:"id"`
	FamilyID    string    `json:"family_id" db:"family_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	StartTime   time.Time `json:"start_time" db:"start_time"`
	EndTime     time.Time `json:"end_time" db:"end_time"`
	Location    *string   `json:"location" db:"location"`
	EventType   string    `json:"event_type" db:"event_type"` // 'appointment', 'event', 'reminder'
	AssignedTo  *string   `json:"assigned_to" db:"assigned_to"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UnifiedCalendarEvent represents a calendar event from external integrations
type UnifiedCalendarEvent struct {
	ID              string    `json:"id" db:"id"`
	FamilyID        string    `json:"family_id" db:"family_id"`
	IntegrationID   string    `json:"integration_id" db:"integration_id"`
	ExternalEventID string    `json:"external_event_id" db:"external_event_id"`
	Title           string    `json:"title" db:"title"`
	Description     *string   `json:"description" db:"description"`
	StartTime       time.Time `json:"start_time" db:"start_time"`
	EndTime         time.Time `json:"end_time" db:"end_time"`
	Location        *string   `json:"location" db:"location"`
	Organizer       *string   `json:"organizer" db:"organizer"`
	Attendees       *string   `json:"attendees" db:"attendees"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// EventType constants
const (
	EventTypeAppointment = "appointment"
	EventTypeEvent       = "event"
	EventTypeReminder    = "reminder"
)

// IsValidEventType checks if an event type is valid
func IsValidEventType(eventType string) bool {
	switch eventType {
	case EventTypeAppointment, EventTypeEvent, EventTypeReminder:
		return true
	default:
		return false
	}
}
