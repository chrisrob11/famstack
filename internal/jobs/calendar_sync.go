package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"famstack/internal/calendar"
	"famstack/internal/database"
	"famstack/internal/jobsystem"
	"famstack/internal/oauth"
)

// CalendarSyncJobType represents the job type for calendar synchronization
const CalendarSyncJobType = "calendar_sync"

// CalendarSyncPayload represents the payload for calendar sync jobs
type CalendarSyncPayload struct {
	UserID     string `json:"user_id"`
	FamilyID   string `json:"family_id"`
	Provider   string `json:"provider"`
	CalendarID string `json:"calendar_id,omitempty"`
	ForceSync  bool   `json:"force_sync,omitempty"`
}

// CalendarSyncHandler handles calendar synchronization jobs
type CalendarSyncHandler struct {
	db           *database.DB
	oauthService *oauth.Service
	googleClient *calendar.GoogleClient
}

// NewCalendarSyncHandler creates a new calendar sync handler
func NewCalendarSyncHandler(db *database.DB, oauthService *oauth.Service, googleClient *calendar.GoogleClient) *CalendarSyncHandler {
	return &CalendarSyncHandler{
		db:           db,
		oauthService: oauthService,
		googleClient: googleClient,
	}
}

// Handle processes calendar sync jobs
func (h *CalendarSyncHandler) Handle(ctx context.Context, job *jobsystem.Job) error {
	var payload CalendarSyncPayload
	data, err := json.Marshal(job.Payload)
	if err != nil {
		return fmt.Errorf("cannot marshal job payload: %w", err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("Starting calendar sync for user %s, provider %s", payload.UserID, payload.Provider)

	// Update sync status to 'syncing'
	if err := h.updateSyncStatus(payload.UserID, "syncing", "", 0); err != nil {
		log.Printf("Failed to update sync status: %v", err)
	}

	switch payload.Provider {
	case "google":
		return h.syncGoogleCalendar(ctx, payload)
	default:
		return fmt.Errorf("unsupported provider: %s", payload.Provider)
	}
}

// syncGoogleCalendar synchronizes Google Calendar events
func (h *CalendarSyncHandler) syncGoogleCalendar(ctx context.Context, payload CalendarSyncPayload) error {
	// Get sync settings for user
	settings, err := h.getSyncSettings(payload.UserID)
	if err != nil {
		return fmt.Errorf("failed to get sync settings: %w", err)
	}

	// Calculate time range for sync
	now := time.Now()
	timeMin := now.Truncate(24 * time.Hour) // Start of today
	timeMax := timeMin.AddDate(0, 0, settings.SyncRangeDays)

	var totalEventsSynced int

	// If no specific calendar ID, get all calendars for user
	if payload.CalendarID == "" {
		calendars, err := h.googleClient.GetCalendars(payload.UserID)
		if err != nil {
			if updateErr := h.updateSyncStatus(payload.UserID, "error", fmt.Sprintf("Failed to get calendars: %v", err), 0); updateErr != nil {
				log.Printf("Failed to update sync status: %v", updateErr)
			}
			return fmt.Errorf("failed to get calendars: %w", err)
		}

		// Sync each calendar
		for _, cal := range calendars {
			if cal.AccessRole == "reader" || cal.AccessRole == "writer" || cal.AccessRole == "owner" {
				eventsSynced, err := h.syncCalendarEvents(payload.UserID, payload.FamilyID, cal.ID, timeMin, timeMax)
				if err != nil {
					log.Printf("Failed to sync calendar %s: %v", cal.ID, err)
					continue
				}
				totalEventsSynced += eventsSynced
			}
		}
	} else {
		// Sync specific calendar
		eventsSynced, err := h.syncCalendarEvents(payload.UserID, payload.FamilyID, payload.CalendarID, timeMin, timeMax)
		if err != nil {
			if updateErr := h.updateSyncStatus(payload.UserID, "error", fmt.Sprintf("Failed to sync calendar: %v", err), 0); updateErr != nil {
				log.Printf("Failed to update sync status: %v", updateErr)
			}
			return fmt.Errorf("failed to sync calendar events: %w", err)
		}
		totalEventsSynced = eventsSynced
	}

	// Update sync status to success
	if err := h.updateSyncStatus(payload.UserID, "success", "", totalEventsSynced); err != nil {
		log.Printf("Failed to update sync status: %v", err)
	}

	log.Printf("Calendar sync completed for user %s. Synced %d events", payload.UserID, totalEventsSynced)
	return nil
}

// syncCalendarEvents syncs events from a specific calendar
func (h *CalendarSyncHandler) syncCalendarEvents(userID, familyID, calendarID string, timeMin, timeMax time.Time) (int, error) {
	// Get events from Google Calendar
	events, err := h.googleClient.GetEvents(userID, calendarID, timeMin, timeMax)
	if err != nil {
		return 0, fmt.Errorf("failed to get events: %w", err)
	}

	eventsSynced := 0

	// Process each event
	for _, event := range events {
		// Skip cancelled events
		if event.Status == "cancelled" {
			continue
		}

		// Convert Google event to our calendar event format
		calEvent, err := h.convertGoogleEvent(event, familyID, userID)
		if err != nil {
			log.Printf("Failed to convert event %s: %v", event.ID, err)
			continue
		}

		// Insert or update event in database
		if err := h.upsertCalendarEvent(calEvent); err != nil {
			log.Printf("Failed to upsert event %s: %v", event.ID, err)
			continue
		}

		eventsSynced++
	}

	return eventsSynced, nil
}

// convertGoogleEvent converts a Google Calendar event to our internal format
func (h *CalendarSyncHandler) convertGoogleEvent(googleEvent calendar.GoogleEvent, familyID, userID string) (*CalendarEvent, error) {
	// Parse start time
	startTime, err := h.parseGoogleDateTime(googleEvent.Start)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	// Parse end time
	endTime, err := h.parseGoogleDateTime(googleEvent.End)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	// Convert attendees
	var attendees []string
	for _, attendee := range googleEvent.Attendees {
		attendees = append(attendees, attendee.Email)
	}

	// Determine if this is a recurring event
	isRecurring := len(googleEvent.Recurrence) > 0
	isRecurringInstance := googleEvent.RecurringEventId != ""

	return &CalendarEvent{
		ID:                  googleEvent.ID,
		FamilyID:            familyID,
		CreatedBy:           userID,
		Title:               googleEvent.Summary,
		Description:         googleEvent.Description,
		Location:            googleEvent.Location,
		StartTime:           startTime,
		EndTime:             &endTime,
		AllDay:              googleEvent.Start.Date != "", // All-day if date instead of dateTime
		Attendees:           attendees,
		SourceType:          "google",
		SourceID:            googleEvent.ID,
		IsRecurring:         isRecurring,
		RecurrenceRules:     googleEvent.Recurrence,
		RecurringEventID:    googleEvent.RecurringEventId,
		IsRecurringInstance: isRecurringInstance,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}

// parseGoogleDateTime parses Google Calendar datetime format
func (h *CalendarSyncHandler) parseGoogleDateTime(dt calendar.GoogleDateTime) (time.Time, error) {
	if dt.DateTime != "" {
		return time.Parse(time.RFC3339, dt.DateTime)
	}
	if dt.Date != "" {
		return time.Parse("2006-01-02", dt.Date)
	}
	return time.Time{}, fmt.Errorf("no datetime or date found")
}

// CalendarEvent represents our internal calendar event structure
type CalendarEvent struct {
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
	// Recurring event fields
	IsRecurring         bool      `json:"is_recurring"`
	RecurrenceRules     []string  `json:"recurrence_rules,omitempty"`
	RecurringEventID    string    `json:"recurring_event_id,omitempty"`
	IsRecurringInstance bool      `json:"is_recurring_instance"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// upsertCalendarEvent inserts or updates a calendar event
func (h *CalendarSyncHandler) upsertCalendarEvent(event *CalendarEvent) error {
	attendeesJSON, err := json.Marshal(event.Attendees)
	if err != nil {
		return fmt.Errorf("failed to marshal attendees: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO calendar_events
		(id, family_id, created_by, title, description, location, start_time, end_time,
		 all_day, attendees, source_type, source_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = h.db.Exec(query,
		event.ID, event.FamilyID, event.CreatedBy, event.Title, event.Description,
		event.Location, event.StartTime, event.EndTime, event.AllDay,
		string(attendeesJSON), event.SourceType, event.SourceID,
		event.CreatedAt, event.UpdatedAt,
	)

	return err
}

// getSyncSettings retrieves sync settings for a user
func (h *CalendarSyncHandler) getSyncSettings(userID string) (*SyncSettings, error) {
	query := `
		SELECT sync_frequency_minutes, sync_range_days
		FROM calendar_sync_settings
		WHERE user_id = ?
	`

	var settings SyncSettings
	err := h.db.QueryRow(query, userID).Scan(&settings.SyncFrequencyMinutes, &settings.SyncRangeDays)
	if err != nil {
		// Return default settings if not found
		return &SyncSettings{
			SyncFrequencyMinutes: 15,
			SyncRangeDays:        30,
		}, nil
	}

	return &settings, nil
}

// updateSyncStatus updates the sync status for a user
func (h *CalendarSyncHandler) updateSyncStatus(userID, status, errorMsg string, eventsSynced int) error {
	query := `
		INSERT OR REPLACE INTO calendar_sync_settings
		(user_id, last_sync_at, sync_status, sync_error, events_synced, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := h.db.Exec(query, userID, time.Now(), status, errorMsg, eventsSynced, time.Now())
	return err
}

// SyncSettings represents calendar sync configuration
type SyncSettings struct {
	SyncFrequencyMinutes int `json:"sync_frequency_minutes"`
	SyncRangeDays        int `json:"sync_range_days"`
}

// ScheduleCalendarSync schedules calendar sync jobs for all users with OAuth tokens
func ScheduleCalendarSync(js *jobsystem.SQLiteJobSystem, db *database.DB) error {
	// Get all users with Google OAuth tokens
	query := `
		SELECT DISTINCT user_id, family_id
		FROM oauth_tokens
		WHERE provider = 'google' AND expires_at > datetime('now')
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query OAuth tokens: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID, familyID string
		if err := rows.Scan(&userID, &familyID); err != nil {
			log.Printf("Failed to scan OAuth token row: %v", err)
			continue
		}

		// Create sync job payload
		payload := map[string]interface{}{
			"userId":   userID,
			"familyId": familyID,
			"provider": "google",
		}

		// Schedule the job
		req := &jobsystem.EnqueueRequest{
			QueueName:  "calendar-sync",
			JobType:    CalendarSyncJobType,
			Payload:    payload,
			Priority:   1,
			MaxRetries: 3,
		}

		_, err := js.Enqueue(req)
		if err != nil {
			log.Printf("Failed to enqueue calendar sync job for user %s: %v", userID, err)
		}
	}

	return nil
}
