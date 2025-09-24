package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"famstack/internal/models"
	"famstack/internal/services"
)

// FamilyAPIHandler handles family-related API requests
type FamilyAPIHandler struct {
	familiesService *services.FamiliesService
}

// NewFamilyAPIHandler creates a new family API handler
func NewFamilyAPIHandler(familiesService *services.FamiliesService) *FamilyAPIHandler {
	return &FamilyAPIHandler{
		familiesService: familiesService,
	}
}

// CreateFamily creates a new family
func (h *FamilyAPIHandler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON data
	var requestData struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Basic validation
	requestData.Name = strings.TrimSpace(requestData.Name)
	if requestData.Name == "" {
		http.Error(w, "Family name is required", http.StatusBadRequest)
		return
	}

	// Use the service to create the family
	family, err := h.familiesService.CreateFamily(requestData.Name)
	if err != nil {
		http.Error(w, "Failed to create family", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(family); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListFamilies lists all families
func (h *FamilyAPIHandler) ListFamilies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Use the service to list families
	families, err := h.familiesService.ListFamilies()
	if err != nil {
		http.Error(w, "Failed to query families", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(families); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetFamily retrieves a specific family by ID
func (h *FamilyAPIHandler) GetFamily(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract family ID from URL path
	familyID := path.Base(r.URL.Path)
	if familyID == "" || familyID == "/" {
		http.Error(w, "Family ID is required", http.StatusBadRequest)
		return
	}

	// Use the service to get the family
	family, err := h.familiesService.GetFamily(familyID)
	if err != nil {
		if err.Error() == "family not found" {
			http.Error(w, "Family not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query family", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(family); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateFamily updates a family's information
func (h *FamilyAPIHandler) UpdateFamily(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract family ID from URL path
	familyID := path.Base(r.URL.Path)
	if familyID == "" || familyID == "/" {
		http.Error(w, "Family ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Name     string `json:"name"`
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build update request with provided fields
	updateReq := &models.UpdateFamilyRequest{}

	if req.Name != "" {
		updateReq.Name = &req.Name
	}

	if req.Timezone != "" {
		updateReq.Timezone = &req.Timezone
	}

	// Validate that at least one field is provided
	if updateReq.Name == nil && updateReq.Timezone == nil {
		http.Error(w, "At least one field (name or timezone) is required", http.StatusBadRequest)
		return
	}
	family, err := h.familiesService.UpdateFamily(familyID, updateReq)
	if err != nil {
		if err.Error() == "family not found" {
			http.Error(w, "Family not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update family: %v", err), http.StatusInternalServerError)
		}
		return
	}

	h.writeJSON(w, family)
}

func (h *FamilyAPIHandler) writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Failed to encode JSON response: %v\n", err)
	}
}
