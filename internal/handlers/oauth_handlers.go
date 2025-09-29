package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"famstack/internal/auth"
	"famstack/internal/integrations"
	"famstack/internal/jobsystem"
	"famstack/internal/oauth"
	"famstack/internal/services"
)

// OAuthHandlers handles OAuth-related HTTP requests
type OAuthHandlers struct {
	oauthService        *oauth.Service
	authService         *auth.Service
	jobSystem           *jobsystem.DBJobSystem
	integrationsService *services.IntegrationsService
}

// NewOAuthHandlers creates new OAuth handlers
func NewOAuthHandlers(oauthService *oauth.Service, authService *auth.Service, jobSystem *jobsystem.DBJobSystem, integrationsService *services.IntegrationsService) *OAuthHandlers {
	return &OAuthHandlers{
		oauthService:        oauthService,
		authService:         authService,
		jobSystem:           jobSystem,
		integrationsService: integrationsService,
	}
}

// HandleGoogleConnect shows configuration form before OAuth
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

	// Show Google Calendar configuration form
	data := map[string]any{
		"User":          user,
		"Provider":      "Google Calendar",
		"DefaultConfig": integrations.DefaultCalendarSyncConfig(),
	}

	RenderTemplate(w, "oauth-configure", data)
}

// HandleGoogleConnectWithConfig processes configuration and starts OAuth
func (h *OAuthHandlers) HandleGoogleConnectWithConfig(w http.ResponseWriter, r *http.Request) {
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

	// Parse configuration from form
	config, err := h.parseCalendarConfig(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid configuration: %v", err), http.StatusBadRequest)
		return
	}

	// Encode config into OAuth state parameter
	encodedState, err := h.encodeConfigInState(user.ID, config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode configuration: %v", err), http.StatusInternalServerError)
		return
	}

	// Use our encoded state instead of just userID
	authURL, err := h.oauthService.GetAuthURLWithCustomState(oauth.ProviderGoogle, encodedState)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate auth URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to Google OAuth
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles Google OAuth callback
func (h *OAuthHandlers) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🔄 OAuth callback started - URL: %s\n", r.URL.String())

	if r.Method != "GET" {
		fmt.Printf("❌ OAuth callback failed: invalid method %s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authorization code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	fmt.Printf("📋 OAuth callback received\n")

	if errorParam != "" {
		fmt.Printf("🔒 Security: OAuth denied or cancelled by user\n")
		http.Redirect(w, r, "/integrations?error=oauth_denied", http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		fmt.Printf("🔒 Security: Invalid OAuth callback attempt from %s\n", r.RemoteAddr)
		http.Redirect(w, r, "/integrations?error=invalid_callback", http.StatusTemporaryRedirect)
		return
	}

	// Get user ID and config from custom state BEFORE processing callback
	fmt.Printf("👤 Decoding user configuration...\n")
	userID, config, err := h.decodeConfigFromState(state)
	if err != nil {
		fmt.Printf("⚠️ Failed to decode config, using fallback\n")
		// Fallback: try to get userID the old way and use default config
		fallbackUserID, fallbackErr := h.oauthService.GetUserIDFromState(state)
		if fallbackErr != nil {
			fmt.Printf("❌ State validation failed\n")
			http.Redirect(w, r, "/integrations?error=invalid_state", http.StatusTemporaryRedirect)
			return
		}
		userID = fallbackUserID
		config = integrations.DefaultCalendarSyncConfig()
		fmt.Printf("✅ Using fallback configuration\n")
	} else {
		fmt.Printf("✅ Successfully decoded configuration\n")
	}

	// Process OAuth callback
	fmt.Printf("🔑 Processing OAuth token exchange...\n")
	token, err := h.oauthService.HandleCallback(oauth.ProviderGoogle, code, state)
	if err != nil {
		fmt.Printf("❌ OAuth token exchange failed\n")
		http.Redirect(w, r, "/integrations?error=callback_failed", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("✅ OAuth token received successfully\n")

	// Get user's family ID
	fmt.Printf("👨‍👩‍👧‍👦 Getting user family info...\n")
	user, err := h.authService.GetFamilyMemberByID(userID)
	if err != nil {
		fmt.Printf("❌ Failed to get user family info\n")
		http.Redirect(w, r, "/integrations?error=user_not_found", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("✅ User family retrieved\n")

	// Convert config struct to map[string]any for storage
	configMap := make(map[string]any)
	configBytes, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("❌ Failed to marshal config\n")
		http.Redirect(w, r, "/integrations?error=config_failed", http.StatusTemporaryRedirect)
		return
	}
	if unmarshalErr := json.Unmarshal(configBytes, &configMap); unmarshalErr != nil {
		fmt.Printf("❌ Failed to unmarshal config to map\n")
		http.Redirect(w, r, "/integrations?error=config_failed", http.StatusTemporaryRedirect)
		return
	}

	// Create integration record with user's configuration
	fmt.Printf("🔗 Creating integration record with config...\n")
	integrationReq := &services.CreateIntegrationRequest{
		IntegrationType: services.TypeCalendar,
		Provider:        services.ProviderGoogle,
		AuthMethod:      services.AuthOAuth2,
		DisplayName:     "Google Calendar",
		Description:     "Sync Google Calendar events",
		Settings:        configMap,
		SettingsType:    "CalendarSyncConfig",
	}

	createdIntegration, err := h.integrationsService.CreateIntegration(user.FamilyID, userID, integrationReq)
	if err != nil {
		fmt.Printf("❌ Failed to create integration\n")
		http.Redirect(w, r, "/integrations?error=integration_failed", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("✅ Integration created successfully\n")

	// Store OAuth credentials for the integration
	fmt.Printf("🔐 Storing OAuth credentials...\n")
	err = h.integrationsService.StoreOAuthCredentials(
		createdIntegration.ID,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Scope,
		&token.ExpiresAt,
	)
	if err != nil {
		fmt.Printf("❌ Failed to store OAuth credentials\n")
		// Integration created but credentials failed - continue with error in URL
		http.Redirect(w, r, "/integrations?error=credentials_failed", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("✅ OAuth credentials stored successfully\n")

	// Success - redirect to integrations page
	fmt.Printf("🎉 OAuth flow completed successfully! Redirecting to integrations page\n")
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

	providerStr := pathParts[4]

	// Validate provider against allowed values
	var provider oauth.Provider
	switch providerStr {
	case "google":
		provider = oauth.ProviderGoogle
	default:
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}

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
	data := map[string]any{
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
	payload := map[string]any{
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
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// parseCalendarConfig parses calendar configuration from form data
func (h *OAuthHandlers) parseCalendarConfig(r *http.Request) (*integrations.CalendarSyncConfig, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	config := integrations.DefaultCalendarSyncConfig()

	// Parse sync frequency (in minutes) with bounds checking
	if freqStr := r.FormValue("sync_frequency_minutes"); freqStr != "" {
		if freq, err := strconv.Atoi(freqStr); err == nil {
			// Validate bounds: 15 minutes to 24 hours (1440 minutes)
			if freq >= 15 && freq <= 1440 {
				config.SyncFrequencyMinutes = freq
			} else {
				return nil, fmt.Errorf("sync frequency must be between 15 and 1440 minutes, got %d", freq)
			}
		} else {
			return nil, fmt.Errorf("invalid sync frequency format: %v", err)
		}
	}

	// Parse sync range (in days) with bounds checking
	if rangeStr := r.FormValue("sync_range_days"); rangeStr != "" {
		if rangeDays, err := strconv.Atoi(rangeStr); err == nil {
			// Validate bounds: 1 day to 365 days (1 year)
			if rangeDays >= 1 && rangeDays <= 365 {
				config.SyncRangeDays = rangeDays
			} else {
				return nil, fmt.Errorf("sync range must be between 1 and 365 days, got %d", rangeDays)
			}
		} else {
			return nil, fmt.Errorf("invalid sync range format: %v", err)
		}
	}

	// Parse boolean preferences
	config.SyncAllDayEvents = r.FormValue("sync_all_day_events") == "on"
	config.SyncPrivateEvents = r.FormValue("sync_private_events") == "on"
	config.SyncDeclinedEvents = r.FormValue("sync_declined_events") == "on"

	// Parse calendars to sync (multiple select) with validation
	calendarsToSync := r.Form["calendars_to_sync"]
	if len(calendarsToSync) > 10 {
		return nil, fmt.Errorf("too many calendars selected, maximum 10 allowed, got %d", len(calendarsToSync))
	}

	// Validate each calendar name
	for _, calendar := range calendarsToSync {
		if len(calendar) == 0 || len(calendar) > 100 {
			return nil, fmt.Errorf("invalid calendar name length, must be 1-100 characters")
		}
		// Basic sanitization - allow alphanumeric, spaces, hyphens, underscores
		for _, char := range calendar {
			if (char < 'a' || char > 'z') &&
				(char < 'A' || char > 'Z') &&
				(char < '0' || char > '9') &&
				char != ' ' && char != '-' && char != '_' && char != '.' {
				return nil, fmt.Errorf("invalid character in calendar name: %c", char)
			}
		}
	}
	config.CalendarsToSync = calendarsToSync

	// Validate configuration (this will auto-correct some values)
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// encodeConfigInState encodes config and userID into OAuth state parameter
func (h *OAuthHandlers) encodeConfigInState(userID string, config *integrations.CalendarSyncConfig) (string, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create state payload with user ID and config
	stateData := map[string]string{
		"user_id": userID,
		"config":  string(configBytes),
	}

	stateBytes, err := json.Marshal(stateData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state data: %w", err)
	}

	// Base64 encode for safe URL transmission
	encodedState := base64.URLEncoding.EncodeToString(stateBytes)
	return encodedState, nil
}

// decodeConfigFromState decodes config and userID from OAuth state parameter
func (h *OAuthHandlers) decodeConfigFromState(state string) (string, *integrations.CalendarSyncConfig, error) {
	// Base64 decode
	stateBytes, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode state: %w", err)
	}

	// Unmarshal state data
	var stateData map[string]string
	if err := json.Unmarshal(stateBytes, &stateData); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal state data: %w", err)
	}

	userID, exists := stateData["user_id"]
	if !exists {
		return "", nil, fmt.Errorf("user_id not found in state")
	}

	configStr, exists := stateData["config"]
	if !exists {
		return "", nil, fmt.Errorf("config not found in state")
	}

	// Unmarshal config
	var config integrations.CalendarSyncConfig
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return userID, &config, nil
}

// RenderTemplate is a placeholder for template rendering
// TODO: Implement proper template rendering
func RenderTemplate(w http.ResponseWriter, templateName string, data any) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<!-- Template: %s -->\n<h1>Calendar Settings</h1>\n<p>OAuth integration page (template rendering not implemented)</p>", templateName)
}
