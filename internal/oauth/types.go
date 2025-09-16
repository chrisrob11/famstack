package oauth

import (
	"time"
)

// Provider represents different OAuth providers
type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
	ProviderApple     Provider = "apple"
)

// OAuthToken represents stored OAuth credentials
type OAuthToken struct {
	ID           string    `json:"id" db:"id"`
	FamilyID     string    `json:"family_id" db:"family_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Provider     Provider  `json:"provider" db:"provider"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	TokenType    string    `json:"token_type" db:"token_type"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	Scope        string    `json:"scope" db:"scope"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// OAuthConfig holds OAuth configuration for providers
type OAuthConfig struct {
	Google *GoogleConfig `json:"google,omitempty"`
}

// GoogleConfig holds Google OAuth configuration
type GoogleConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// OAuthState represents temporary OAuth state for CSRF protection
type OAuthState struct {
	State     string    `json:"state"`
	Provider  Provider  `json:"provider"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
