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

// Layered Calendar Data Structures for /api/calendar/days

// DaysResponse represents the response for multi-day calendar requests
type DaysResponse struct {
	StartDate       string               `json:"startDate"`
	EndDate         string               `json:"endDate"`
	Timezone        string               `json:"timezone"`
	RequestedPeople []string             `json:"requestedPeople"`
	Days            []DayView            `json:"days"`
	Metadata        DaysResponseMetadata `json:"metadata"`
}

// DaysResponseMetadata contains summary information about the response
type DaysResponseMetadata struct {
	TotalEvents  int       `json:"totalEvents"`
	LastUpdated  time.Time `json:"lastUpdated"`
	MaxDaysLimit int       `json:"maxDaysLimit"`
}

// DayView represents calendar view data for a single day with layered layout
type DayView struct {
	Date   string          `json:"date"`
	Layers []CalendarLayer `json:"layers"`
}

// CalendarLayer represents a column of non-overlapping events
type CalendarLayer struct {
	LayerIndex int                 `json:"layerIndex"`
	Events     []CalendarViewEvent `json:"events"`
}

// CalendarViewEvent represents an event with pre-calculated layout positioning
type CalendarViewEvent struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	StartSlot    int             `json:"startSlot"` // 0-359 (15-minute intervals)
	EndSlot      int             `json:"endSlot"`   // 0-359
	Color        string          `json:"color"`
	OwnerID      string          `json:"ownerId"`
	AttendeeIDs  []string        `json:"attendeeIds"`
	OverlapGroup int             `json:"overlapGroup"` // Total events in this overlap group
	OverlapIndex int             `json:"overlapIndex"` // Position within overlap group (0-based)
	Attendees    []EventAttendee `json:"attendees"`
	IsPrivate    bool            `json:"isPrivate"`
	Location     *string         `json:"location"`
	Description  *string         `json:"description"`
}
