package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"famstack/internal/auth"
	"famstack/internal/models"
	"famstack/internal/services"
)

// CalendarAPIHandler handles calendar-related API requests
type CalendarAPIHandler struct {
	calendarService *services.CalendarService
}

// NewCalendarAPIHandler creates a new calendar API handler
func NewCalendarAPIHandler(calendarService *services.CalendarService) *CalendarAPIHandler {
	return &CalendarAPIHandler{
		calendarService: calendarService,
	}
}

// GetEvents retrieves unified calendar events for a specific date or date range
func (h *CalendarAPIHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🗓️  Calendar API called: %s\n", r.URL.String())

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	date := r.URL.Query().Get("date")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	familyID := r.URL.Query().Get("family_id")

	fmt.Printf("🗓️  Query: date=%s, start_date=%s, end_date=%s, family_id=%s\n", date, startDateStr, endDateStr, familyID)

	// Default to current family if not specified
	if familyID == "" {
		session := auth.GetSessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		familyID = session.FamilyID
	}

	var startDate, endDate time.Time

	if date != "" {
		// Single date query - use same date for start and end
		parsedDate, err := time.ParseInLocation("2006-01-02", date, time.UTC)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		startDate = parsedDate
		endDate = startDate.Add(24 * time.Hour)
	} else if startDateStr != "" && endDateStr != "" {
		// Date range query
		var err error
		startDate, err = time.ParseInLocation("2006-01-02", startDateStr, time.UTC)
		if err != nil {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
		endDate, err = time.ParseInLocation("2006-01-02", endDateStr, time.UTC)
		if err != nil {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}
		endDate = endDate.Add(24 * time.Hour) // Include the end date
	} else {
		// Default to today's events in UTC
		now := time.Now().UTC()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		endDate = startDate.Add(24 * time.Hour)
	}

	// Use the service to get events
	fmt.Printf("🗓️  Querying events for family %s from %s to %s\n", familyID, startDate.Format(time.RFC3339), endDate.Format(time.RFC3339))
	events, err := h.calendarService.GetUnifiedCalendarEvents(familyID, startDate, endDate)
	if err != nil {
		fmt.Printf("❌ Calendar query error: %v\n", err)
		// Return empty array instead of error to prevent frontend crashes
		events = []models.UnifiedCalendarEvent{}
	}

	// Ensure we never return nil, always return an empty array
	if events == nil {
		events = []models.UnifiedCalendarEvent{}
	}

	fmt.Printf("✅ Found %d events\n", len(events))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// CreateEvent creates a new unified calendar event
func (h *CalendarAPIHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var eventData models.CreateUnifiedCalendarEventRequest
	if err := json.NewDecoder(r.Body).Decode(&eventData); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Set defaults
	if eventData.FamilyID == "" {
		session := auth.GetSessionFromContext(r.Context())
		if session == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		eventData.FamilyID = session.FamilyID
	}
	// Note: EventType and Color are not part of CreateUnifiedCalendarEventRequest
	// This is for external integration events

	// Use the service to create the event
	event, err := h.calendarService.CreateUnifiedCalendarEvent(&eventData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create event: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateEvent updates a unified calendar event
func (h *CalendarAPIHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement UpdateUnifiedCalendarEvent in CalendarService
	// For now, return not implemented
	http.Error(w, "Update unified calendar event not yet implemented", http.StatusNotImplemented)
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

	// Use the service to get the event
	event, err := h.calendarService.GetUnifiedCalendarEvent(eventID)
	if err != nil {
		if err.Error() == "event not found" {
			http.Error(w, "Event not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query event", http.StatusInternalServerError)
		}
		return
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

	// Use the service to delete the event
	err := h.calendarService.DeleteEvent(eventID)
	if err != nil {
		if err.Error() == "event not found" {
			http.Error(w, "Event not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete event: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetCalendarDays retrieves multi-day calendar data with layered layout
func (h *CalendarAPIHandler) GetCalendarDays(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🗓️  Calendar Days API called: %s\n", r.URL.String())

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	peopleParam := r.URL.Query().Get("people")
	timezoneParam := r.URL.Query().Get("timezone")

	// Validate required parameters
	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "startDate and endDate are required", http.StatusBadRequest)
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid startDate format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid endDate format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Validate date range (max 31 days)
	daysDiff := endDate.Sub(startDate).Hours() / 24
	if daysDiff < 0 {
		http.Error(w, "endDate must be after startDate", http.StatusBadRequest)
		return
	}
	if daysDiff > 31 {
		http.Error(w, "Date range cannot exceed 31 days", http.StatusBadRequest)
		return
	}

	// Parse people filter
	var requestedPeople []string
	if peopleParam != "" {
		requestedPeople = strings.Split(peopleParam, ",")
		// Trim whitespace
		for i, person := range requestedPeople {
			requestedPeople[i] = strings.TrimSpace(person)
		}
	}

	// Get family ID from session
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	familyID := session.FamilyID

	// Set timezone (default to family timezone or UTC)
	timezone := "UTC"
	if timezoneParam != "" {
		timezone = timezoneParam
	}

	fmt.Printf("🗓️  Querying layered calendar: family=%s, start=%s, end=%s, people=%v, timezone=%s\n",
		familyID, startDateStr, endDateStr, requestedPeople, timezone)

	// Get events using existing service
	events, err := h.calendarService.GetUnifiedCalendarEvents(familyID, startDate, endDate.Add(24*time.Hour))
	if err != nil {
		fmt.Printf("❌ Calendar days query error: %v\n", err)
		events = []models.UnifiedCalendarEvent{}
	}

	// Filter events by people if specified
	if len(requestedPeople) > 0 {
		events = h.filterEventsByPeople(events, requestedPeople)
	}

	// Convert to layered format
	response := h.convertToLayeredResponse(events, startDate, endDate, requestedPeople, timezone)

	fmt.Printf("✅ Returning %d days with %d total events\n", len(response.Days), response.Metadata.TotalEvents)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// filterEventsByPeople filters events to only include those involving the specified people
func (h *CalendarAPIHandler) filterEventsByPeople(events []models.UnifiedCalendarEvent, requestedPeople []string) []models.UnifiedCalendarEvent {
	if len(requestedPeople) == 0 {
		return events
	}

	var filtered []models.UnifiedCalendarEvent
	for _, event := range events {
		// Check if event owner is in requested people
		ownerMatches := false
		if event.CreatedBy != nil {
			for _, person := range requestedPeople {
				if *event.CreatedBy == person {
					ownerMatches = true
					break
				}
			}
		}

		// Check if any attendee is in requested people
		attendeeMatches := false
		for _, attendee := range event.Attendees {
			for _, person := range requestedPeople {
				if attendee.ID == person {
					attendeeMatches = true
					break
				}
			}
			if attendeeMatches {
				break
			}
		}

		if ownerMatches || attendeeMatches {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// convertToLayeredResponse converts unified events to layered calendar format
func (h *CalendarAPIHandler) convertToLayeredResponse(
	events []models.UnifiedCalendarEvent,
	startDate, endDate time.Time,
	requestedPeople []string,
	timezone string,
) models.DaysResponse {
	days := make([]models.DayView, 0)
	totalEvents := 0

	// Process each day in the range
	for d := startDate; !d.After(endDate); d = d.Add(24 * time.Hour) {
		dayStr := d.Format("2006-01-02")

		// Filter events for this day
		dayEvents := h.filterEventsForDay(events, d)

		// Convert to layered format
		layers := h.calculateEventLayers(dayEvents, timezone)

		dayView := models.DayView{
			Date:   dayStr,
			Layers: layers,
		}

		// Count events for metadata
		dayEventCount := 0
		for _, layer := range layers {
			dayEventCount += len(layer.Events)
		}

		days = append(days, dayView)
		totalEvents += dayEventCount
	}

	return models.DaysResponse{
		StartDate:       startDate.Format("2006-01-02"),
		EndDate:         endDate.Format("2006-01-02"),
		Timezone:        timezone,
		RequestedPeople: requestedPeople,
		Days:            days,
		Metadata: models.DaysResponseMetadata{
			TotalEvents:  totalEvents,
			LastUpdated:  time.Now(),
			MaxDaysLimit: 31,
		},
	}
}

// filterEventsForDay returns events that occur on the specified day
func (h *CalendarAPIHandler) filterEventsForDay(events []models.UnifiedCalendarEvent, day time.Time) []models.UnifiedCalendarEvent {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	var dayEvents []models.UnifiedCalendarEvent
	for _, event := range events {
		// Check if event overlaps with this day
		if event.StartTime.Before(dayEnd) && event.EndTime.After(dayStart) {
			dayEvents = append(dayEvents, event)
		}
	}

	return dayEvents
}

// calculateEventLayers implements the layer assignment algorithm
func (h *CalendarAPIHandler) calculateEventLayers(events []models.UnifiedCalendarEvent, timezone string) []models.CalendarLayer {
	if len(events) == 0 {
		return []models.CalendarLayer{}
	}

	// Convert events to slot-based format
	viewEvents := make([]models.CalendarViewEvent, 0, len(events))
	for _, event := range events {
		viewEvent := h.convertToViewEvent(event, timezone)
		viewEvents = append(viewEvents, viewEvent)
	}

	// First pass: assign events to layers
	layers := []models.CalendarLayer{}

	for i, event := range viewEvents {
		// Find the first layer where this event can fit
		layerIndex := h.findAvailableLayer(layers, event)

		// Ensure we have enough layers
		for len(layers) <= layerIndex {
			layers = append(layers, models.CalendarLayer{
				LayerIndex: len(layers),
				Events:     []models.CalendarViewEvent{},
			})
		}

		// Set initial overlap info (will be updated in second pass)
		viewEvents[i].OverlapGroup = 1
		viewEvents[i].OverlapIndex = layerIndex

		// Add event to the layer
		layers[layerIndex].Events = append(layers[layerIndex].Events, viewEvents[i])
	}

	// Second pass: calculate overlap groups based on actual layer usage
	h.calculateOverlapGroupsFromLayers(layers)

	return layers
}

// findAvailableLayer finds the first layer where an event can be placed without conflicts
func (h *CalendarAPIHandler) findAvailableLayer(layers []models.CalendarLayer, newEvent models.CalendarViewEvent) int {
	for layerIndex, layer := range layers {
		hasConflict := false

		for _, existingEvent := range layer.Events {
			if h.eventsOverlap(newEvent, existingEvent) {
				hasConflict = true
				break
			}
		}

		if !hasConflict {
			return layerIndex
		}
	}

	// No available layer found, create a new one
	return len(layers)
}

// eventsOverlap checks if two events overlap in time
func (h *CalendarAPIHandler) eventsOverlap(event1, event2 models.CalendarViewEvent) bool {
	return event1.StartSlot < event2.EndSlot && event1.EndSlot > event2.StartSlot
}

// calculateOverlapGroupsFromLayers determines overlap groups based on layer assignment
func (h *CalendarAPIHandler) calculateOverlapGroupsFromLayers(layers []models.CalendarLayer) {
	// Build a map of all events for easier lookup
	allEvents := make(map[string]models.CalendarViewEvent)
	eventToLayer := make(map[string]int)

	for layerIndex, layer := range layers {
		for _, event := range layer.Events {
			allEvents[event.ID] = event
			eventToLayer[event.ID] = layerIndex
		}
	}

	// For each event, find its "overlap group" - all events that are transitively connected
	for currentLayerIndex, currentLayer := range layers {
		for currentEventIndex, currentEvent := range currentLayer.Events {
			// Find all events that are transitively connected to this event
			connectedEvents := h.findConnectedEvents(currentEvent, allEvents)

			// Find all layers that contain connected events
			layersUsed := make(map[int]bool)
			for _, connectedEvent := range connectedEvents {
				layerIndex := eventToLayer[connectedEvent.ID]
				layersUsed[layerIndex] = true
			}

			// The overlap group size is the number of layers used by connected events
			overlapGroup := len(layersUsed)

			// Update the event
			layers[currentLayerIndex].Events[currentEventIndex].OverlapGroup = overlapGroup
			layers[currentLayerIndex].Events[currentEventIndex].OverlapIndex = currentLayerIndex

			// Debug logging (commented out for production)
			// fmt.Printf("🔍 Event '%s' (layer %d): overlapGroup=%d, overlapIndex=%d, layers=%v, connected=%d\n",
			//	currentEvent.Title, currentLayerIndex, overlapGroup, currentLayerIndex, layersUsed, len(connectedEvents))
		}
	}
}

// findConnectedEvents finds all events that are transitively connected through overlaps
func (h *CalendarAPIHandler) findConnectedEvents(startEvent models.CalendarViewEvent, allEvents map[string]models.CalendarViewEvent) []models.CalendarViewEvent {
	visited := make(map[string]bool)
	var result []models.CalendarViewEvent

	// DFS to find all connected events
	var dfs func(models.CalendarViewEvent)
	dfs = func(event models.CalendarViewEvent) {
		if visited[event.ID] {
			return
		}
		visited[event.ID] = true
		result = append(result, event)

		// Find all events that overlap with this one
		for _, otherEvent := range allEvents {
			if !visited[otherEvent.ID] && h.eventsOverlap(event, otherEvent) {
				dfs(otherEvent)
			}
		}
	}

	dfs(startEvent)
	return result
}

// convertToViewEvent converts a UnifiedCalendarEvent to CalendarViewEvent with slot calculation
func (h *CalendarAPIHandler) convertToViewEvent(event models.UnifiedCalendarEvent, timezone string) models.CalendarViewEvent {
	// Convert times to slots (15-minute intervals, 0-359)
	startSlot := h.timeToSlot(event.StartTime)
	endSlot := h.timeToSlot(event.EndTime)

	// Ensure endSlot is at least startSlot + 1
	if endSlot <= startSlot {
		endSlot = startSlot + 1
	}

	// Extract attendee IDs
	attendeeIDs := make([]string, len(event.Attendees))
	for i, attendee := range event.Attendees {
		attendeeIDs[i] = attendee.ID
	}

	// Determine owner ID
	ownerID := ""
	if event.CreatedBy != nil {
		ownerID = *event.CreatedBy
	}

	return models.CalendarViewEvent{
		ID:           event.ID,
		Title:        event.Title,
		StartSlot:    startSlot,
		EndSlot:      endSlot,
		Color:        event.Color,
		OwnerID:      ownerID,
		AttendeeIDs:  attendeeIDs,
		OverlapGroup: 1, // Default to 1, will be updated in calculateOverlapInfo
		OverlapIndex: 0, // Default to 0, will be updated in calculateOverlapInfo
		Attendees:    event.Attendees,
		IsPrivate:    false, // TODO: Implement privacy logic
		Location:     event.Location,
		Description:  event.Description,
	}
}

// timeToSlot converts a time to a slot number (0-359 for 24 hours in 15-minute intervals)
func (h *CalendarAPIHandler) timeToSlot(t time.Time) int {
	// Get minutes since midnight
	minutesSinceMidnight := t.Hour()*60 + t.Minute()

	// Convert to 15-minute slots
	slot := minutesSinceMidnight / 15

	// Ensure slot is within valid range (0-359)
	if slot < 0 {
		slot = 0
	}
	if slot >= 360 {
		slot = 359
	}

	return slot
}
