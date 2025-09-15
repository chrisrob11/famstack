package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"famstack/internal/config"
)

// Service provides encryption and decryption capabilities
type Service struct {
	provider interface {
		GetEncryptionKey() ([]byte, string, error)     // Returns key + identifier for new encryption
		GetDecryptionKey(keyId string) ([]byte, error) // Returns key for decrypting existing data
	}
	config config.EncryptionSettings
}

// NewService creates a new encryption service from configuration
func NewService(encryptionConfig config.EncryptionSettings) (*Service, error) {
	provider, err := encryptionConfig.GetActiveProvider()
	if err != nil {
		return nil, fmt.Errorf("invalid encryption configuration: %w", err)
	}

	var keyProvider interface {
		GetEncryptionKey() ([]byte, string, error)
		GetDecryptionKey(keyId string) ([]byte, error)
	}

	switch provider {
	case "fixed_key":
		if encryptionConfig.FixedKey == nil {
			return nil, fmt.Errorf("fixed_key configuration is required")
		}
		keyProvider, err = NewFixedProvider(*encryptionConfig.FixedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create fixed key provider: %w", err)
		}

	case "keyring":
		if encryptionConfig.Keyring == nil {
			return nil, fmt.Errorf("keyring configuration is required")
		}
		keyProvider, err = NewKeyringProvider(*encryptionConfig.Keyring)
		if err != nil {
			return nil, fmt.Errorf("failed to create keyring provider: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported encryption provider: %s", provider)
	}

	return &Service{
		provider: keyProvider,
		config:   encryptionConfig,
	}, nil
}

// Encrypt encrypts a string using AES-256-GCM and returns base64 encoded result
func (s *Service) Encrypt(plaintext string) (string, error) {
	// Get the encryption key and its identifier from the provider
	key, keyId, err := s.provider.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt the data
	ciphertext, err := s.encryptWithKey(key, plaintext)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	// Prefix with key identifier for future decryption
	return fmt.Sprintf("%s:%s", keyId, ciphertext), nil
}

// Decrypt decrypts a string that was encrypted with Encrypt
func (s *Service) Decrypt(ciphertext string) (string, error) {
	// Parse key identifier from ciphertext
	parts := strings.SplitN(ciphertext, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid ciphertext format: missing key identifier")
	}

	keyId := parts[0]
	encodedData := parts[1]

	// Get the key for this identifier
	key, err := s.provider.GetDecryptionKey(keyId)
	if err != nil {
		return "", fmt.Errorf("failed to get decryption key: %w", err)
	}

	// Decrypt the data
	plaintext, err := s.decryptWithKey(key, encodedData)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// encryptWithKey performs AES-256-GCM encryption
func (s *Service) encryptWithKey(key []byte, plaintext string) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptWithKey performs AES-256-GCM decryption
func (s *Service) decryptWithKey(key []byte, encodedData string) (string, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// ExportActiveKey exports the active key in hex format for backup
func (s *Service) ExportActiveKey() (string, error) {
	key, _, err := s.provider.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("failed to get active key: %w", err)
	}

	return fmt.Sprintf("%x", key), nil
}
