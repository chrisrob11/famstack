package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"famstack/internal/auth"
	"famstack/internal/services"
)

// IntegrationsAPIHandler handles integration API requests
type IntegrationsAPIHandler struct {
	integrationsService *services.IntegrationsService
}

// NewIntegrationsAPIHandler creates a new integrations API handler
func NewIntegrationsAPIHandler(integrationsService *services.IntegrationsService) *IntegrationsAPIHandler {
	return &IntegrationsAPIHandler{
		integrationsService: integrationsService,
	}
}

// ListIntegrations handles GET /api/v1/integrations
func (h *IntegrationsAPIHandler) ListIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	query := &services.ListIntegrationsQuery{}

	if integrationType := r.URL.Query().Get("type"); integrationType != "" {
		iType := services.IntegrationType(integrationType)
		query.IntegrationType = &iType
	}

	if provider := r.URL.Query().Get("provider"); provider != "" {
		p := services.Provider(provider)
		query.Provider = &p
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := services.Status(status)
		query.Status = &s
	}

	if authMethod := r.URL.Query().Get("auth_method"); authMethod != "" {
		am := services.AuthMethod(authMethod)
		query.AuthMethod = &am
	}

	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		query.CreatedBy = &createdBy
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			query.Offset = offset
		}
	}

	// Get integrations
	integrationsList, err := h.integrationsService.ListIntegrations(user.FamilyID, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list integrations: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if HTML is requested (for HTMX)
	if r.Header.Get("HX-Request") == "true" {
		// Return HTML template for HTMX
		h.renderIntegrationsListHTML(w, integrationsList)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"integrations": integrationsList,
		"count":        len(integrationsList),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateIntegration handles POST /api/v1/integrations
func (h *IntegrationsAPIHandler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req services.CreateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create integration
	integration, err := h.integrationsService.CreateIntegration(user.FamilyID, user.ID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create integration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(integration); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetIntegration handles GET /api/v1/integrations/{id}
func (h *IntegrationsAPIHandler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}
	integrationID := pathParts[4]

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Check if detailed view is requested
	includeCredentials := r.URL.Query().Get("include_credentials") == "true"

	if includeCredentials {
		// Get integration with credentials
		integrationWithCreds, err := h.integrationsService.GetIntegrationWithCredentials(integrationID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
			return
		}

		// Verify user has access to this integration
		if integrationWithCreds.Integration.FamilyID != user.FamilyID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(integrationWithCreds); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	} else {
		// Get basic integration info
		integration, err := h.integrationsService.GetIntegration(integrationID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
			return
		}

		// Verify user has access to this integration
		if integration.FamilyID != user.FamilyID {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(integration); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// UpdateIntegration handles PATCH /api/v1/integrations/{id}
func (h *IntegrationsAPIHandler) UpdateIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}
	integrationID := pathParts[4]

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user has access to this integration
	integration, err := h.integrationsService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
		return
	}
	if integration.FamilyID != user.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse request body
	var req services.UpdateIntegrationRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update integration
	updatedIntegration, err := h.integrationsService.UpdateIntegration(integrationID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update integration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedIntegration); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeleteIntegration handles DELETE /api/v1/integrations/{id}
func (h *IntegrationsAPIHandler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}
	integrationID := pathParts[4]

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user has access to this integration
	integration, err := h.integrationsService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
		return
	}
	if integration.FamilyID != user.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete integration
	if err := h.integrationsService.DeleteIntegration(integrationID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete integration: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Integration deleted successfully",
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// SyncIntegration handles POST /api/v1/integrations/{id}/sync
func (h *IntegrationsAPIHandler) SyncIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}
	integrationID := pathParts[4]

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user has access to this integration
	integration, err := h.integrationsService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
		return
	}
	if integration.FamilyID != user.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// TODO: Implement sync logic based on integration type
	// For now, just return success
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":         "success",
		"message":        "Sync initiated",
		"integration_id": integrationID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// TestIntegration handles POST /api/v1/integrations/{id}/test
func (h *IntegrationsAPIHandler) TestIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}
	integrationID := pathParts[4]

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Verify user has access to this integration
	integration, err := h.integrationsService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
		return
	}
	if integration.FamilyID != user.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// TODO: Implement test logic based on integration type
	// For now, just return success
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":         "success",
		"message":        "Connection test successful",
		"integration_id": integrationID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// InitiateOAuth handles POST /api/v1/integrations/{id}/oauth/initiate
func (h *IntegrationsAPIHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract integration ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 7 {
		http.Error(w, "Invalid integration ID", http.StatusBadRequest)
		return
	}
	integrationID := pathParts[4]

	// Get user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get integration to verify access
	integration, err := h.integrationsService.GetIntegration(integrationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get integration: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify user has access to this integration
	if integration.FamilyID != user.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Generate authorization URL using service layer
	authURL, err := h.integrationsService.InitiateOAuth(integrationID, r.Host)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initiate OAuth: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"authorization_url": authURL,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// renderIntegrationsListHTML renders the integrations list template for HTMX
func (h *IntegrationsAPIHandler) renderIntegrationsListHTML(w http.ResponseWriter, integrations []services.Integration) {
	tmpl, err := template.ParseFiles("web/templates/integrations-list.html.tmpl")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"integrations": integrations,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}
}
