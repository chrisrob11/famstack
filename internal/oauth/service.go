package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"

	"famstack/internal/database"
	"famstack/internal/encryption"
)

// Service handles OAuth operations
type Service struct {
	db            *database.DB
	config        *OAuthConfig
	googleConfig  *oauth2.Config
	encryptionSvc *encryption.Service
}

// NewService creates a new OAuth service
func NewService(db *database.DB, config *OAuthConfig, encryptionSvc *encryption.Service) *Service {
	var googleConfig *oauth2.Config
	if config.Google != nil {
		googleConfig = &oauth2.Config{
			ClientID:     config.Google.ClientID,
			ClientSecret: config.Google.ClientSecret,
			RedirectURL:  config.Google.RedirectURL,
			Scopes:       config.Google.Scopes,
			Endpoint:     google.Endpoint,
		}
	}

	return &Service{
		db:            db,
		config:        config,
		googleConfig:  googleConfig,
		encryptionSvc: encryptionSvc,
	}
}

// GetAuthURL generates OAuth authorization URL for provider
func (s *Service) GetAuthURL(provider Provider, userID string) (string, error) {
	state, err := s.generateState(provider, userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	switch provider {
	case ProviderGoogle:
		return s.getGoogleAuthURL(state)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

// HandleCallback processes OAuth callback
func (s *Service) HandleCallback(provider Provider, code, state string) (*OAuthToken, error) {
	// Verify state
	if !s.verifyState(state) {
		return nil, fmt.Errorf("invalid state parameter")
	}

	switch provider {
	case ProviderGoogle:
		return s.handleGoogleCallback(code, state)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// GetToken retrieves stored OAuth token for user and provider
func (s *Service) GetToken(userID string, provider Provider) (*OAuthToken, error) {
	query := `
		SELECT id, family_id, user_id, provider, access_token, refresh_token,
		       token_type, expires_at, scope, created_at, updated_at
		FROM oauth_tokens
		WHERE user_id = ? AND provider = ?
	`

	var token OAuthToken
	var encryptedAccessToken, encryptedRefreshToken string
	err := s.db.QueryRow(query, userID, provider).Scan(
		&token.ID, &token.FamilyID, &token.UserID, &token.Provider,
		&encryptedAccessToken, &encryptedRefreshToken, &token.TokenType,
		&token.ExpiresAt, &token.Scope, &token.CreatedAt, &token.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Decrypt tokens
	if s.encryptionSvc != nil {
		token.AccessToken, err = s.encryptionSvc.Decrypt(encryptedAccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt access token: %w", err)
		}

		if encryptedRefreshToken != "" {
			token.RefreshToken, err = s.encryptionSvc.Decrypt(encryptedRefreshToken)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
			}
		}
	} else {
		token.AccessToken = encryptedAccessToken
		token.RefreshToken = encryptedRefreshToken
	}

	return &token, nil
}

// GetUserIDFromState extracts user ID from OAuth state parameter
func (s *Service) GetUserIDFromState(state string) (string, error) {
	stateData, err := s.getState(state)
	if err != nil {
		return "", fmt.Errorf("failed to get state data: %w", err)
	}
	return stateData.UserID, nil
}

// RefreshToken refreshes an expired OAuth token
func (s *Service) RefreshToken(token *OAuthToken) (*OAuthToken, error) {
	switch token.Provider {
	case ProviderGoogle:
		return s.refreshGoogleToken(token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", token.Provider)
	}
}

// Private helper methods

func (s *Service) generateState(provider Provider, userID string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	state := hex.EncodeToString(bytes)

	// Store state temporarily with user ID for callback processing
	// TODO: Store in cache/database with expiration
	// For now, encode userID in state (in production, use proper state storage)
	stateData := &OAuthState{
		State:     state,
		UserID:    userID,
		Provider:  provider,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
	}

	if err := s.saveState(stateData); err != nil {
		return "", fmt.Errorf("failed to save state: %w", err)
	}

	return state, nil
}

func (s *Service) verifyState(state string) bool {
	// Verify state exists in database and hasn't expired
	_, err := s.getState(state)
	return err == nil
}

func (s *Service) getGoogleAuthURL(state string) (string, error) {
	if s.googleConfig == nil {
		return "", fmt.Errorf("google OAuth not configured")
	}

	return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

func (s *Service) handleGoogleCallback(code, state string) (*OAuthToken, error) {
	if s.googleConfig == nil {
		return nil, fmt.Errorf("google OAuth not configured")
	}

	// Extract user ID from stored state
	stateData, err := s.getState(state)
	if err != nil {
		return nil, fmt.Errorf("failed to get state data: %w", err)
	}

	// Get user's family ID
	familyID, err := s.getUserFamilyID(stateData.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user family ID: %w", err)
	}

	// Exchange authorization code for token
	ctx := context.Background()
	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Create our OAuth token struct
	oauthToken := &OAuthToken{
		ID:           generateID(),
		FamilyID:     familyID,
		UserID:       stateData.UserID,
		Provider:     ProviderGoogle,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Scope:        calendar.CalendarReadonlyScope,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save token to database
	if err := s.saveToken(oauthToken); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	// Clean up state
	if err := s.deleteState(state); err != nil {
		// Log error but don't fail the entire operation
		fmt.Printf("Warning: Failed to delete OAuth state: %v\n", err)
	}

	return oauthToken, nil
}

func (s *Service) refreshGoogleToken(token *OAuthToken) (*OAuthToken, error) {
	if s.googleConfig == nil {
		return nil, fmt.Errorf("google OAuth not configured")
	}

	// Create oauth2.Token from our stored token
	oauth2Token := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.ExpiresAt,
	}

	// Create token source that will refresh automatically
	ctx := context.Background()
	tokenSource := s.googleConfig.TokenSource(ctx, oauth2Token)

	// Get fresh token
	freshToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update our token struct
	token.AccessToken = freshToken.AccessToken
	if freshToken.RefreshToken != "" {
		token.RefreshToken = freshToken.RefreshToken
	}
	token.TokenType = freshToken.TokenType
	token.ExpiresAt = freshToken.Expiry
	token.UpdatedAt = time.Now()

	// Save updated token
	if err := s.saveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return token, nil
}

// saveToken stores OAuth token in database with encryption
func (s *Service) saveToken(token *OAuthToken) error {
	var encryptedAccessToken, encryptedRefreshToken string
	var err error

	// Encrypt tokens before storing
	if s.encryptionSvc != nil {
		encryptedAccessToken, err = s.encryptionSvc.Encrypt(token.AccessToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt access token: %w", err)
		}

		if token.RefreshToken != "" {
			encryptedRefreshToken, err = s.encryptionSvc.Encrypt(token.RefreshToken)
			if err != nil {
				return fmt.Errorf("failed to encrypt refresh token: %w", err)
			}
		}
	} else {
		encryptedAccessToken = token.AccessToken
		encryptedRefreshToken = token.RefreshToken
	}

	query := `
		INSERT OR REPLACE INTO oauth_tokens
		(id, family_id, user_id, provider, access_token, refresh_token,
		 token_type, expires_at, scope, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		token.ID, token.FamilyID, token.UserID, token.Provider,
		encryptedAccessToken, encryptedRefreshToken, token.TokenType,
		token.ExpiresAt, token.Scope, token.CreatedAt, token.UpdatedAt,
	)

	return err
}

// generateID creates a new random ID
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// This should rarely happen, but we need to handle it
		panic(fmt.Sprintf("Failed to generate random bytes: %v", err))
	}
	return hex.EncodeToString(bytes)
}

// GetOAuth2Config returns the oauth2.Config for external use
func (s *Service) GetOAuth2Config() *oauth2.Config {
	return s.googleConfig
}

// GetOAuth2Token converts OAuthToken to oauth2.Token
func (s *Service) GetOAuth2Token(token *OAuthToken) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.ExpiresAt,
	}
}

// saveState stores OAuth state temporarily
func (s *Service) saveState(state *OAuthState) error {
	query := `
		INSERT OR REPLACE INTO oauth_states
		(state, user_id, provider, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		state.State, state.UserID, state.Provider,
		state.ExpiresAt, state.CreatedAt,
	)

	return err
}

// getState retrieves OAuth state
func (s *Service) getState(state string) (*OAuthState, error) {
	query := `
		SELECT state, user_id, provider, expires_at, created_at
		FROM oauth_states
		WHERE state = ? AND expires_at > datetime('now')
	`

	var stateData OAuthState
	err := s.db.QueryRow(query, state).Scan(
		&stateData.State, &stateData.UserID, &stateData.Provider,
		&stateData.ExpiresAt, &stateData.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	return &stateData, nil
}

// deleteState removes OAuth state
func (s *Service) deleteState(state string) error {
	query := `DELETE FROM oauth_states WHERE state = ?`
	_, err := s.db.Exec(query, state)
	return err
}

// getUserFamilyID gets the family ID for a user
func (s *Service) getUserFamilyID(userID string) (string, error) {
	query := `SELECT family_id FROM family_members WHERE id = ?`

	var familyID string
	err := s.db.QueryRow(query, userID).Scan(&familyID)
	if err != nil {
		return "", fmt.Errorf("failed to get user family ID: %w", err)
	}

	return familyID, nil
}

// DeleteToken removes OAuth token for user and provider
func (s *Service) DeleteToken(userID string, provider Provider) error {
	query := `DELETE FROM oauth_tokens WHERE user_id = ? AND provider = ?`
	_, err := s.db.Exec(query, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
