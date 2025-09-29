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

	"famstack/internal/encryption"
	"famstack/internal/services"
)

// Service handles OAuth operations
type Service struct {
	oauthService  *services.OAuthService
	config        *OAuthConfig
	googleConfig  *oauth2.Config
	encryptionSvc *encryption.Service
}

// NewService creates a new OAuth service
func NewService(oauthService *services.OAuthService, config *OAuthConfig, encryptionSvc *encryption.Service) *Service {
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
		oauthService:  oauthService,
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

// GetAuthURLWithCustomState generates OAuth authorization URL with custom state
func (s *Service) GetAuthURLWithCustomState(provider Provider, customState string) (string, error) {
	// Store the custom state in our state tracking system
	err := s.storeCustomState(customState)
	if err != nil {
		return "", fmt.Errorf("failed to store custom state: %w", err)
	}

	switch provider {
	case ProviderGoogle:
		return s.getGoogleAuthURL(customState)
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
	serviceToken, err := s.oauthService.GetToken(userID, string(provider))
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// Convert from service token to OAuth token and decrypt
	token := &OAuthToken{
		ID:        serviceToken.ID,
		FamilyID:  serviceToken.FamilyID,
		UserID:    serviceToken.UserID,
		Provider:  Provider(serviceToken.Provider),
		TokenType: serviceToken.TokenType,
		ExpiresAt: serviceToken.ExpiresAt,
		Scope:     serviceToken.Scope,
		CreatedAt: serviceToken.CreatedAt,
		UpdatedAt: serviceToken.UpdatedAt,
	}

	// Decrypt tokens
	if s.encryptionSvc != nil {
		token.AccessToken, err = s.encryptionSvc.Decrypt(serviceToken.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt access token: %w", err)
		}

		if serviceToken.RefreshToken != "" {
			token.RefreshToken, err = s.encryptionSvc.Decrypt(serviceToken.RefreshToken)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
			}
		}
	} else {
		token.AccessToken = serviceToken.AccessToken
		token.RefreshToken = serviceToken.RefreshToken
	}

	return token, nil
}

// GetUserIDFromState extracts user ID from OAuth state parameter
func (s *Service) GetUserIDFromState(state string) (string, error) {
	stateData, err := s.oauthService.GetState(state)
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
	stateData := &services.OAuthState{
		State:     state,
		UserID:    userID,
		Provider:  string(provider),
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.oauthService.SaveState(stateData); err != nil {
		return "", fmt.Errorf("failed to save state: %w", err)
	}

	return state, nil
}

// storeCustomState stores a custom state (containing userID and config) in the database
func (s *Service) storeCustomState(customState string) error {
	// For custom states, we don't have separate provider/userID, so we store the full state
	// The verifyState method will work the same way
	stateData := &services.OAuthState{
		State:     customState,
		UserID:    "",       // Will be extracted from state later
		Provider:  "google", // Assume Google for now
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	return s.oauthService.SaveState(stateData)
}

func (s *Service) verifyState(state string) bool {
	// Verify state exists in database and hasn't expired
	fmt.Printf("🔍 Verifying OAuth state\n")
	_, err := s.oauthService.GetState(state)
	if err != nil {
		fmt.Printf("❌ State verification failed\n")
		return false
	}
	fmt.Printf("✅ State verified successfully\n")
	return true
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
	stateData, err := s.oauthService.GetState(state)
	if err != nil {
		return nil, fmt.Errorf("failed to get state data: %w", err)
	}

	// Get user's family ID
	familyID, err := s.oauthService.GetUserFamilyID(stateData.UserID)
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
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Save token to database
	if err := s.saveTokenWithEncryption(oauthToken); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	// Clean up state
	if err := s.oauthService.DeleteState(state); err != nil {
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
	token.UpdatedAt = time.Now().UTC()

	// Save updated token
	if err := s.saveTokenWithEncryption(token); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return token, nil
}

// saveTokenWithEncryption stores OAuth token in database with encryption
func (s *Service) saveTokenWithEncryption(token *OAuthToken) error {
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

	// Convert to service token format
	serviceToken := &services.OAuthToken{
		ID:           token.ID,
		FamilyID:     token.FamilyID,
		UserID:       token.UserID,
		Provider:     string(token.Provider),
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.ExpiresAt,
		Scope:        token.Scope,
		CreatedAt:    token.CreatedAt,
		UpdatedAt:    token.UpdatedAt,
	}

	return s.oauthService.SaveToken(serviceToken)
}

// generateID creates a new cryptographically secure random ID
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if crypto/rand fails (should be extremely rare)
		return fmt.Sprintf("oauth_fallback_%d", time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("oauth_%s", hex.EncodeToString(bytes))
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

// DeleteToken removes OAuth token for user and provider
func (s *Service) DeleteToken(userID string, provider Provider) error {
	return s.oauthService.DeleteToken(userID, string(provider))
}
