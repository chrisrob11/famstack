package api

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"

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
