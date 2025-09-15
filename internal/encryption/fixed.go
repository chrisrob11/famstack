package encryption

import (
	"encoding/hex"
	"fmt"

	"famstack/internal/config"
)

// FixedProvider implements key management using fixed keys from configuration
type FixedProvider struct {
	key []byte
}

// NewFixedProvider creates a new fixed key provider
func NewFixedProvider(config config.FixedKeyConfig) (*FixedProvider, error) {
	if config.Value == "" {
		return nil, fmt.Errorf("fixed key value is required")
	}

	// Decode hex string to bytes
	key, err := hex.DecodeString(config.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key format: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("key must be exactly 32 bytes (64 hex characters), got %d bytes", len(key))
	}

	return &FixedProvider{
		key: key,
	}, nil
}

// GetEncryptionKey returns the fixed key for encryption (always the same)
func (p *FixedProvider) GetEncryptionKey() ([]byte, string, error) {
	// Fixed provider always uses the same key with "fixed" identifier
	return p.key, "fixed", nil
}

// GetDecryptionKey returns the fixed key for decryption
func (p *FixedProvider) GetDecryptionKey(keyId string) ([]byte, error) {
	// For fixed provider, keyId is ignored - always return the same key
	// This allows decrypting data encrypted with any identifier
	return p.key, nil
}
