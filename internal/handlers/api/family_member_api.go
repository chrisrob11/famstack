package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"famstack/internal/auth"
	"famstack/internal/models"
	"famstack/internal/services"
)

// FamilyMemberAPIHandler handles HTTP requests for family member management
type FamilyMemberAPIHandler struct {
	service *services.FamilyMemberService
}

// NewFamilyMemberAPIHandler creates a new family member API handler
func NewFamilyMemberAPIHandler(service *services.FamilyMemberService) *FamilyMemberAPIHandler {
	return &FamilyMemberAPIHandler{
		service: service,
	}
}

// ListFamilyMembers handles GET /api/families/{family_id}/members
func (h *FamilyMemberAPIHandler) ListFamilyMembers(w http.ResponseWriter, r *http.Request) {
	// Extract family ID from URL path
	familyID := h.extractFamilyIDFromPath(r.URL.Path)
	if familyID == "" {
		http.Error(w, "Family ID is required", http.StatusBadRequest)
		return
	}

	// Verify user has access to this family
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// For now, ensure user can only access their own family
	if session.FamilyID != familyID {
		http.Error(w, "Access denied to this family", http.StatusForbidden)
		return
	}

	// List family members
	members, err := h.service.ListFamilyMembers(familyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list family members: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"family_members": members,
	})
}

// GetFamilyMember handles GET /api/v1/families/members/{member_id}
func (h *FamilyMemberAPIHandler) GetFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Extract member ID from URL path
	memberID := h.extractIDFromPath(r.URL.Path, "/api/v1/families/members/")
	if memberID == "" {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Get family member
	member, err := h.service.GetFamilyMember(memberID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Family member not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get family member: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Verify family access
	session := auth.GetSessionFromContext(r.Context())
	if session == nil || session.FamilyID != member.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"family_member": member,
	})
}

// CreateFamilyMember handles POST /api/v1/families/members
func (h *FamilyMemberAPIHandler) CreateFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Extract family ID from session context
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req models.CreateFamilyMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create family member
	member, err := h.service.CreateFamilyMember(session.FamilyID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create family member: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"family_member": member,
		"message":       "Family member created successfully",
	})
}

// UpdateFamilyMember handles PATCH /api/v1/families/members/{member_id}
func (h *FamilyMemberAPIHandler) UpdateFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Extract member ID from URL path
	memberID := h.extractIDFromPath(r.URL.Path, "/api/v1/families/members/")
	if memberID == "" {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Verify family access
	member, err := h.service.GetFamilyMember(memberID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Family member not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get family member: %v", err), http.StatusInternalServerError)
		}
		return
	}

	session := auth.GetSessionFromContext(r.Context())
	if session == nil || session.FamilyID != member.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse request body
	var req models.UpdateFamilyMemberRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update family member
	updatedMember, err := h.service.UpdateFamilyMember(memberID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update family member: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"family_member": updatedMember,
		"message":       "Family member updated successfully",
	})
}

// DeleteFamilyMember handles DELETE /api/v1/families/members/{member_id}
func (h *FamilyMemberAPIHandler) DeleteFamilyMember(w http.ResponseWriter, r *http.Request) {
	// Extract member ID from URL path
	memberID := h.extractIDFromPath(r.URL.Path, "/api/v1/families/members/")
	if memberID == "" {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Verify family access
	member, err := h.service.GetFamilyMember(memberID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Family member not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get family member: %v", err), http.StatusInternalServerError)
		}
		return
	}

	session := auth.GetSessionFromContext(r.Context())
	if session == nil || session.FamilyID != member.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete family member
	if err := h.service.DeleteFamilyMember(memberID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete family member: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"message": "Family member deleted successfully",
	})
}

// GetFamilyMembersWithStats handles GET /api/v1/families/members?stats=true
func (h *FamilyMemberAPIHandler) GetFamilyMembersWithStats(w http.ResponseWriter, r *http.Request) {
	// Check if stats are requested
	if r.URL.Query().Get("stats") != "true" {
		h.ListFamilyMembers(w, r)
		return
	}

	// Extract family ID from session context
	session := auth.GetSessionFromContext(r.Context())
	if session == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get family members with stats
	members, err := h.service.GetFamilyMembersWithStats(session.FamilyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get family members with stats: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"family_members": members,
	})
}

// LinkUserToMember handles POST /api/v1/families/members/{member_id}/link-user
func (h *FamilyMemberAPIHandler) LinkUserToMember(w http.ResponseWriter, r *http.Request) {
	// Extract member ID from URL path
	memberID := h.extractIDFromPath(r.URL.Path, "/api/v1/families/members/")
	if memberID == "" {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify family access
	member, err := h.service.GetFamilyMember(memberID)
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	session := auth.GetSessionFromContext(r.Context())
	if session == nil || session.FamilyID != member.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Link user to member
	if err := h.service.LinkUserToFamilyMember(memberID, req.UserID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to link user: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"message": "User linked to family member successfully",
	})
}

// UnlinkUserFromMember handles POST /api/v1/families/members/{member_id}/unlink-user
func (h *FamilyMemberAPIHandler) UnlinkUserFromMember(w http.ResponseWriter, r *http.Request) {
	// Extract member ID from URL path
	memberID := h.extractIDFromPath(r.URL.Path, "/api/v1/families/members/")
	if memberID == "" {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Verify family access
	member, err := h.service.GetFamilyMember(memberID)
	if err != nil {
		http.Error(w, "Family member not found", http.StatusNotFound)
		return
	}

	session := auth.GetSessionFromContext(r.Context())
	if session == nil || session.FamilyID != member.FamilyID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Unlink user from member
	if err := h.service.UnlinkUserFromFamilyMember(memberID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unlink user: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"message": "User unlinked from family member successfully",
	})
}

// Helper methods

func (h *FamilyMemberAPIHandler) extractIDFromPath(path, prefix string) string {
	// Handle both v1 and non-v1 paths for member ID extraction
	var actualPrefix string
	if strings.Contains(path, "/api/v1/families/") && !strings.HasPrefix(prefix, "/api/v1/") {
		// Convert non-v1 prefix to v1 prefix for compatibility
		actualPrefix = strings.Replace(prefix, "/api/families/", "/api/v1/families/", 1)
	} else {
		actualPrefix = prefix
	}

	if !strings.HasPrefix(path, actualPrefix) {
		return ""
	}

	id := strings.TrimPrefix(path, actualPrefix)
	// Remove any trailing slashes or path segments
	if slashIndex := strings.Index(id, "/"); slashIndex != -1 {
		id = id[:slashIndex]
	}

	return id
}

func (h *FamilyMemberAPIHandler) extractFamilyIDFromPath(path string) string {
	// Handle both formats:
	// /api/families/{family_id}/members
	// /api/v1/families/{family_id}/members
	var prefix string
	if strings.HasPrefix(path, "/api/v1/families/") {
		prefix = "/api/v1/families/"
	} else if strings.HasPrefix(path, "/api/families/") {
		prefix = "/api/families/"
	} else {
		return ""
	}

	// Remove the prefix
	remaining := strings.TrimPrefix(path, prefix)

	// Extract family ID (everything before the next slash)
	if slashIndex := strings.Index(remaining, "/"); slashIndex != -1 {
		return remaining[:slashIndex]
	}

	// If no slash found, the entire remaining string is the family ID
	return remaining
}

func (h *FamilyMemberAPIHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Failed to encode JSON response: %v\n", err)
	}
}
