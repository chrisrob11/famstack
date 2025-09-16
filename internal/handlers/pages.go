package handlers

import (
	"html/template"
	"net/http"

	"famstack/internal/auth"
	"famstack/internal/database"
)

// PageHandler handles all page requests
type PageHandler struct {
	db          *database.DB
	authService *auth.Service
}

// NewPageHandler creates a new page handler
func NewPageHandler(db *database.DB, authService *auth.Service) *PageHandler {
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

// ServePage serves different pages based on the page type
func (h *PageHandler) ServePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract page type from URL path
	pageType := h.getPageTypeFromPath(r.URL.Path)

	// Check authentication for protected pages
	if pageType != "login" && !h.isAuthenticated(r) {
		// Redirect to login page
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// If authenticated and trying to access login page, redirect to tasks
	if pageType == "login" && h.isAuthenticated(r) {
		http.Redirect(w, r, "/tasks", http.StatusFound)
		return
	}

	// Determine template and page data
	templateName, pageData := h.getPageTemplate(pageType)

	// Parse template
	tmpl, err := template.ParseFiles(templateName)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	// Execute template
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, pageData); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// isAuthenticated checks if the user has a valid authentication token
func (h *PageHandler) isAuthenticated(r *http.Request) bool {
	// Try to extract token from cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return false
	}

	// Validate token
	_, err = h.authService.ValidateToken(cookie.Value)
	return err == nil
}

// getPageTypeFromPath extracts the page type from the URL path
func (h *PageHandler) getPageTypeFromPath(urlPath string) string {
	switch urlPath {
	case "/login":
		return "login"
	case "/tasks", "/":
		return "tasks"
	case "/schedules":
		return "schedules"
	case "/family/setup", "/family":
		return "family"
	default:
		return "tasks" // Default fallback
	}
}

// getPageTemplate returns the template path and page data for a given page type
func (h *PageHandler) getPageTemplate(pageType string) (string, PageData) {
	baseData := PageData{
		CSRFToken: "dummy-csrf-token", // TODO: Implement proper CSRF token generation
		PageType:  pageType,
	}

	switch pageType {
	case "login":
		baseData.PageTitle = "Login - FamStack"
		return "web/templates/login.html.tmpl", baseData
	case "tasks":
		baseData.PageTitle = "Daily Tasks - FamStack"
		return "web/templates/app.html.tmpl", baseData
	case "schedules":
		baseData.PageTitle = "Schedules - FamStack"
		return "web/templates/app.html.tmpl", baseData
	case "family":
		baseData.PageTitle = "Family Setup - FamStack"
		return "web/templates/app.html.tmpl", baseData
	default:
		baseData.PageTitle = "FamStack"
		return "web/templates/app.html.tmpl", baseData
	}
}
