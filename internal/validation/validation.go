package validation

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface for ValidationErrors
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}

	var messages []string
	for _, err := range v {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// IsEmpty returns true if there are no validation errors
func (v ValidationErrors) IsEmpty() bool {
	return len(v) == 0
}

// ToError returns nil if no errors, otherwise returns the ValidationErrors as an error
func (v ValidationErrors) ToError() error {
	if v.IsEmpty() {
		return nil
	}
	return v
}

// Validator provides methods for building validation errors
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// AddErrorf adds a validation error with formatted message
func (v *Validator) AddErrorf(field, format string, args ...any) {
	v.AddError(field, fmt.Sprintf(format, args...))
}

// Required checks if a string field is not empty
func (v *Validator) Required(fieldName, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(fieldName, fmt.Sprintf("%s is required", fieldName))
	}
}

// MinLength checks if a string field meets minimum length
func (v *Validator) MinLength(fieldName, value string, minLen int) {
	if len(strings.TrimSpace(value)) < minLen {
		v.AddErrorf(fieldName, "%s must be at least %d characters", fieldName, minLen)
	}
}

// MaxLength checks if a string field doesn't exceed maximum length
func (v *Validator) MaxLength(fieldName, value string, maxLen int) {
	if len(value) > maxLen {
		v.AddErrorf(fieldName, "%s must be no more than %d characters", fieldName, maxLen)
	}
}

// OneOf checks if a value is one of the allowed values
func (v *Validator) OneOf(fieldName, value string, allowed []string) {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return
		}
	}
	v.AddErrorf(fieldName, "%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
}

// PositiveInt checks if an integer is positive
func (v *Validator) PositiveInt(field string, value int, fieldName string) {
	if value <= 0 {
		v.AddError(field, fmt.Sprintf("%s must be positive", fieldName))
	}
}

// Errors returns the collected validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// IsValid returns true if there are no validation errors
func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

// ToError returns nil if valid, otherwise returns the validation errors as an error
func (v *Validator) ToError() error {
	return v.errors.ToError()
}

// Legacy validation functions for compatibility with HTMX handlers
var (
	ErrRequired      = errors.New("field is required")
	ErrTooLong       = errors.New("field is too long")
	ErrInvalidFormat = errors.New("field has invalid format")
)

// stringValidationOptions has a list of ways to calidate a string
type stringValidationOptions struct {
	maxLength  int
	allowEmpty bool
}

func WithMaxLen(max int) func(*stringValidationOptions) {
	return func(op *stringValidationOptions) {
		op.maxLength = max
	}
}

// StringOpt allows validations to be applied in generic
// option based way.
type StringOpt func(*stringValidationOptions)

// ValidateStr will validate a string based on the supplied options
func ValidateStr(field, value string, opts ...StringOpt) error {
	opt := stringValidationOptions{}
	for _, v := range opts {
		v(&opt)
	}
	validator := Validator{}
	if !opt.allowEmpty {
		validator.Required(field, value)
	}

	if opt.maxLength > 0 {
		validator.MaxLength(field, value, opt.maxLength)
	}

	return validator.Errors()
}

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
