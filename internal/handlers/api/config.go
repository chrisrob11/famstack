package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"famstack/internal/config"
)

// ConfigAPIHandler handles configuration API requests
type ConfigAPIHandler struct {
	configManager *config.Manager
}

// NewConfigAPIHandler creates a new config API handler
func NewConfigAPIHandler(configManager *config.Manager) *ConfigAPIHandler {
	return &ConfigAPIHandler{
		configManager: configManager,
	}
}

// GetConfig returns the current configuration
func (h *ConfigAPIHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := h.configManager.GetConfig()

	// Create response without copying the config struct (to avoid mutex copy)
	response := map[string]interface{}{
		"version":  cfg.Version,
		"server":   cfg.Server,
		"oauth":    cfg.OAuth,
		"features": cfg.Features,
	}

	// Remove sensitive data from response
	if oauth, ok := response["oauth"].(config.OAuthConfig); ok && oauth.Google != nil {
		// Create a copy of OAuth config without client secret
		googleCfg := *oauth.Google
		googleCfg.ClientSecret = ""
		response["oauth"] = config.OAuthConfig{Google: &googleCfg}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode config", http.StatusInternalServerError)
		return
	}
}

// UpdateOAuthProvider updates OAuth provider configuration
func (h *ConfigAPIHandler) UpdateOAuthProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract provider from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}
	provider := pathParts[5] // /api/v1/config/oauth/{provider}

	var req config.OAuthProvider
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ClientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	// Mark as configured if both client_id and client_secret are provided
	req.Configured = req.ClientID != "" && req.ClientSecret != ""

	// Update configuration
	if err := h.configManager.UpdateOAuthProvider(provider, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update %s config: %v", provider, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("%s OAuth configuration updated", provider),
		"provider":   provider,
		"configured": req.Configured,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetOAuthProvider returns OAuth provider configuration
func (h *ConfigAPIHandler) GetOAuthProvider(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract provider from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}
	provider := pathParts[5] // /api/v1/config/oauth/{provider}

	providerConfig, err := h.configManager.GetOAuthProvider(provider)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get %s config: %v", provider, err), http.StatusNotFound)
		return
	}

	// Remove sensitive data from response
	responseCfg := *providerConfig
	responseCfg.ClientSecret = "" // Don't send client_secret in GET requests

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseCfg); err != nil {
		http.Error(w, "Failed to encode config", http.StatusInternalServerError)
		return
	}
}

// UpdateServerConfig updates server configuration
func (h *ConfigAPIHandler) UpdateServerConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req config.ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate port
	if req.Port == "" {
		http.Error(w, "port is required", http.StatusBadRequest)
		return
	}

	if err := h.configManager.UpdateServerConfig(req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update server config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Server configuration updated",
		"config":  req,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateFeatureConfig updates feature configuration
func (h *ConfigAPIHandler) UpdateFeatureConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req config.FeatureConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.configManager.UpdateFeatureConfig(req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update feature config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Feature configuration updated",
		"config":  req,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
