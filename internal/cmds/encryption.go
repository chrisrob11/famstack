package cmds

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/urfave/cli/v2"

	"famstack/internal/config"
	"famstack/internal/encryption"
)

// EncryptionCommand returns the encryption command configuration
func EncryptionCommand() *cli.Command {
	return &cli.Command{
		Name:  "encryption",
		Usage: "Encryption key management",
		Subcommands: []*cli.Command{
			{
				Name:   "status",
				Usage:  "Show encryption provider status",
				Action: encryptionStatus,
			},
			{
				Name:   "export-key",
				Usage:  "Export the current master key for backup",
				Action: exportKey,
			},
			{
				Name:   "generate-key",
				Usage:  "Generate a new fixed key for development",
				Action: generateFixedKey,
			},
		},
	}
}

// encryptionStatus shows the current encryption configuration status
func encryptionStatus(ctx *cli.Context) error {
	// Use default config for now - in future this would load from config file
	encryptionConfig := config.DefaultEncryptionSettings()

	provider, err := encryptionConfig.GetActiveProvider()
	if err != nil {
		return fmt.Errorf("encryption configuration error: %w", err)
	}

	fmt.Printf("🔒 Active Provider: %s\n", provider)

	switch provider {
	case "keyring":
		fmt.Printf("   Service: %s\n", encryptionConfig.Keyring.Service)
		fmt.Printf("   Keys:\n")
		for keyName, status := range encryptionConfig.Keyring.Keys {
			var statusIcon string
			switch status {
			case "active":
				statusIcon = "✅"
			case "deprecated":
				statusIcon = "⚠️"
			default:
				statusIcon = "📦"
			}
			fmt.Printf("     %s %s (%s)\n", statusIcon, keyName, status)
		}
	case "fixed_key":
		fmt.Println("   ⚠️  Using fixed key (development only)")
	}

	return nil
}

// exportKey exports the current master key for backup
func exportKey(ctx *cli.Context) error {
	// Use default config for now
	encryptionConfig := config.DefaultEncryptionSettings()

	encryptionService, err := encryption.NewService(*encryptionConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	key, err := encryptionService.ExportActiveKey()
	if err != nil {
		return fmt.Errorf("failed to export key: %w", err)
	}

	fmt.Printf("🔐 Master Key (SAVE SECURELY): %s\n", key)
	fmt.Println("⚠️  Anyone with this key can decrypt your data!")
	fmt.Println("💡 Store this in a secure password manager or safe location")

	return nil
}

// generateFixedKey generates a new fixed key for development
func generateFixedKey(ctx *cli.Context) error {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("failed to generate random key: %w", err)
	}

	keyHex := hex.EncodeToString(key)

	fmt.Printf("🔑 Generated Fixed Key:\n")
	fmt.Printf("   %s\n", keyHex)
	fmt.Printf("\n📋 Config Example:\n")
	fmt.Printf(`{
  "encryptionSettings": {
    "fixed_key": {
      "value": "%s"
    }
  }
}`, keyHex)
	fmt.Printf("\n\n💡 Or set environment variable:\n")
	fmt.Printf("   export FAMSTACK_FIXED_KEY_VALUE=%s\n", keyHex)

	return nil
}
