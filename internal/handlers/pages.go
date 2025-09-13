package handlers

import (
	"html/template"
	"net/http"

	"famstack/internal/database"
)

// PageHandler handles all page requests
type PageHandler struct {
	db *database.DB
}

// NewPageHandler creates a new page handler
func NewPageHandler(db *database.DB) *PageHandler {
	return &PageHandler{
		db: db,
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

// getPageTypeFromPath extracts the page type from the URL path
func (h *PageHandler) getPageTypeFromPath(urlPath string) string {
	switch urlPath {
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
