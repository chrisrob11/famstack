package integrations

import (
	"time"
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
	UserID          string          `json:"user_id" db:"user_id"`
	IntegrationType IntegrationType `json:"integration_type" db:"integration_type"`
	Provider        Provider        `json:"provider" db:"provider"`
	AuthMethod      AuthMethod      `json:"auth_method" db:"auth_method"`
	Status          Status          `json:"status" db:"status"`
	DisplayName     string          `json:"display_name" db:"display_name"`
	Description     string          `json:"description" db:"description"`
	Settings        string          `json:"settings" db:"settings"` // JSON
	LastSyncAt      *time.Time      `json:"last_sync_at" db:"last_sync_at"`
	LastError       string          `json:"last_error" db:"last_error"`
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
	UserID          *string          `json:"user_id"`
	Limit           int              `json:"limit"`
	Offset          int              `json:"offset"`
}
