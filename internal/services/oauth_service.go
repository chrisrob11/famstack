package services

import (
	"database/sql"
	"fmt"
	"time"

	"famstack/internal/database"
)

// OAuthService handles OAuth token and state database operations
type OAuthService struct {
	db *database.Fascade
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(db *database.Fascade) *OAuthService {
	return &OAuthService{db: db}
}

// OAuthToken represents an OAuth token
type OAuthToken struct {
	ID           string    `json:"id" db:"id"`
	FamilyID     string    `json:"family_id" db:"family_id"`
	UserID       string    `json:"created_by" db:"created_by"`
	Provider     string    `json:"provider" db:"provider"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	TokenType    string    `json:"token_type" db:"token_type"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	Scope        string    `json:"scope" db:"scope"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// OAuthState represents temporary OAuth state
type OAuthState struct {
	State     string    `json:"state" db:"state"`
	UserID    string    `json:"created_by" db:"created_by"`
	Provider  string    `json:"provider" db:"provider"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetToken retrieves stored OAuth token for user and provider
func (s *OAuthService) GetToken(userID string, provider string) (*OAuthToken, error) {
	query := `
		SELECT id, family_id, created_by, provider, access_token, refresh_token,
		       token_type, expires_at, scope, created_at, updated_at
		FROM oauth_tokens
		WHERE created_by = ? AND provider = ?
	`

	var token OAuthToken
	err := s.db.QueryRow(query, userID, provider).Scan(
		&token.ID, &token.FamilyID, &token.UserID, &token.Provider,
		&token.AccessToken, &token.RefreshToken, &token.TokenType,
		&token.ExpiresAt, &token.Scope, &token.CreatedAt, &token.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

// SaveToken stores OAuth token in database
func (s *OAuthService) SaveToken(token *OAuthToken) error {
	query := `
		INSERT OR REPLACE INTO oauth_tokens
		(id, family_id, created_by, provider, access_token, refresh_token,
		 token_type, expires_at, scope, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		token.ID, token.FamilyID, token.UserID, token.Provider,
		token.AccessToken, token.RefreshToken, token.TokenType,
		token.ExpiresAt, token.Scope, token.CreatedAt, token.UpdatedAt,
	)

	return err
}

// DeleteToken removes OAuth token for user and provider
func (s *OAuthService) DeleteToken(userID string, provider string) error {
	query := `DELETE FROM oauth_tokens WHERE created_by = ? AND provider = ?`
	_, err := s.db.Exec(query, userID, provider)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// SaveState stores OAuth state temporarily
func (s *OAuthService) SaveState(state *OAuthState) error {
	query := `
		INSERT OR REPLACE INTO oauth_states
		(state, created_by, provider, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		state.State, state.UserID, state.Provider,
		state.ExpiresAt, state.CreatedAt,
	)

	return err
}

// GetState retrieves OAuth state
func (s *OAuthService) GetState(state string) (*OAuthState, error) {
	query := `
		SELECT state, created_by, provider, expires_at, created_at
		FROM oauth_states
		WHERE state = ?
	`

	var stateData OAuthState
	err := s.db.QueryRow(query, state).Scan(
		&stateData.State, &stateData.UserID, &stateData.Provider,
		&stateData.ExpiresAt, &stateData.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("state not found")
		}
		return nil, fmt.Errorf("failed to get state: %w", err)
	}

	// Check if state has expired (do this in Go to handle timezone properly)
	if time.Now().UTC().After(stateData.ExpiresAt) {
		return nil, fmt.Errorf("state expired")
	}

	return &stateData, nil
}

// DeleteState removes OAuth state
func (s *OAuthService) DeleteState(state string) error {
	query := `DELETE FROM oauth_states WHERE state = ?`
	_, err := s.db.Exec(query, state)
	return err
}

// GetUserFamilyID gets the family ID for a user
func (s *OAuthService) GetUserFamilyID(userID string) (string, error) {
	query := `SELECT family_id FROM family_members WHERE id = ?`

	var familyID string
	err := s.db.QueryRow(query, userID).Scan(&familyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to get user family ID: %w", err)
	}

	return familyID, nil
}

// GetUsersWithTokens returns all users with OAuth tokens for a specific provider
func (s *OAuthService) GetUsersWithTokens(provider string) ([]OAuthToken, error) {
	query := `
		SELECT id, family_id, created_by, provider, access_token, refresh_token,
		       token_type, expires_at, scope, created_at, updated_at
		FROM oauth_tokens
		WHERE provider = ? AND expires_at > datetime('now')
	`

	rows, err := s.db.Query(query, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query OAuth tokens: %w", err)
	}
	defer rows.Close()

	var tokens []OAuthToken
	for rows.Next() {
		var token OAuthToken
		if scanErr := rows.Scan(
			&token.ID, &token.FamilyID, &token.UserID, &token.Provider,
			&token.AccessToken, &token.RefreshToken, &token.TokenType,
			&token.ExpiresAt, &token.Scope, &token.CreatedAt, &token.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan OAuth token: %w", scanErr)
		}
		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating OAuth tokens: %w", err)
	}

	return tokens, nil
}
