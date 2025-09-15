package encryption

import (
	"strings"
	"testing"

	"famstack/internal/config"
)

func TestFixedKeyProvider(t *testing.T) {
	// Test fixed key provider creation
	fixedConfig := config.FixedKeyConfig{
		Value: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}

	provider, err := NewFixedProvider(fixedConfig)
	if err != nil {
		t.Fatalf("Failed to create fixed provider: %v", err)
	}

	// Test GetEncryptionKey
	key, keyId, err := provider.GetEncryptionKey()
	if err != nil {
		t.Fatalf("GetEncryptionKey failed: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	if keyId != "fixed" {
		t.Errorf("Expected keyId 'fixed', got '%s'", keyId)
	}

	// Test GetDecryptionKey
	decryptKey, err := provider.GetDecryptionKey("any-key-id")
	if err != nil {
		t.Fatalf("GetDecryptionKey failed: %v", err)
	}

	// Should return the same key regardless of keyId
	if string(key) != string(decryptKey) {
		t.Error("GetDecryptionKey should return the same key as GetEncryptionKey for fixed provider")
	}
}

func TestFixedKeyProviderInvalidKey(t *testing.T) {
	tests := []struct {
		name        string
		keyValue    string
		expectError string
	}{
		{
			name:        "empty key",
			keyValue:    "",
			expectError: "fixed key value is required",
		},
		{
			name:        "invalid hex",
			keyValue:    "invalid-hex",
			expectError: "invalid hex key format",
		},
		{
			name:        "wrong length",
			keyValue:    "0123456789abcdef", // 16 bytes instead of 32
			expectError: "key must be exactly 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := config.FixedKeyConfig{Value: tt.keyValue}
			_, err := NewFixedProvider(config)

			if err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectError, err)
			}
		})
	}
}

func TestEncryptionServiceWithFixedKey(t *testing.T) {
	// Create encryption service with fixed key
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
	}

	service, err := NewService(encryptionConfig)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	// Test data
	testCases := []string{
		"Hello, World!",
		"Short",
		"This is a much longer string that contains various characters: !@#$%^&*()_+{}|:<>?[]\\;',./\"",
		"🔐 Unicode string with emojis! 🚀🎉",
		"", // Empty string
	}

	for _, plaintext := range testCases {
		t.Run("encrypt_decrypt_"+plaintext[:min(10, len(plaintext))], func(t *testing.T) {
			// Test encryption
			ciphertext, err := service.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Verify ciphertext format (should have prefix)
			if !strings.Contains(ciphertext, ":") {
				t.Error("Ciphertext should contain key identifier prefix")
			}

			parts := strings.SplitN(ciphertext, ":", 2)
			if len(parts) != 2 {
				t.Error("Ciphertext should have exactly one ':' separator")
			}

			if parts[0] != "fixed" {
				t.Errorf("Expected key identifier 'fixed', got '%s'", parts[0])
			}

			// Test decryption
			decrypted, err := service.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify round-trip
			if plaintext != decrypted {
				t.Errorf("Round-trip failed: expected '%s', got '%s'", plaintext, decrypted)
			}
		})
	}
}

func TestEncryptionConsistency(t *testing.T) {
	// Create two services with the same fixed key
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210",
		},
	}

	service1, err := NewService(encryptionConfig)
	if err != nil {
		t.Fatalf("Failed to create first service: %v", err)
	}

	service2, err := NewService(encryptionConfig)
	if err != nil {
		t.Fatalf("Failed to create second service: %v", err)
	}

	plaintext := "Cross-service encryption test"

	// Encrypt with service1
	ciphertext, err := service1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Service1 encryption failed: %v", err)
	}

	// Decrypt with service2
	decrypted, err := service2.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Service2 decryption failed: %v", err)
	}

	if plaintext != decrypted {
		t.Errorf("Cross-service test failed: expected '%s', got '%s'", plaintext, decrypted)
	}
}

func TestEncryptionRandomness(t *testing.T) {
	// Verify that encrypting the same plaintext produces different ciphertexts (due to random nonces)
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
	}

	service, err := NewService(encryptionConfig)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	plaintext := "Randomness test"

	// Encrypt the same plaintext multiple times
	ciphertexts := make([]string, 5)
	for i := range ciphertexts {
		ciphertext, err := service.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encryption %d failed: %v", i, err)
		}
		ciphertexts[i] = ciphertext
	}

	// Verify all ciphertexts are different (randomness from nonces)
	for i := 0; i < len(ciphertexts); i++ {
		for j := i + 1; j < len(ciphertexts); j++ {
			if ciphertexts[i] == ciphertexts[j] {
				t.Errorf("Ciphertexts %d and %d are identical, lack of randomness", i, j)
			}
		}
	}

	// Verify all decrypt to the same plaintext
	for i, ciphertext := range ciphertexts {
		decrypted, err := service.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decryption %d failed: %v", i, err)
		}
		if decrypted != plaintext {
			t.Errorf("Decryption %d: expected '%s', got '%s'", i, plaintext, decrypted)
		}
	}
}

func TestDecryptionErrors(t *testing.T) {
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba",
		},
	}

	service, err := NewService(encryptionConfig)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name        string
		ciphertext  string
		expectError string
	}{
		{
			name:        "no separator",
			ciphertext:  "noseparator",
			expectError: "missing key identifier",
		},
		{
			name:        "empty key id",
			ciphertext:  ":somedata",
			expectError: "ciphertext too short", // Fixed provider will try to decrypt with any keyId
		},
		{
			name:        "invalid base64",
			ciphertext:  "fixed:invalid-base64!",
			expectError: "failed to decode base64",
		},
		{
			name:        "corrupted data",
			ciphertext:  "fixed:dGVzdA==", // "test" in base64, but not valid encrypted data
			expectError: "ciphertext too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Decrypt(tt.ciphertext)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tt.name)
			}
			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectError, err)
			}
		})
	}
}

func TestExportActiveKey(t *testing.T) {
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
	}

	service, err := NewService(encryptionConfig)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	exportedKey, err := service.ExportActiveKey()
	if err != nil {
		t.Fatalf("ExportActiveKey failed: %v", err)
	}

	// Should return the key in hex format
	expectedKey := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"
	if exportedKey != expectedKey {
		t.Errorf("Expected exported key '%s', got '%s'", expectedKey, exportedKey)
	}
}

func TestMultipleProviderError(t *testing.T) {
	// Test that having multiple providers configured returns an error
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		Keyring: &config.KeyringConfig{
			Service: "test",
			Keys: map[string]config.KeyStatus{
				"test-key": "active",
			},
		},
	}

	_, err := NewService(encryptionConfig)
	if err == nil {
		t.Error("Expected error when multiple providers are configured")
	}

	if !strings.Contains(err.Error(), "multiple encryption providers configured") {
		t.Errorf("Expected 'multiple encryption providers configured' error, got: %v", err)
	}
}

func TestNoProviderError(t *testing.T) {
	// Test that having no providers configured returns an error
	encryptionConfig := config.EncryptionSettings{}

	_, err := NewService(encryptionConfig)
	if err == nil {
		t.Error("Expected error when no providers are configured")
	}

	if !strings.Contains(err.Error(), "no encryption provider configured") {
		t.Errorf("Expected 'no encryption provider configured' error, got: %v", err)
	}
}

// Helper function for string truncation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
