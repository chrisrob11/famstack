package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ContextKey is used for context keys to avoid collisions
type ContextKey string

const (
	// SessionContextKey is the key for storing session in context
	SessionContextKey ContextKey = "session"
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
)

// Middleware handles authentication for HTTP requests
type Middleware struct {
	authService *Service
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(authService *Service) *Middleware {
	return &Middleware{
		authService: authService,
	}
}

// RequireAuth middleware that requires valid authentication
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request
		token, err := m.extractToken(r)
		if err != nil {
			m.writeError(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Validate token and get session
		session, err := m.authService.ValidateToken(token)
		if err != nil {
			m.writeError(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Get user info
		user, err := m.authService.GetUserByToken(token)
		if err != nil {
			m.writeError(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Add to context
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		ctx = context.WithValue(ctx, UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireEntityAction middleware that requires specific entity/action permissions
func (m *Middleware) RequireEntityAction(entity Entity, action Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First ensure authentication
			m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				session := GetSessionFromContext(r.Context())
				if session == nil {
					m.writeError(w, "Authentication required", http.StatusUnauthorized)
					return
				}

				// Create authorization checker
				auth := NewAuthorizationService(session)

				// For resource-specific actions, extract owner ID from request
				var resourceOwnerID *string
				if action == ActionUpdate || action == ActionDelete {
					resourceOwnerID = m.extractResourceOwnerID(r, entity)
				}

				// Check permission
				if !auth.HasPermission(entity, action, resourceOwnerID) {
					if auth.CanUpgradeToAccess(entity, action, resourceOwnerID) {
						m.writeUpgradeRequired(w, entity, action)
						return
					}
					m.writeError(w, "Insufficient permissions", http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}

// OptionalAuth middleware that extracts auth info if present but doesn't require it
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token
		token, err := m.extractToken(r)
		if err != nil {
			// No token, proceed without auth
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate token
		session, err := m.authService.ValidateToken(token)
		if err != nil {
			// Invalid token, proceed without auth
			next.ServeHTTP(w, r)
			return
		}

		// Get user info
		user, err := m.authService.GetUserByToken(token)
		if err != nil {
			// Can't get user, proceed without auth
			next.ServeHTTP(w, r)
			return
		}

		// Add to context
		ctx := context.WithValue(r.Context(), SessionContextKey, session)
		ctx = context.WithValue(ctx, UserContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts JWT token from Authorization header or cookie
func (m *Middleware) extractToken(r *http.Request) (string, error) {
	// Try Authorization header first (Bearer token)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	// Try cookie
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("no token found")
}

// extractResourceOwnerID extracts the owner ID of a resource being accessed
func (m *Middleware) extractResourceOwnerID(r *http.Request, entity Entity) *string {
	// This would typically parse the URL path or request body to determine resource ownership
	// For now, we'll implement basic logic - can be expanded based on API design

	switch entity {
	case EntityTask:
		// For tasks, we might need to query the database to get the assigned_to field
		// This is a simplified implementation
		return nil // TODO: Implement task ownership lookup
	case EntityCalendar:
		// For calendar events, similar logic
		return nil // TODO: Implement calendar event ownership lookup
	default:
		return nil
	}
}

// writeError writes a JSON error response
func (m *Middleware) writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error":   "authentication_error",
		"message": message,
	}

	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		fmt.Printf("error ending response: %v\n", encodeErr)
	}
}

// writeUpgradeRequired writes a response indicating upgrade is needed
func (m *Middleware) writeUpgradeRequired(w http.ResponseWriter, entity Entity, action Action) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	response := map[string]interface{}{
		"error":            "upgrade_required",
		"message":          "This action requires full access. Please enter your password.",
		"entity":           entity,
		"action":           action,
		"upgrade_endpoint": "/auth/upgrade",
	}

	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		fmt.Printf("error encoding response: %v\n", encodeErr)
	}
}

// Helper functions for extracting from context

// GetSessionFromContext extracts session from request context
func GetSessionFromContext(ctx context.Context) *Session {
	if session, ok := ctx.Value(SessionContextKey).(*Session); ok {
		return session
	}
	return nil
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) *User {
	if user, ok := ctx.Value(UserContextKey).(*User); ok {
		return user
	}
	return nil
}
