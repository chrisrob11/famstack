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

// EventAttendee represents an attendee with display information for events
type EventAttendee struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Initial  string `json:"initial"`
	Color    string `json:"color"`
	Response string `json:"response"` // needsAction, accepted, declined, tentative
}

// UnifiedCalendarEvent represents a calendar event from external integrations
type UnifiedCalendarEvent struct {
	ID          string    `json:"id" db:"id"`
	FamilyID    string    `json:"family_id" db:"family_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	StartTime   time.Time `json:"start_time" db:"start_time"`
	EndTime     time.Time `json:"end_time" db:"end_time"`
	Location    *string   `json:"location" db:"location"`
	AllDay      bool      `json:"all_day" db:"all_day"`
	EventType   string    `json:"event_type" db:"event_type"`
	Color       string    `json:"color" db:"color"`
	CreatedBy   *string   `json:"created_by" db:"created_by"`
	Priority    int       `json:"priority" db:"priority"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Attendees is a constructed field with full family member display data.
	// This replaces the previous []string approach to provide richer UI data.
	Attendees []EventAttendee `json:"attendees"`
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
