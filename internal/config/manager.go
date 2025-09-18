package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Config represents the application configuration
type Config struct {
	Version  string        `json:"version"`
	Server   ServerConfig  `json:"server"`
	OAuth    OAuthConfig   `json:"oauth"`
	Features FeatureConfig `json:"features"`
	mu       sync.RWMutex  `json:"-"`
	path     string        `json:"-"`
}

// ServerConfig holds server-specific settings
type ServerConfig struct {
	Port    string `json:"port"`
	DevMode bool   `json:"dev_mode"`
}

// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
	Google *OAuthProvider `json:"google,omitempty"`
}

// OAuthProvider holds OAuth configuration for a specific provider
type OAuthProvider struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	Configured   bool     `json:"configured"`
}

// FeatureConfig holds feature flags
type FeatureConfig struct {
	CalendarSync       bool `json:"calendar_sync"`
	EmailNotifications bool `json:"email_notifications"`
}

// Manager handles configuration file operations
type Manager struct {
	config *Config
}

// NewManager creates a new config manager
func NewManager(configPath string) (*Manager, error) {
	manager := &Manager{}

	// Load or create config
	config, err := manager.loadOrCreateConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	manager.config = config
	return manager, nil
}

// loadOrCreateConfig loads existing config or creates a new one
func (m *Manager) loadOrCreateConfig(path string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config
		config := m.createDefaultConfig(path)
		if err := m.saveConfig(config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		fmt.Printf("Created default configuration file: %s\n", path)
		return config, nil
	}

	// Load existing config
	return m.loadConfig(path)
}

// createDefaultConfig creates a config with default values
func (m *Manager) createDefaultConfig(path string) *Config {
	return &Config{
		Version: "1.0",
		path:    path,
		Server: ServerConfig{
			Port:    "8080",
			DevMode: false,
		},
		OAuth: OAuthConfig{
			Google: &OAuthProvider{
				ClientID:     "",
				ClientSecret: "",
				RedirectURL:  "http://localhost:8080/oauth/google/callback",
				Scopes:       []string{"https://www.googleapis.com/auth/calendar.readonly"},
				Configured:   false,
			},
		},
		Features: FeatureConfig{
			CalendarSync:       true,
			EmailNotifications: false,
		},
	}
}

// loadConfig loads configuration from file
func (m *Manager) loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config.path = path
	return &config, nil
}

// saveConfig saves configuration to file with proper locking
func (m *Manager) saveConfig(config *Config) error {
	config.mu.Lock()
	defer config.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(config.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(config.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns a copy of the current configuration
func (m *Manager) GetConfig() *Config {
	m.config.mu.RLock()
	defer m.config.mu.RUnlock()

	// Return a copy without the mutex to prevent external modifications
	configCopy := Config{
		Version:  m.config.Version,
		Server:   m.config.Server,
		OAuth:    m.config.OAuth,
		Features: m.config.Features,
		path:     m.config.path,
		// Don't copy the mutex
	}
	return &configCopy
}

// UpdateOAuthProvider updates OAuth configuration for a provider
func (m *Manager) UpdateOAuthProvider(provider string, config *OAuthProvider) error {
	// Update config in memory with proper locking
	func() {
		m.config.mu.Lock()
		defer m.config.mu.Unlock()

		switch provider {
		case "google":
			m.config.OAuth.Google = config
		}
	}()

	// Validate provider after updating
	if provider != "google" {
		return fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	// Save to file (this has its own locking)
	return m.saveConfig(m.config)
}

// GetOAuthProvider returns OAuth configuration for a provider
func (m *Manager) GetOAuthProvider(provider string) (*OAuthProvider, error) {
	m.config.mu.RLock()
	defer m.config.mu.RUnlock()

	switch provider {
	case "google":
		if m.config.OAuth.Google == nil {
			return nil, fmt.Errorf("google OAuth not configured")
		}
		// Return a copy
		providerCopy := *m.config.OAuth.Google
		return &providerCopy, nil
	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}

// UpdateServerConfig updates server configuration
func (m *Manager) UpdateServerConfig(config ServerConfig) error {
	// Update config in memory with proper locking
	func() {
		m.config.mu.Lock()
		defer m.config.mu.Unlock()
		m.config.Server = config
	}()

	// Save to file (this has its own locking)
	return m.saveConfig(m.config)
}

// UpdateFeatureConfig updates feature configuration
func (m *Manager) UpdateFeatureConfig(config FeatureConfig) error {
	// Update config in memory with proper locking
	func() {
		m.config.mu.Lock()
		defer m.config.mu.Unlock()
		m.config.Features = config
	}()

	// Save to file (this has its own locking)
	return m.saveConfig(m.config)
}
