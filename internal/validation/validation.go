package validation

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrRequired      = errors.New("field is required")
	ErrTooLong       = errors.New("field is too long")
	ErrInvalidFormat = errors.New("field has invalid format")
	ErrUserNotFound  = errors.New("user not found")
)

// ValidateTitle validates task title input
func ValidateTitle(title string) error {
	title = strings.TrimSpace(title)

	if title == "" {
		return ErrRequired
	}

	if utf8.RuneCountInString(title) > 255 {
		return ErrTooLong
	}

	// Remove potentially dangerous characters
	if strings.ContainsAny(title, "<>\"'&") {
		return ErrInvalidFormat
	}

	return nil
}

// ValidateUserID validates user ID format
func ValidateUserID(userID string) error {
	userID = strings.TrimSpace(userID)

	if userID == "" || userID == "unassigned" {
		return nil // These are valid
	}

	// Basic format validation - alphanumeric with underscores
	for _, r := range userID {
		if !isValidUserIDChar(r) {
			return ErrInvalidFormat
		}
	}

	if len(userID) > 50 {
		return ErrTooLong
	}

	return nil
}

// SanitizeTitle cleans and trims title input
func SanitizeTitle(title string) string {
	title = strings.TrimSpace(title)

	// Remove line breaks and excessive whitespace
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.ReplaceAll(title, "\r", " ")
	title = strings.ReplaceAll(title, "\t", " ")

	// Collapse multiple spaces
	for strings.Contains(title, "  ") {
		title = strings.ReplaceAll(title, "  ", " ")
	}

	return title
}

// isValidUserIDChar checks if a character is valid for user IDs
func isValidUserIDChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}
