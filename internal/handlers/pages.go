package handlers

import (
	"html/template"
	"net/http"

	"famstack/internal/auth"
	"famstack/internal/database"
)

// PageHandler handles all page requests
type PageHandler struct {
	db          *database.Fascade
	authService *auth.Service
}

// NewPageHandler creates a new page handler
func NewPageHandler(db *database.Fascade, authService *auth.Service) *PageHandler {
	return &PageHandler{
		db:          db,
		authService: authService,
	}
}

// PageData represents common data passed to all page templates
type PageData struct {
	CSRFToken string
	PageTitle string
	PageType  string // "tasks", "family", etc.
}

// SPAConfig represents configuration data for the SPA
type SPAConfig struct {
	APIBaseURL string          `json:"apiBaseUrl"`
	CSRFToken  string          `json:"csrfToken"`
	Features   map[string]bool `json:"features"`
}

// ServePage serves the SPA for all routes
func (h *PageHandler) ServePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Debug: Log that we're serving SPA
	println("ServePage called for:", r.URL.Path)

	// For SPA, we serve the same HTML file for all routes
	// The client-side router will handle the actual routing

	// Create config data for the SPA
	config := h.getSPAConfig(r)

	// Parse the SPA template
	tmpl, err := template.ParseFiles("web/app.html")
	if err != nil {
		http.Error(w, "SPA template error", http.StatusInternalServerError)
		return
	}

	// Execute template with config
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, config); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// getSPAConfig returns configuration data for the SPA
func (h *PageHandler) getSPAConfig(r *http.Request) SPAConfig {
	return SPAConfig{
		APIBaseURL: "/api/v1",
		CSRFToken:  "dummy-csrf-token", // TODO: Implement proper CSRF token generation
		Features: map[string]bool{
			"tasks":        true,
			"calendar":     true,
			"family":       true,
			"schedules":    true,
			"integrations": true,
		},
	}
}
