package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Handlers provides HTTP handlers for authentication endpoints
type Handlers struct {
	authService *Service
}

// NewHandlers creates new authentication handlers
func NewHandlers(authService *Service) *Handlers {
	return &Handlers{
		authService: authService,
	}
}

// HandleLogin handles user login requests
func (h *Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		h.writeError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate user
	authResponse, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		h.writeError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set JWT token in HTTP-only cookie
	h.setAuthCookie(w, authResponse.Token)

	// Return response (without token in JSON for security)
	response := map[string]interface{}{
		"user":        authResponse.User,
		"session":     authResponse.Session,
		"permissions": authResponse.Permissions,
	}

	h.writeJSON(w, response)
}

// HandleLogout handles user logout requests
func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the auth cookie
	h.clearAuthCookie(w)

	h.writeJSON(w, map[string]string{"message": "Logged out successfully"})
}

// HandleDowngrade handles requests to downgrade to shared mode
func (h *Handlers) HandleDowngrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current token
	token, err := h.extractToken(r)
	if err != nil {
		h.writeError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Downgrade to shared mode
	tokenResponse, err := h.authService.DowngradeToShared(token)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set new token in cookie
	h.setAuthCookie(w, tokenResponse.Token)

	// Return response
	response := map[string]interface{}{
		"session":     tokenResponse.Session,
		"permissions": tokenResponse.Permissions,
		"message":     "Switched to family mode",
	}

	h.writeJSON(w, response)
}

// HandleUpgrade handles requests to upgrade from shared mode
func (h *Handlers) HandleUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current token
	token, err := h.extractToken(r)
	if err != nil {
		h.writeError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req PasswordUpgradeRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		h.writeError(w, fmt.Sprintf("Invalid request body: %v", decodeErr), http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		h.writeError(w, "Password is required", http.StatusBadRequest)
		return
	}

	// Upgrade with password verification
	tokenResponse, err := h.authService.UpgradeWithPassword(token, req.Password)
	if err != nil {
		if err.Error() == "invalid password" {
			h.writeError(w, "Invalid password", http.StatusUnauthorized)
		} else if err.Error() == "too many upgrade attempts, please try again later" {
			h.writeError(w, err.Error(), http.StatusTooManyRequests)
		} else {
			h.writeError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	// Set new token in cookie
	h.setAuthCookie(w, tokenResponse.Token)

	// Return response
	response := map[string]interface{}{
		"session":     tokenResponse.Session,
		"permissions": tokenResponse.Permissions,
		"message":     "Switched to personal mode",
	}

	h.writeJSON(w, response)
}

// HandleRefresh handles token refresh requests
func (h *Handlers) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current token
	token, err := h.extractToken(r)
	if err != nil {
		h.writeError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Refresh token
	tokenResponse, err := h.authService.RefreshToken(token)
	if err != nil {
		h.writeError(w, "Failed to refresh token", http.StatusUnauthorized)
		return
	}

	// Set new token in cookie
	h.setAuthCookie(w, tokenResponse.Token)

	// Return response
	response := map[string]interface{}{
		"session":     tokenResponse.Session,
		"permissions": tokenResponse.Permissions,
		"message":     "Token refreshed",
	}

	h.writeJSON(w, response)
}

// HandleMe handles requests for current user info
func (h *Handlers) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current session and user from context (set by middleware)
	session := GetSessionFromContext(r.Context())
	user := GetUserFromContext(r.Context())

	if session == nil || user == nil {
		h.writeError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Return user info and permissions
	response := map[string]interface{}{
		"user":        user,
		"session":     session,
		"permissions": GetPermissionList(session.Role),
	}

	h.writeJSON(w, response)
}

// Helper methods

// extractToken extracts JWT token from Authorization header or cookie
func (h *Handlers) extractToken(r *http.Request) (string, error) {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:], nil
		}
	}

	// Try cookie
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("no token found")
}

// setAuthCookie sets the JWT token as an HTTP-only cookie
func (h *Handlers) setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	}
	http.SetCookie(w, cookie)
}

// clearAuthCookie clears the authentication cookie
func (h *Handlers) clearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Expire immediately
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

// writeJSON writes a JSON response
func (h *Handlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but can't change response at this point
		fmt.Printf("Failed to encode JSON response: %v\n", err)
	}
}

// writeError writes a JSON error response
func (h *Handlers) writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := map[string]interface{}{
		"error":   "authentication_error",
		"message": message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Failed to encode error response: %v\n", err)
	}
}
