package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
)

// FamilyAPIHandler handles family-related API requests
type FamilyAPIHandler struct {
	db *database.DB
}

// NewFamilyAPIHandler creates a new family API handler
func NewFamilyAPIHandler(db *database.DB) *FamilyAPIHandler {
	return &FamilyAPIHandler{
		db: db,
	}
}

// CreateFamily creates a new family
func (h *FamilyAPIHandler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON data
	var family models.Family
	if err := json.NewDecoder(r.Body).Decode(&family); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Generate secure random ID
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "Failed to generate ID", http.StatusInternalServerError)
		return
	}
	family.ID = hex.EncodeToString(bytes)

	// Basic validation
	family.Name = strings.TrimSpace(family.Name)
	if family.Name == "" {
		http.Error(w, "Family name is required", http.StatusBadRequest)
		return
	}

	// Set created timestamp
	family.CreatedAt = time.Now()

	// Insert into database
	query := `INSERT INTO families (id, name, created_at) VALUES (?, ?, ?)`
	_, err := h.db.Exec(query, family.ID, family.Name, family.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to create family", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(family); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListFamilies lists all families
func (h *FamilyAPIHandler) ListFamilies(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Query all families from database
	query := `SELECT id, name, created_at FROM families ORDER BY created_at DESC`
	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, "Failed to query families", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var families []models.Family
	for rows.Next() {
		var family models.Family
		if err := rows.Scan(&family.ID, &family.Name, &family.CreatedAt); err != nil {
			http.Error(w, "Failed to scan family", http.StatusInternalServerError)
			return
		}
		families = append(families, family)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading families", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(families); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetFamily retrieves a specific family by ID
func (h *FamilyAPIHandler) GetFamily(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract family ID from URL path
	familyID := path.Base(r.URL.Path)
	if familyID == "" || familyID == "/" {
		http.Error(w, "Family ID is required", http.StatusBadRequest)
		return
	}

	// Query family from database
	query := `SELECT id, name, created_at FROM families WHERE id = ?`
	var family models.Family
	err := h.db.QueryRow(query, familyID).Scan(&family.ID, &family.Name, &family.CreatedAt)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			http.Error(w, "Family not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to query family", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(family); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateUser creates a new family member
func (h *FamilyAPIHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON data
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Generate secure random ID
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "Failed to generate ID", http.StatusInternalServerError)
		return
	}
	user.ID = hex.EncodeToString(bytes)

	// Basic validation
	user.Name = strings.TrimSpace(user.Name)
	if user.Name == "" {
		http.Error(w, "User name is required", http.StatusBadRequest)
		return
	}

	user.Email = strings.TrimSpace(user.Email)
	// Email is now optional - trim whitespace but allow empty values

	if user.FamilyID == "" {
		user.FamilyID = "fam1" // Default family for now
	}

	if user.Role == "" {
		user.Role = "child"
	}

	if !models.IsValidUserRole(user.Role) {
		http.Error(w, "Invalid user role", http.StatusBadRequest)
		return
	}

	// Set created timestamp
	user.CreatedAt = time.Now()

	// Insert into database (note: password_hash would be set during actual registration)
	// Handle optional email - use NULL if empty
	var emailValue interface{}
	if user.Email == "" {
		emailValue = nil
	} else {
		emailValue = user.Email
	}

	query := `INSERT INTO users (id, family_id, name, email, password_hash, role, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := h.db.Exec(query, user.ID, user.FamilyID, user.Name, emailValue, "temporary_hash", user.Role, user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if user.Email != "" {
				http.Error(w, "Email already exists", http.StatusConflict)
			} else {
				http.Error(w, "User creation failed due to constraint violation", http.StatusConflict)
			}
		} else if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			http.Error(w, "Invalid family ID", http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// ListUsers lists all users in a family
func (h *FamilyAPIHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get family ID from query parameter
	familyID := r.URL.Query().Get("family_id")
	if familyID == "" {
		// Default to first family if no family_id provided (for backwards compatibility)
		familyID = "fam1"
	}

	// Query users from database
	query := `SELECT id, family_id, name, email, role, created_at FROM users WHERE family_id = ? ORDER BY role DESC, name ASC`
	rows, err := h.db.Query(query, familyID)
	if err != nil {
		http.Error(w, "Failed to query users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var email sql.NullString
		if err := rows.Scan(&user.ID, &user.FamilyID, &user.Name, &email, &user.Role, &user.CreatedAt); err != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			return
		}
		// Handle nullable email
		if email.Valid {
			user.Email = email.String
		} else {
			user.Email = ""
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
