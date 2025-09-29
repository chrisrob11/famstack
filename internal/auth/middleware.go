package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"famstack/internal/models"
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
			m.writeError(w, r, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Validate token and get session
		session, err := m.authService.ValidateToken(token)
		if err != nil {
			m.writeError(w, r, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Get user info
		user, err := m.authService.GetFamilyMemberByToken(token)
		if err != nil {
			m.writeError(w, r, "User not found", http.StatusUnauthorized)
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
					m.writeError(w, r, "Authentication required", http.StatusUnauthorized)
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
					m.writeError(w, r, "Insufficient permissions", http.StatusForbidden)
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
		user, err := m.authService.GetFamilyMemberByToken(token)
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
		// For tasks, extract owner ID from the task
		return extractTaskOwnerID(r)
	case EntityCalendar:
		// For calendar events, extract owner ID from the event
		return extractCalendarEventOwnerID(r)
	case EntityFamily:
		// For families, the user's family_id should match the family being accessed
		// Extract family ID from URL path (e.g., /api/v1/families/{family_id})
		path := r.URL.Path
		if strings.Contains(path, "/families/") {
			parts := strings.Split(path, "/families/")
			if len(parts) > 1 {
				familyID := strings.Split(parts[1], "/")[0]
				return &familyID
			}
		}
		return nil
	case EntitySchedule:
		// For schedules, we'll handle the ownership check in the handler
		// since it requires database access. Return nil to skip middleware-level ownership check.
		return nil
	default:
		return nil
	}
}

// writeError writes a JSON error response or redirects to login for page requests
func (m *Middleware) writeError(w http.ResponseWriter, r *http.Request, message string, status int) {
	// For page requests, redirect to login instead of returning JSON
	if status == http.StatusUnauthorized && m.isPageRequest(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

// isPageRequest determines if this is a page request vs API request
func (m *Middleware) isPageRequest(r *http.Request) bool {
	// Check if the path starts with /api/ - if so, it's an API request
	return !strings.HasPrefix(r.URL.Path, "/api/")
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
func GetUserFromContext(ctx context.Context) *models.FamilyMember {
	if user, ok := ctx.Value(UserContextKey).(*models.FamilyMember); ok {
		return user
	}
	return nil
}

// extractTaskOwnerID extracts the owner ID for a task resource
func extractTaskOwnerID(r *http.Request) *string {
	// In a real implementation, this would query the database to get the task's assigned_to field
	// For now, return nil to indicate authorization should be handled by family-level access
	// TODO: Implement database query to get task.assigned_to or task.family_id
	return nil
}

// extractCalendarEventOwnerID extracts the owner ID for a calendar event resource
func extractCalendarEventOwnerID(r *http.Request) *string {
	// In a real implementation, this would query the database to get the event's created_by field
	// For now, return nil to indicate authorization should be handled by family-level access
	// TODO: Implement database query to get event.created_by or event.family_id
	return nil
}
