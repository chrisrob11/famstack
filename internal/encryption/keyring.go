package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/zalando/go-keyring"

	"famstack/internal/config"
)

// KeyringProvider implements key management using system keyring
type KeyringProvider struct {
	serviceName string
	keys        map[string]config.KeyStatus
}

// NewKeyringProvider creates a new keyring provider
func NewKeyringProvider(config config.KeyringConfig) (*KeyringProvider, error) {
	if config.Service == "" {
		return nil, fmt.Errorf("keyring service name is required")
	}

	if len(config.Keys) == 0 {
		return nil, fmt.Errorf("at least one key must be configured")
	}

	// Validate that exactly one key is active
	activeCount := 0
	for _, status := range config.Keys {
		if status == "active" {
			activeCount++
		}
	}

	if activeCount == 0 {
		return nil, fmt.Errorf("no active key found in keyring configuration")
	}
	if activeCount > 1 {
		return nil, fmt.Errorf("multiple active keys found in keyring configuration")
	}

	return &KeyringProvider{
		serviceName: config.Service,
		keys:        config.Keys,
	}, nil
}

// GetEncryptionKey returns the active key for encryption
func (p *KeyringProvider) GetEncryptionKey() ([]byte, string, error) {
	// Find the active key
	for keyName, status := range p.keys {
		if status == "active" {
			key, err := p.getOrCreateKey(keyName)
			if err != nil {
				return nil, "", fmt.Errorf("failed to get active key %s: %w", keyName, err)
			}
			return key, keyName, nil
		}
	}

	return nil, "", fmt.Errorf("no active key found in keyring configuration")
}

// GetDecryptionKey retrieves a key for decryption by its identifier
func (p *KeyringProvider) GetDecryptionKey(keyId string) ([]byte, error) {
	// Check if this key is configured
	status, exists := p.keys[keyId]
	if !exists {
		return nil, fmt.Errorf("key %s is not configured", keyId)
	}

	// Don't allow access to deprecated keys
	if status == "deprecated" {
		return nil, fmt.Errorf("key %s is deprecated and cannot be used for decryption", keyId)
	}

	return p.getOrCreateKey(keyId)
}

// getOrCreateKey retrieves a key from the keyring or creates it if it doesn't exist
func (p *KeyringProvider) getOrCreateKey(keyName string) ([]byte, error) {
	// Try to get existing key from keyring
	keyData, err := keyring.Get(p.serviceName, keyName)
	if err == keyring.ErrNotFound {
		// Key doesn't exist, create a new one
		return p.createNewKey(keyName)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve key from keyring: %w", err)
	}

	// Decode the base64 key
	key, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from keyring: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("key from keyring has invalid length: expected 32 bytes, got %d", len(key))
	}

	return key, nil
}

// createNewKey generates a new 32-byte key and stores it in the keyring
func (p *KeyringProvider) createNewKey(keyName string) ([]byte, error) {
	// Generate a new 32-byte key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode to base64 for storage
	encoded := base64.StdEncoding.EncodeToString(key)

	// Store in keyring
	if err := keyring.Set(p.serviceName, keyName, encoded); err != nil {
		return nil, fmt.Errorf("failed to store key in keyring: %w", err)
	}

	return key, nil
}

// ExportKey retrieves and returns a key in hex format for backup purposes
func (p *KeyringProvider) ExportKey(keyName string) (string, error) {
	key, err := p.getOrCreateKey(keyName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", key), nil
}
