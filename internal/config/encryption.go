package config

import "fmt"

// EncryptionSettings holds configuration for encryption providers
type EncryptionSettings struct {
	FixedKey *FixedKeyConfig `json:"fixed_key,omitempty"`
	Keyring  *KeyringConfig  `json:"keyring,omitempty"`
	// Future: Vault, KMS, etc.
}

// FixedKeyConfig for development/testing environments
type FixedKeyConfig struct {
	Value string `json:"value" env:"FAMSTACK_FIXED_KEY_VALUE"`
}

// KeyringConfig for system keyring storage
type KeyringConfig struct {
	Service string               `json:"service" env:"FAMSTACK_KEYRING_SERVICE"`
	Keys    map[string]KeyStatus `json:"keys"`
}

// KeyStatus represents the status of a key
type KeyStatus string

const (
	KeyStatusActive     KeyStatus = "active"
	KeyStatusInactive   KeyStatus = "inactive"
	KeyStatusDeprecated KeyStatus = "deprecated"
)

// GetActiveProvider returns which encryption provider is configured
func (es *EncryptionSettings) GetActiveProvider() (string, error) {
	providers := []string{}

	if es.FixedKey != nil && es.FixedKey.Value != "" {
		providers = append(providers, "fixed_key")
	}
	if es.Keyring != nil && len(es.Keyring.Keys) > 0 {
		providers = append(providers, "keyring")
	}

	if len(providers) == 0 {
		return "", fmt.Errorf("no encryption provider configured")
	}
	if len(providers) > 1 {
		return "", fmt.Errorf("multiple encryption providers configured: %v", providers)
	}

	return providers[0], nil
}

// GetActiveKeyName returns the name of the active key for keyring provider
func (kc *KeyringConfig) GetActiveKeyName() (string, error) {
	for keyName, status := range kc.Keys {
		if status == KeyStatusActive {
			return keyName, nil
		}
	}
	return "", fmt.Errorf("no active key found in keyring configuration")
}

// GetAvailableKeys returns all keys that can be used for decryption
func (kc *KeyringConfig) GetAvailableKeys() []string {
	var available []string
	for keyName, status := range kc.Keys {
		if status == KeyStatusActive || status == KeyStatusInactive {
			available = append(available, keyName)
		}
	}
	return available
}

// DefaultEncryptionSettings returns default configuration
func DefaultEncryptionSettings() *EncryptionSettings {
	return &EncryptionSettings{
		Keyring: &KeyringConfig{
			Service: "famstack",
			Keys: map[string]KeyStatus{
				"famstack-master-key": KeyStatusActive,
			},
		},
	}
}
