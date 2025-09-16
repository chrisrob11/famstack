package calendar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"famstack/internal/oauth"
)

// GoogleClient handles Google Calendar API interactions
type GoogleClient struct {
	oauthService *oauth.Service
}

// NewGoogleClient creates a new Google Calendar client
func NewGoogleClient(oauthService *oauth.Service) *GoogleClient {
	return &GoogleClient{
		oauthService: oauthService,
	}
}

// GoogleEvent represents a Google Calendar event
type GoogleEvent struct {
	ID          string           `json:"id"`
	Summary     string           `json:"summary"`
	Description string           `json:"description"`
	Start       GoogleDateTime   `json:"start"`
	End         GoogleDateTime   `json:"end"`
	Attendees   []GoogleAttendee `json:"attendees,omitempty"`
	Location    string           `json:"location,omitempty"`
	Status      string           `json:"status"`
	Created     time.Time        `json:"created"`
	Updated     time.Time        `json:"updated"`
	// Recurring event fields
	Recurrence        []string        `json:"recurrence,omitempty"`        // RRULE, EXDATE, etc.
	RecurringEventId  string          `json:"recurringEventId,omitempty"`  // ID of the master recurring event
	OriginalStartTime *GoogleDateTime `json:"originalStartTime,omitempty"` // For modified instances
}

// GoogleDateTime represents Google Calendar date/time format
type GoogleDateTime struct {
	DateTime string `json:"dateTime,omitempty"`
	Date     string `json:"date,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

// GoogleAttendee represents a Google Calendar attendee
type GoogleAttendee struct {
	Email          string `json:"email"`
	DisplayName    string `json:"displayName,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
}

// GetEvents fetches events from Google Calendar
func (c *GoogleClient) GetEvents(userID string, calendarID string, timeMin, timeMax time.Time) ([]GoogleEvent, error) {
	// Get OAuth token for user
	token, err := c.oauthService.GetToken(userID, oauth.ProviderGoogle)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}

	// Create OAuth2 token source with auto-refresh
	oauth2Config := c.oauthService.GetOAuth2Config()
	oauth2Token := c.oauthService.GetOAuth2Token(token)
	tokenSource := oauth2Config.TokenSource(context.Background(), oauth2Token)

	// Create Calendar service
	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Build events list call
	eventsCall := calendarService.Events.List(calendarID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		MaxResults(2500).
		ShowDeleted(false).
		ShowHiddenInvitations(false)

	// Execute the request
	events, err := eventsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve events: %w", err)
	}

	// Convert to our custom format
	var googleEvents []GoogleEvent
	for _, item := range events.Items {
		googleEvent := GoogleEvent{
			ID:               item.Id,
			Summary:          item.Summary,
			Description:      item.Description,
			Location:         item.Location,
			Status:           item.Status,
			Recurrence:       item.Recurrence,
			RecurringEventId: item.RecurringEventId,
		}

		// Convert original start time for recurring event instances
		if item.OriginalStartTime != nil {
			googleEvent.OriginalStartTime = &GoogleDateTime{
				DateTime: item.OriginalStartTime.DateTime,
				Date:     item.OriginalStartTime.Date,
				TimeZone: item.OriginalStartTime.TimeZone,
			}
		}

		// Convert start time
		if item.Start != nil {
			googleEvent.Start = GoogleDateTime{
				DateTime: item.Start.DateTime,
				Date:     item.Start.Date,
				TimeZone: item.Start.TimeZone,
			}
		}

		// Convert end time
		if item.End != nil {
			googleEvent.End = GoogleDateTime{
				DateTime: item.End.DateTime,
				Date:     item.End.Date,
				TimeZone: item.End.TimeZone,
			}
		}

		// Convert attendees
		for _, attendee := range item.Attendees {
			googleEvent.Attendees = append(googleEvent.Attendees, GoogleAttendee{
				Email:          attendee.Email,
				DisplayName:    attendee.DisplayName,
				ResponseStatus: attendee.ResponseStatus,
			})
		}

		// Parse timestamps
		if item.Created != "" {
			if created, err := time.Parse(time.RFC3339, item.Created); err == nil {
				googleEvent.Created = created
			}
		}
		if item.Updated != "" {
			if updated, err := time.Parse(time.RFC3339, item.Updated); err == nil {
				googleEvent.Updated = updated
			}
		}

		googleEvents = append(googleEvents, googleEvent)
	}

	return googleEvents, nil
}

// GetCalendars fetches list of calendars for the user
func (c *GoogleClient) GetCalendars(userID string) ([]GoogleCalendar, error) {
	// Get OAuth token for user
	token, err := c.oauthService.GetToken(userID, oauth.ProviderGoogle)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}

	// Create OAuth2 token source with auto-refresh
	oauth2Config := c.oauthService.GetOAuth2Config()
	oauth2Token := c.oauthService.GetOAuth2Token(token)
	tokenSource := oauth2Config.TokenSource(context.Background(), oauth2Token)

	// Create Calendar service
	ctx := context.Background()
	calendarService, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Get calendar list
	calendarList, err := calendarService.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve calendar list: %w", err)
	}

	// Convert to our custom format
	var googleCalendars []GoogleCalendar
	for _, item := range calendarList.Items {
		googleCalendars = append(googleCalendars, GoogleCalendar{
			ID:          item.Id,
			Summary:     item.Summary,
			Description: item.Description,
			Primary:     item.Primary,
			AccessRole:  item.AccessRole,
			Selected:    item.Selected,
		})
	}

	return googleCalendars, nil
}

// GoogleCalendar represents a Google Calendar
type GoogleCalendar struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
	AccessRole  string `json:"accessRole"`
	Selected    bool   `json:"selected,omitempty"`
}
