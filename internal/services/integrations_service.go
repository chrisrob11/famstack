package services

import (
	"encoding/json"
	"fmt"
	"time"

	"famstack/internal/database"
	"famstack/internal/encryption"
)

// IntegrationType represents different categories of integrations
type IntegrationType string

const (
	TypeCalendar      IntegrationType = "calendar"
	TypeStorage       IntegrationType = "storage"
	TypeCommunication IntegrationType = "communication"
	TypeAutomation    IntegrationType = "automation"
	TypeSmartHome     IntegrationType = "smart_home"
	TypeFinance       IntegrationType = "finance"
	TypeHealth        IntegrationType = "health"
	TypeShopping      IntegrationType = "shopping"
)

// Provider represents the service provider
type Provider string

const (
	// Calendar providers
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
	ProviderApple     Provider = "apple"
	ProviderCalDAV    Provider = "caldav"

	// Storage providers
	ProviderDropbox     Provider = "dropbox"
	ProviderOneDrive    Provider = "onedrive"
	ProviderGoogleDrive Provider = "google_drive"
	ProviderICloud      Provider = "icloud"

	// Communication providers
	ProviderSlack   Provider = "slack"
	ProviderDiscord Provider = "discord"
	ProviderTeams   Provider = "teams"
	ProviderEmail   Provider = "email"

	// Automation providers
	ProviderIFTTT         Provider = "ifttt"
	ProviderZapier        Provider = "zapier"
	ProviderHomeAssistant Provider = "home_assistant"

	// Smart home providers
	ProviderHomeKit    Provider = "homekit"
	ProviderAlexa      Provider = "alexa"
	ProviderGoogleHome Provider = "google_home"
)

// AuthMethod represents how the integration authenticates
type AuthMethod string

const (
	AuthOAuth2    AuthMethod = "oauth2"
	AuthAPIKey    AuthMethod = "api_key"
	AuthBasicAuth AuthMethod = "basic_auth"
	AuthWebhook   AuthMethod = "webhook"
	AuthToken     AuthMethod = "token"
)

// Status represents the current state of an integration
type Status string

const (
	StatusConnected    Status = "connected"
	StatusDisconnected Status = "disconnected"
	StatusError        Status = "error"
	StatusPending      Status = "pending"
	StatusSyncing      Status = "syncing"
)

// Integration represents a single integration
type Integration struct {
	ID              string          `json:"id" db:"id"`
	FamilyID        string          `json:"family_id" db:"family_id"`
	CreatedBy       string          `json:"created_by" db:"created_by"`
	IntegrationType IntegrationType `json:"integration_type" db:"integration_type"`
	Provider        Provider        `json:"provider" db:"provider"`
	AuthMethod      AuthMethod      `json:"auth_method" db:"auth_method"`
	Status          Status          `json:"status" db:"status"`
	DisplayName     string          `json:"display_name" db:"display_name"`
	Description     string          `json:"description" db:"description"`
	Settings        string          `json:"settings" db:"settings"` // JSON
	LastSyncAt      *time.Time      `json:"last_sync_at" db:"last_sync_at"`
	LastError       *string         `json:"last_error" db:"last_error"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// OAuthCredentials represents OAuth2 credentials for an integration
type OAuthCredentials struct {
	ID            string     `json:"id" db:"id"`
	IntegrationID string     `json:"integration_id" db:"integration_id"`
	AccessToken   string     `json:"access_token" db:"access_token"`   // encrypted
	RefreshToken  string     `json:"refresh_token" db:"refresh_token"` // encrypted
	TokenType     string     `json:"token_type" db:"token_type"`
	ExpiresAt     *time.Time `json:"expires_at" db:"expires_at"`
	Scope         string     `json:"scope" db:"scope"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// APICredentials represents API keys or other auth credentials
type APICredentials struct {
	ID              string     `json:"id" db:"id"`
	IntegrationID   string     `json:"integration_id" db:"integration_id"`
	CredentialType  string     `json:"credential_type" db:"credential_type"`
	CredentialName  string     `json:"credential_name" db:"credential_name"`
	CredentialValue string     `json:"credential_value" db:"credential_value"` // encrypted
	ExpiresAt       *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// SyncHistory represents a sync operation history
type SyncHistory struct {
	ID            string     `json:"id" db:"id"`
	IntegrationID string     `json:"integration_id" db:"integration_id"`
	SyncType      string     `json:"sync_type" db:"sync_type"` // manual, scheduled, webhook
	Status        string     `json:"status" db:"status"`       // success, error, partial
	ItemsSynced   int        `json:"items_synced" db:"items_synced"`
	ErrorMessage  string     `json:"error_message" db:"error_message"`
	StartedAt     time.Time  `json:"started_at" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// IntegrationWithCredentials combines integration with its credentials
type IntegrationWithCredentials struct {
	Integration       *Integration      `json:"integration"`
	OAuthCreds        *OAuthCredentials `json:"oauth_credentials,omitempty"`
	APICreds          []APICredentials  `json:"api_credentials,omitempty"`
	RecentSyncHistory []SyncHistory     `json:"recent_sync_history,omitempty"`
}

// CreateIntegrationRequest represents a request to create a new integration
type CreateIntegrationRequest struct {
	IntegrationType IntegrationType `json:"integration_type" validate:"required"`
	Provider        Provider        `json:"provider" validate:"required"`
	AuthMethod      AuthMethod      `json:"auth_method" validate:"required"`
	DisplayName     string          `json:"display_name" validate:"required"`
	Description     string          `json:"description"`
	Settings        map[string]any  `json:"settings"`
}

// UpdateIntegrationRequest represents a request to update an integration
type UpdateIntegrationRequest struct {
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	Settings    map[string]any `json:"settings"`
	Status      Status         `json:"status"`
}

// ListIntegrationsQuery represents query parameters for listing integrations
type ListIntegrationsQuery struct {
	IntegrationType *IntegrationType `json:"integration_type"`
	Provider        *Provider        `json:"provider"`
	Status          *Status          `json:"status"`
	AuthMethod      *AuthMethod      `json:"auth_method"`
	CreatedBy       *string          `json:"created_by"`
	Limit           int              `json:"limit"`
	Offset          int              `json:"offset"`
}

// IntegrationsService handles integration operations
type IntegrationsService struct {
	db            *database.Fascade
	encryptionSvc *encryption.Service
}

// NewIntegrationsService creates a new integrations service
func NewIntegrationsService(db *database.Fascade, encryptionSvc *encryption.Service) *IntegrationsService {
	return &IntegrationsService{
		db:            db,
		encryptionSvc: encryptionSvc,
	}
}

// CreateIntegration creates a new integration
func (s *IntegrationsService) CreateIntegration(familyID, userID string, req *CreateIntegrationRequest) (*Integration, error) {
	settingsJSON := ""
	if req.Settings != nil {
		data, err := json.Marshal(req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}
		settingsJSON = string(data)
	}

	integration := &Integration{
		ID:              generateID(),
		FamilyID:        familyID,
		CreatedBy:       userID,
		IntegrationType: req.IntegrationType,
		Provider:        req.Provider,
		AuthMethod:      req.AuthMethod,
		Status:          StatusPending,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Settings:        settingsJSON,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	query := `
		INSERT INTO integrations
		(id, family_id, created_by, integration_type, provider, auth_method, status,
		 display_name, description, settings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		integration.ID, integration.FamilyID, integration.CreatedBy,
		integration.IntegrationType, integration.Provider, integration.AuthMethod,
		integration.Status, integration.DisplayName, integration.Description,
		integration.Settings, integration.CreatedAt, integration.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create integration: %w", err)
	}

	return integration, nil
}

// GetIntegration retrieves an integration by ID
func (s *IntegrationsService) GetIntegration(integrationID string) (*Integration, error) {
	query := `
		SELECT id, family_id, created_by, integration_type, provider, auth_method,
		       status, display_name, description, settings, last_sync_at,
		       last_error, created_at, updated_at
		FROM integrations
		WHERE id = ?
	`

	var integration Integration
	err := s.db.QueryRow(query, integrationID).Scan(
		&integration.ID, &integration.FamilyID, &integration.CreatedBy,
		&integration.IntegrationType, &integration.Provider, &integration.AuthMethod,
		&integration.Status, &integration.DisplayName, &integration.Description,
		&integration.Settings, &integration.LastSyncAt, &integration.LastError,
		&integration.CreatedAt, &integration.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get integration: %w", err)
	}

	return &integration, nil
}

// GetIntegrationWithCredentials retrieves an integration with its credentials
func (s *IntegrationsService) GetIntegrationWithCredentials(integrationID string) (*IntegrationWithCredentials, error) {
	integration, err := s.GetIntegration(integrationID)
	if err != nil {
		return nil, err
	}

	result := &IntegrationWithCredentials{
		Integration: integration,
	}

	// Get OAuth credentials if they exist
	if integration.AuthMethod == AuthOAuth2 {
		oauthCreds, oauthErr := s.getOAuthCredentials(integrationID)
		if oauthErr == nil {
			result.OAuthCreds = oauthCreds
		}
	}

	// Get API credentials
	apiCreds, err := s.getAPICredentials(integrationID)
	if err == nil {
		result.APICreds = apiCreds
	}

	// Get recent sync history
	syncHistory, err := s.getRecentSyncHistory(integrationID, 5)
	if err == nil {
		result.RecentSyncHistory = syncHistory
	}

	return result, nil
}

// ListIntegrations lists integrations with optional filters
func (s *IntegrationsService) ListIntegrations(familyID string, query *ListIntegrationsQuery) ([]Integration, error) {
	sql := `
		SELECT id, family_id, created_by, integration_type, provider, auth_method,
		       status, display_name, description, settings, last_sync_at,
		       last_error, created_at, updated_at
		FROM integrations
		WHERE family_id = ?
	`
	args := []any{familyID}

	// Add filters
	if query.IntegrationType != nil {
		sql += " AND integration_type = ?"
		args = append(args, *query.IntegrationType)
	}
	if query.Provider != nil {
		sql += " AND provider = ?"
		args = append(args, *query.Provider)
	}
	if query.Status != nil {
		sql += " AND status = ?"
		args = append(args, *query.Status)
	}
	if query.AuthMethod != nil {
		sql += " AND auth_method = ?"
		args = append(args, *query.AuthMethod)
	}
	if query.CreatedBy != nil {
		sql += " AND created_by = ?"
		args = append(args, *query.CreatedBy)
	}

	sql += " ORDER BY created_at DESC"

	// Add pagination
	if query.Limit > 0 {
		sql += " LIMIT ?"
		args = append(args, query.Limit)
		if query.Offset > 0 {
			sql += " OFFSET ?"
			args = append(args, query.Offset)
		}
	}

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list integrations: %w", err)
	}
	defer rows.Close()

	var integrations []Integration
	for rows.Next() {
		var integration Integration
		err := rows.Scan(
			&integration.ID, &integration.FamilyID, &integration.CreatedBy,
			&integration.IntegrationType, &integration.Provider, &integration.AuthMethod,
			&integration.Status, &integration.DisplayName, &integration.Description,
			&integration.Settings, &integration.LastSyncAt, &integration.LastError,
			&integration.CreatedAt, &integration.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan integration: %w", err)
		}
		integrations = append(integrations, integration)
	}

	return integrations, nil
}

// UpdateIntegration updates an integration
func (s *IntegrationsService) UpdateIntegration(integrationID string, req *UpdateIntegrationRequest) (*Integration, error) {
	// Get existing integration
	integration, err := s.GetIntegration(integrationID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.DisplayName != "" {
		integration.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		integration.Description = req.Description
	}
	if req.Settings != nil {
		data, marshalErr := json.Marshal(req.Settings)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", marshalErr)
		}
		integration.Settings = string(data)
	}
	if req.Status != "" {
		integration.Status = req.Status
	}
	integration.UpdatedAt = time.Now()

	query := `
		UPDATE integrations
		SET display_name = ?, description = ?, settings = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = s.db.Exec(query,
		integration.DisplayName, integration.Description, integration.Settings,
		integration.Status, integration.UpdatedAt, integration.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update integration: %w", err)
	}

	return integration, nil
}

// DeleteIntegration deletes an integration and all its credentials
func (s *IntegrationsService) DeleteIntegration(integrationID string) error {
	// Note: credentials will be deleted by CASCADE
	query := "DELETE FROM integrations WHERE id = ?"
	_, err := s.db.Exec(query, integrationID)
	if err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}
	return nil
}

// StoreOAuthCredentials stores encrypted OAuth credentials
func (s *IntegrationsService) StoreOAuthCredentials(integrationID string, accessToken, refreshToken, tokenType, scope string, expiresAt *time.Time) error {
	// Encrypt tokens
	encryptedAccessToken, err := s.encryptionSvc.Encrypt(accessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken string
	if refreshToken != "" {
		encryptedRefreshToken, err = s.encryptionSvc.Encrypt(refreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	query := `
		INSERT OR REPLACE INTO integration_oauth_credentials
		(id, integration_id, access_token, refresh_token, token_type, expires_at, scope, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err = s.db.Exec(query,
		generateID(), integrationID, encryptedAccessToken, encryptedRefreshToken,
		tokenType, expiresAt, scope, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to store OAuth credentials: %w", err)
	}

	// Update integration status to connected
	statusQuery := `UPDATE integrations SET status = ?, updated_at = ? WHERE id = ?`
	_, err = s.db.Exec(statusQuery, StatusConnected, now, integrationID)
	if err != nil {
		return fmt.Errorf("failed to update integration status: %w", err)
	}

	return nil
}

// Helper functions

func (s *IntegrationsService) getOAuthCredentials(integrationID string) (*OAuthCredentials, error) {
	query := `
		SELECT id, integration_id, access_token, refresh_token, token_type, expires_at, scope, created_at, updated_at
		FROM integration_oauth_credentials
		WHERE integration_id = ?
	`

	var creds OAuthCredentials
	var encryptedAccessToken, encryptedRefreshToken string

	err := s.db.QueryRow(query, integrationID).Scan(
		&creds.ID, &creds.IntegrationID, &encryptedAccessToken, &encryptedRefreshToken,
		&creds.TokenType, &creds.ExpiresAt, &creds.Scope, &creds.CreatedAt, &creds.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Decrypt tokens
	creds.AccessToken, err = s.encryptionSvc.Decrypt(encryptedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	if encryptedRefreshToken != "" {
		creds.RefreshToken, err = s.encryptionSvc.Decrypt(encryptedRefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
		}
	}

	return &creds, nil
}

func (s *IntegrationsService) getAPICredentials(integrationID string) ([]APICredentials, error) {
	query := `
		SELECT id, integration_id, credential_type, credential_name, credential_value, expires_at, created_at, updated_at
		FROM integration_api_credentials
		WHERE integration_id = ?
	`

	rows, err := s.db.Query(query, integrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials []APICredentials
	for rows.Next() {
		var cred APICredentials
		var encryptedValue string

		err := rows.Scan(
			&cred.ID, &cred.IntegrationID, &cred.CredentialType, &cred.CredentialName,
			&encryptedValue, &cred.ExpiresAt, &cred.CreatedAt, &cred.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Decrypt credential value
		cred.CredentialValue, err = s.encryptionSvc.Decrypt(encryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt credential: %w", err)
		}

		credentials = append(credentials, cred)
	}

	return credentials, nil
}

func (s *IntegrationsService) getRecentSyncHistory(integrationID string, limit int) ([]SyncHistory, error) {
	query := `
		SELECT id, integration_id, sync_type, status, items_synced, error_message, started_at, completed_at, created_at
		FROM integration_sync_history
		WHERE integration_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, integrationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []SyncHistory
	for rows.Next() {
		var sync SyncHistory
		err := rows.Scan(
			&sync.ID, &sync.IntegrationID, &sync.SyncType, &sync.Status,
			&sync.ItemsSynced, &sync.ErrorMessage, &sync.StartedAt,
			&sync.CompletedAt, &sync.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, sync)
	}

	return history, nil
}

// InitiateOAuth generates an OAuth authorization URL for an integration
func (s *IntegrationsService) InitiateOAuth(integrationID, host string) (string, error) {
	// Get integration to determine provider
	integration, err := s.GetIntegration(integrationID)
	if err != nil {
		return "", fmt.Errorf("failed to get integration: %w", err)
	}

	// Generate authorization URL based on provider
	switch integration.Provider {
	case ProviderGoogle:
		return fmt.Sprintf("http://%s/oauth/google/connect", host), nil
	default:
		return "", fmt.Errorf("OAuth not supported for provider: %s", integration.Provider)
	}
}

// generateID creates a new random ID
func generateID() string {
	// Simple implementation - in production you might want something more robust
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
