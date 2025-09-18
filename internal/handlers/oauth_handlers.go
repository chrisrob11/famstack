package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"famstack/internal/auth"
	"famstack/internal/integrations"
	"famstack/internal/jobsystem"
	"famstack/internal/oauth"
)

// OAuthHandlers handles OAuth-related HTTP requests
type OAuthHandlers struct {
	oauthService        *oauth.Service
	authService         *auth.Service
	jobSystem           *jobsystem.SQLiteJobSystem
	integrationsService *integrations.Service
}

// NewOAuthHandlers creates new OAuth handlers
func NewOAuthHandlers(oauthService *oauth.Service, authService *auth.Service, jobSystem *jobsystem.SQLiteJobSystem, integrationsService *integrations.Service) *OAuthHandlers {
	return &OAuthHandlers{
		oauthService:        oauthService,
		authService:         authService,
		jobSystem:           jobSystem,
		integrationsService: integrationsService,
	}
}

// HandleGoogleConnect initiates Google OAuth flow
func (h *OAuthHandlers) HandleGoogleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Generate OAuth URL
	authURL, err := h.oauthService.GetAuthURL(oauth.ProviderGoogle, user.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate auth URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to Google OAuth
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles Google OAuth callback
func (h *OAuthHandlers) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authorization code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		http.Redirect(w, r, "/integrations?error=oauth_denied", http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		http.Redirect(w, r, "/integrations?error=invalid_callback", http.StatusTemporaryRedirect)
		return
	}

	// Process OAuth callback
	token, err := h.oauthService.HandleCallback(oauth.ProviderGoogle, code, state)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/integrations?error=callback_failed&details=%s", err.Error()), http.StatusTemporaryRedirect)
		return
	}

	// Get user ID from state (OAuth service should decode this)
	userID, err := h.oauthService.GetUserIDFromState(state)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/integrations?error=invalid_state&details=%s", err.Error()), http.StatusTemporaryRedirect)
		return
	}

	// Get user's family ID
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/integrations?error=user_not_found&details=%s", err.Error()), http.StatusTemporaryRedirect)
		return
	}

	// Create integration record
	integrationReq := &integrations.CreateIntegrationRequest{
		IntegrationType: integrations.TypeCalendar,
		Provider:        integrations.ProviderGoogle,
		AuthMethod:      integrations.AuthOAuth2,
		DisplayName:     "Google Calendar",
		Description:     "Sync Google Calendar events",
		Settings: map[string]any{
			"sync_calendars": []string{"primary"},
			"sync_interval":  "15m",
		},
	}

	createdIntegration, err := h.integrationsService.CreateIntegration(user.FamilyID, userID, integrationReq)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/integrations?error=integration_failed&details=%s", err.Error()), http.StatusTemporaryRedirect)
		return
	}

	// Store OAuth credentials for the integration
	err = h.integrationsService.StoreOAuthCredentials(
		createdIntegration.ID,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Scope,
		&token.ExpiresAt,
	)
	if err != nil {
		// Integration created but credentials failed - log error but continue
		fmt.Printf("Warning: Failed to store OAuth credentials for integration %s: %v\n", createdIntegration.ID, err)
	}

	// Success - redirect to integrations page
	http.Redirect(w, r, "/integrations?success=google_connected", http.StatusTemporaryRedirect)
}

// HandleDisconnectProvider disconnects OAuth provider
func (h *OAuthHandlers) HandleDisconnectProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Extract provider from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	provider := oauth.Provider(pathParts[4])

	// Delete OAuth token for this provider
	err := h.oauthService.DeleteToken(user.ID, provider)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to disconnect %s: %v", provider, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"message":  fmt.Sprintf("%s disconnected successfully", provider),
		"provider": string(provider),
		"user_id":  user.ID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleCalendarSettings displays calendar settings page
func (h *OAuthHandlers) HandleCalendarSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Check if Google is connected
	googleConnected := false
	if _, err := h.oauthService.GetToken(user.ID, oauth.ProviderGoogle); err == nil {
		googleConnected = true
	}

	// Prepare template data
	data := map[string]interface{}{
		"User":            user,
		"GoogleConnected": googleConnected,
		"LastSync":        nil, // TODO: Get from database
		"EventsSynced":    0,   // TODO: Get from database
		"SyncStatus":      "Never synced",
		"SyncStatusClass": "sync-pending",
	}

	// Render template
	RenderTemplate(w, "calendar-settings", data)
}

// HandleSyncNow triggers immediate calendar sync
func (h *OAuthHandlers) HandleSyncNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from context
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Check if user has Google OAuth connected
	_, err := h.oauthService.GetToken(user.ID, oauth.ProviderGoogle)
	if err != nil {
		http.Error(w, "Google Calendar not connected", http.StatusBadRequest)
		return
	}

	// Enqueue calendar sync job
	payload := map[string]interface{}{
		"user_id":    user.ID,
		"family_id":  user.FamilyID,
		"provider":   "google",
		"force_sync": true,
	}

	_, err = h.jobSystem.Enqueue(&jobsystem.EnqueueRequest{
		QueueName:  "calendar-sync",
		JobType:    "calendar_sync",
		Payload:    payload,
		Priority:   2, // Higher priority for manual sync
		MaxRetries: 3,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start sync: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Calendar sync started",
		"user_id": user.ID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// RenderTemplate is a placeholder for template rendering
// TODO: Implement proper template rendering
func RenderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<!-- Template: %s -->\n<h1>Calendar Settings</h1>\n<p>OAuth integration page (template rendering not implemented)</p>", templateName)
}
