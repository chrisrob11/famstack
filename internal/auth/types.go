package auth

import (
	"famstack/internal/models"
	"time"
)

// Role represents user permission levels
type Role string

const (
	RoleShared Role = "shared" // Downgraded mode with minimal permissions
	RoleUser   Role = "user"   // Standard family member
	RoleAdmin  Role = "admin"  // Family administrator
)

// Entity represents resources that can be acted upon
type Entity string

const (
	EntityTask     Entity = "task"
	EntityFamily   Entity = "family"
	EntityUser     Entity = "user"
	EntityCalendar Entity = "calendar"
	EntitySchedule Entity = "schedule"
	EntitySetting  Entity = "setting"
)

// Action represents operations that can be performed
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// PermissionScope defines the scope of access
type PermissionScope string

const (
	ScopeNone PermissionScope = "none" // No access
	ScopeOwn  PermissionScope = "own"  // Only items owned/assigned to user
	ScopeAny  PermissionScope = "any"  // All items in family
)

// Permission represents a specific permission string
type Permission string

// Session represents an authenticated session (now derived from JWT)
type Session struct {
	UserID       string    `json:"user_id"`
	FamilyID     string    `json:"family_id"`
	Role         Role      `json:"role"`
	OriginalRole Role      `json:"original_role"`
	ExpiresAt    time.Time `json:"expires_at"`
	IssuedAt     time.Time `json:"issued_at"`
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

// CanDowngrade checks if the session can be downgraded to shared mode
func (s *Session) CanDowngrade() bool {
	return s.Role != RoleShared
}

// CanUpgrade checks if the session can be upgraded from shared mode
func (s *Session) CanUpgrade() bool {
	return s.Role == RoleShared
}

// FromJWTClaims creates a Session from JWT claims
func SessionFromJWTClaims(claims *JWTClaims) *Session {
	return &Session{
		UserID:       claims.UserID,
		FamilyID:     claims.FamilyID,
		Role:         claims.Role,
		OriginalRole: claims.OriginalRole,
		ExpiresAt:    claims.ExpiresAt.Time,
		IssuedAt:     claims.IssuedAt.Time,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// PasswordUpgradeRequest represents a password challenge for upgrading permissions
type PasswordUpgradeRequest struct {
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after authentication
type AuthResponse struct {
	User        *models.FamilyMember `json:"user"`
	Session     *Session             `json:"session"`
	Token       string               `json:"token"`
	Permissions []string             `json:"permissions"`
}

// TokenResponse represents a response containing just a token
type TokenResponse struct {
	Token       string   `json:"token"`
	Session     *Session `json:"session"`
	Permissions []string `json:"permissions"`
}

// CreateUserRequest represents a request to create a new user with auth details
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Role      Role   `json:"role" validate:"required"`
	FamilyID  string `json:"family_id" validate:"required"`
}
