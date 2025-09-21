package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"famstack/internal/auth"
	"famstack/internal/jobsystem"
	"famstack/internal/models"
	"famstack/internal/services"
)

type ScheduleHandler struct {
	schedulesService *services.SchedulesService
	jobSystem        *jobsystem.DBJobSystem
}

func NewScheduleHandler(schedulesService *services.SchedulesService) *ScheduleHandler {
	return &ScheduleHandler{
		schedulesService: schedulesService,
	}
}

func NewScheduleHandlerWithJobSystem(schedulesService *services.SchedulesService, jobSystem *jobsystem.DBJobSystem) *ScheduleHandler {
	return &ScheduleHandler{
		schedulesService: schedulesService,
		jobSystem:        jobSystem,
	}
}

// ListSchedules returns all active task schedules for a family
func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get family ID from session context
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	familyID := session.FamilyID

	// Use the service to get schedules
	schedules, err := h.schedulesService.ListSchedules(familyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query schedules: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedules); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// CreateSchedule creates a new task schedule
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTaskScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Get family ID and user ID from session context
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	familyID := session.FamilyID
	createdBy := session.UserID

	// Use the service to create the schedule
	schedule, err := h.schedulesService.CreateSchedule(familyID, createdBy, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create schedule: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(schedule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetSchedule retrieves a specific task schedule
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract schedule ID from URL path
	scheduleID := path.Base(r.URL.Path)
	if scheduleID == "" || scheduleID == "/" {
		http.Error(w, "Schedule ID is required", http.StatusBadRequest)
		return
	}

	// Use the service to get the schedule
	schedule, err := h.schedulesService.GetSchedule(scheduleID)
	if err != nil {
		if err.Error() == "schedule not found" {
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query schedule", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schedule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// UpdateSchedule updates a task schedule
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract schedule ID from URL path
	scheduleID := path.Base(r.URL.Path)
	if scheduleID == "" || scheduleID == "/" {
		http.Error(w, "Schedule ID is required", http.StatusBadRequest)
		return
	}

	// Get session to check authorization
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the schedule to check ownership
	schedule, getErr := h.schedulesService.GetSchedule(scheduleID)
	if getErr != nil {
		if getErr.Error() == "schedule not found" {
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query schedule", http.StatusInternalServerError)
		}
		return
	}

	// Check if user can update this schedule:
	// 1. User is an admin (has admin role)
	// 2. User is the creator of the schedule
	canUpdate := session.Role == auth.RoleAdmin || session.UserID == schedule.CreatedBy

	if !canUpdate {
		http.Error(w, "Insufficient permissions: only admins or schedule creators can update schedules", http.StatusForbidden)
		return
	}

	var req models.UpdateTaskScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Use the service to update the schedule
	updatedSchedule, err := h.schedulesService.UpdateSchedule(scheduleID, &req)
	if err != nil {
		if err.Error() == "schedule not found" {
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update schedule: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedSchedule); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteSchedule deletes a task schedule
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract schedule ID from URL path
	scheduleID := path.Base(r.URL.Path)
	if scheduleID == "" || scheduleID == "/" {
		http.Error(w, "Schedule ID is required", http.StatusBadRequest)
		return
	}

	// Get session to check authorization
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the schedule to check ownership
	schedule, getErr := h.schedulesService.GetSchedule(scheduleID)
	if getErr != nil {
		if getErr.Error() == "schedule not found" {
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query schedule", http.StatusInternalServerError)
		}
		return
	}

	// Check if user can delete this schedule:
	// 1. User is an admin (has admin role)
	// 2. User is the creator of the schedule
	canDelete := session.Role == auth.RoleAdmin || session.UserID == schedule.CreatedBy

	if !canDelete {
		http.Error(w, "Insufficient permissions: only admins or schedule creators can delete schedules", http.StatusForbidden)
		return
	}

	// Use the service to delete the schedule
	deleteErr := h.schedulesService.DeleteSchedule(scheduleID)
	if deleteErr != nil {
		if deleteErr.Error() == "schedule not found" {
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete schedule: %v", deleteErr), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
