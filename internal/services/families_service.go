package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"famstack/internal/database"
	"famstack/internal/models"
)

// FamiliesService handles all family database operations
type FamiliesService struct {
	db *database.Fascade
}

// NewFamiliesService creates a new families service
func NewFamiliesService(db *database.Fascade) *FamiliesService {
	return &FamiliesService{db: db}
}

// GetFamily returns a family by ID
func (s *FamiliesService) GetFamily(familyID string) (*models.Family, error) {
	query := `SELECT id, name, timezone, created_at FROM families WHERE id = ?`

	var family models.Family
	err := s.db.QueryRow(query, familyID).Scan(
		&family.ID, &family.Name, &family.Timezone, &family.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("family not found")
		}
		return nil, fmt.Errorf("failed to get family: %w", err)
	}

	return &family, nil
}

// ListFamilies returns all families (mainly for admin purposes)
func (s *FamiliesService) ListFamilies() ([]models.Family, error) {
	query := `SELECT id, name, timezone, created_at FROM families ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list families: %w", err)
	}
	defer rows.Close()

	var families []models.Family
	for rows.Next() {
		var family models.Family
		if scanErr := rows.Scan(&family.ID, &family.Name, &family.Timezone, &family.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("failed to scan family: %w", scanErr)
		}
		families = append(families, family)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating families: %w", err)
	}

	return families, nil
}

// CreateFamily creates a new family
func (s *FamiliesService) CreateFamily(name string) (*models.Family, error) {
	familyID := generateFamilyID()
	now := time.Now().UTC()

	// Use server's local timezone as default for new families
	serverTimezone := time.Now().Location().String()

	query := `INSERT INTO families (id, name, timezone, created_at) VALUES (?, ?, ?, ?)`

	_, err := s.db.Exec(query, familyID, name, serverTimezone, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create family: %w", err)
	}

	return &models.Family{
		ID:        familyID,
		Name:      name,
		Timezone:  serverTimezone,
		CreatedAt: now,
	}, nil
}

// UpdateFamily updates a family's information
func (s *FamiliesService) UpdateFamily(familyID string, req *models.UpdateFamilyRequest) (*models.Family, error) {
	// Check if there are any changes to make
	if req.Name == nil && req.Timezone == nil {
		return s.GetFamily(familyID) // No changes
	}

	// Validate timezone if provided
	if req.Timezone != nil {
		if err := validateTimezone(*req.Timezone); err != nil {
			return nil, fmt.Errorf("invalid timezone: %w", err)
		}
	}

	// Build dynamic query based on what fields are being updated
	var setParts []string
	var args []any

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}

	if req.Timezone != nil {
		setParts = append(setParts, "timezone = ?")
		args = append(args, *req.Timezone)
	}

	// Add familyID to args for the WHERE clause
	args = append(args, familyID)

	query := fmt.Sprintf("UPDATE families SET %s WHERE id = ?", strings.Join(setParts, ", "))

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update family: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("family not found")
	}

	return s.GetFamily(familyID)
}

// DeleteFamily deletes a family (and all associated data via CASCADE)
func (s *FamiliesService) DeleteFamily(familyID string) error {
	query := `DELETE FROM families WHERE id = ?`

	result, err := s.db.Exec(query, familyID)
	if err != nil {
		return fmt.Errorf("failed to delete family: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family not found")
	}

	return nil
}

// GetFamilyStatistics returns statistics about a family
func (s *FamiliesService) GetFamilyStatistics(familyID string) (*models.FamilyStatistics, error) {
	query := `
		SELECT
			COUNT(DISTINCT fm.id) as member_count,
			COUNT(DISTINCT t.id) as total_tasks,
			COUNT(DISTINCT CASE WHEN t.status = 'completed' THEN t.id END) as completed_tasks,
			COUNT(DISTINCT CASE WHEN t.status != 'completed' THEN t.id END) as pending_tasks
		FROM families f
		LEFT JOIN family_members fm ON f.id = fm.family_id AND fm.is_active = true
		LEFT JOIN tasks t ON f.id = t.family_id
		WHERE f.id = ?
		GROUP BY f.id
	`

	var stats models.FamilyStatistics
	var memberCount, totalTasks, completedTasks, pendingTasks int

	err := s.db.QueryRow(query, familyID).Scan(
		&memberCount, &totalTasks, &completedTasks, &pendingTasks,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Family exists but has no members/tasks
			return &models.FamilyStatistics{
				FamilyID:       familyID,
				MemberCount:    0,
				TotalTasks:     0,
				CompletedTasks: 0,
				PendingTasks:   0,
				CompletionRate: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get family statistics: %w", err)
	}

	// Calculate completion rate
	var completionRate float64
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	stats = models.FamilyStatistics{
		FamilyID:       familyID,
		MemberCount:    memberCount,
		TotalTasks:     totalTasks,
		CompletedTasks: completedTasks,
		PendingTasks:   pendingTasks,
		CompletionRate: completionRate,
	}

	return &stats, nil
}

// Helper functions

func generateFamilyID() string {
	return fmt.Sprintf("fam_%d", time.Now().UTC().UnixNano())
}

// validateTimezone checks if a timezone string is valid
func validateTimezone(timezone string) error {
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}

	// Try to load the timezone location to validate it
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	return nil
}

// GetFamilyTimezone retrieves the timezone for a family using the provided database connection
// This is a utility function that can be used by other services
func GetFamilyTimezone(db *database.Fascade, familyID string) (string, error) {
	query := `SELECT timezone FROM families WHERE id = ?`
	var timezone sql.NullString

	err := db.QueryRow(query, familyID).Scan(&timezone)
	if err != nil {
		if err == sql.ErrNoRows {
			return "UTC", nil // Default to UTC if family not found
		}
		return "", fmt.Errorf("failed to get family timezone: %w", err)
	}

	if !timezone.Valid || timezone.String == "" {
		return "UTC", nil // Default to UTC if timezone is null or empty
	}

	return timezone.String, nil
}
