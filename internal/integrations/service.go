package integrations

import (
	"encoding/json"
	"fmt"
	"time"

	"famstack/internal/database"
	"famstack/internal/encryption"
)

// Service handles integration operations
type Service struct {
	db            *database.DB
	encryptionSvc *encryption.Service
}

// NewService creates a new integrations service
func NewService(db *database.DB, encryptionSvc *encryption.Service) *Service {
	return &Service{
		db:            db,
		encryptionSvc: encryptionSvc,
	}
}

// CreateIntegration creates a new integration
func (s *Service) CreateIntegration(familyID, userID string, req *CreateIntegrationRequest) (*Integration, error) {
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
		UserID:          userID,
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
		(id, family_id, user_id, integration_type, provider, auth_method, status,
		 display_name, description, settings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		integration.ID, integration.FamilyID, integration.UserID,
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
func (s *Service) GetIntegration(integrationID string) (*Integration, error) {
	query := `
		SELECT id, family_id, user_id, integration_type, provider, auth_method,
		       status, display_name, description, settings, last_sync_at,
		       last_error, created_at, updated_at
		FROM integrations
		WHERE id = ?
	`

	var integration Integration
	err := s.db.QueryRow(query, integrationID).Scan(
		&integration.ID, &integration.FamilyID, &integration.UserID,
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
func (s *Service) GetIntegrationWithCredentials(integrationID string) (*IntegrationWithCredentials, error) {
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
func (s *Service) ListIntegrations(familyID string, query *ListIntegrationsQuery) ([]Integration, error) {
	sql := `
		SELECT id, family_id, user_id, integration_type, provider, auth_method,
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
	if query.UserID != nil {
		sql += " AND user_id = ?"
		args = append(args, *query.UserID)
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
			&integration.ID, &integration.FamilyID, &integration.UserID,
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
func (s *Service) UpdateIntegration(integrationID string, req *UpdateIntegrationRequest) (*Integration, error) {
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
func (s *Service) DeleteIntegration(integrationID string) error {
	// Note: credentials will be deleted by CASCADE
	query := "DELETE FROM integrations WHERE id = ?"
	_, err := s.db.Exec(query, integrationID)
	if err != nil {
		return fmt.Errorf("failed to delete integration: %w", err)
	}
	return nil
}

// StoreOAuthCredentials stores encrypted OAuth credentials
func (s *Service) StoreOAuthCredentials(integrationID string, accessToken, refreshToken, tokenType, scope string, expiresAt *time.Time) error {
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

	return err
}

// Helper functions

func (s *Service) getOAuthCredentials(integrationID string) (*OAuthCredentials, error) {
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

func (s *Service) getAPICredentials(integrationID string) ([]APICredentials, error) {
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

func (s *Service) getRecentSyncHistory(integrationID string, limit int) ([]SyncHistory, error) {
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

// generateID creates a new random ID
func generateID() string {
	// Simple implementation - in production you might want something more robust
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
