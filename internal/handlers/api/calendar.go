package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

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
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	date := r.URL.Query().Get("date")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	familyID := r.URL.Query().Get("family_id")

	// Default to current family if not specified
	if familyID == "" {
		familyID = "fam1" // Default family for now
	}

	var startDate, endDate time.Time

	if date != "" {
		// Single date query - use same date for start and end
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		startDate = parsedDate
		endDate = parsedDate.Add(24 * time.Hour)
	} else if startDateStr != "" && endDateStr != "" {
		// Date range query
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}
		endDate = endDate.Add(24 * time.Hour) // Include the end date
	} else {
		// Default to today's events
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.Add(24 * time.Hour)
	}

	// Use the service to get events
	events, err := h.calendarService.GetUnifiedCalendarEvents(familyID, startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
		return
	}

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
		eventData.FamilyID = "fam1" // Default family
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
