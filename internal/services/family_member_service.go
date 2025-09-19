package services

import (
	"database/sql"
	"fmt"
	"time"

	"famstack/internal/models"
)

// FamilyMemberService handles family member operations
type FamilyMemberService struct {
	db *sql.DB
}

// NewFamilyMemberService creates a new family member service
func NewFamilyMemberService(db *sql.DB) *FamilyMemberService {
	return &FamilyMemberService{
		db: db,
	}
}

// ListFamilyMembers returns all family members for a family
func (s *FamilyMemberService) ListFamilyMembers(familyID string) ([]*models.FamilyMember, error) {
	query := `
		SELECT id, family_id, name, nickname, member_type, age,
			   avatar_url, email, role, email_verified, last_login_at,
			   display_order, is_active, created_at, updated_at
		FROM family_members
		WHERE family_id = ? AND is_active = true
		ORDER BY display_order ASC, created_at ASC
	`

	rows, err := s.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list family members: %w", err)
	}
	defer rows.Close()

	var members []*models.FamilyMember
	for rows.Next() {
		member, scanErr := s.scanFamilyMember(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan family member: %w", scanErr)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating family members: %w", err)
	}

	return members, nil
}

// GetFamilyMember returns a specific family member by ID
func (s *FamilyMemberService) GetFamilyMember(memberID string) (*models.FamilyMember, error) {
	query := `
		SELECT id, family_id, name, nickname, member_type, age,
			   avatar_url, email, role, email_verified, last_login_at,
			   display_order, is_active, created_at, updated_at
		FROM family_members
		WHERE id = ?
	`

	row := s.db.QueryRow(query, memberID)
	return s.scanFamilyMember(row)
}

// CreateFamilyMember creates a new family member
func (s *FamilyMemberService) CreateFamilyMember(familyID string, req *models.CreateFamilyMemberRequest) (*models.FamilyMember, error) {
	// Generate ID
	memberID := fmt.Sprintf("member-%d", time.Now().UnixNano())

	// Set default display order if not provided
	displayOrder := 0
	if req.DisplayOrder != nil {
		displayOrder = *req.DisplayOrder
	}

	query := `
		INSERT INTO family_members (id, family_id, name, nickname, member_type, age, avatar_url, display_order, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := s.db.Exec(query, memberID, familyID, req.Name, req.Nickname, req.MemberType, req.Age, req.AvatarURL, displayOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to create family member: %w", err)
	}

	return s.GetFamilyMember(memberID)
}

// UpdateFamilyMember updates an existing family member
func (s *FamilyMemberService) UpdateFamilyMember(memberID string, req *models.UpdateFamilyMemberRequest) (*models.FamilyMember, error) {
	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Nickname != nil {
		setParts = append(setParts, "nickname = ?")
		args = append(args, *req.Nickname)
	}
	if req.MemberType != nil {
		setParts = append(setParts, "member_type = ?")
		args = append(args, *req.MemberType)
	}
	if req.Age != nil {
		setParts = append(setParts, "age = ?")
		args = append(args, *req.Age)
	}
	if req.AvatarURL != nil {
		setParts = append(setParts, "avatar_url = ?")
		args = append(args, *req.AvatarURL)
	}
	if req.DisplayOrder != nil {
		setParts = append(setParts, "display_order = ?")
		args = append(args, *req.DisplayOrder)
	}
	if req.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *req.IsActive)
	}

	if len(setParts) == 1 { // Only updated_at
		return s.GetFamilyMember(memberID) // No changes, return current
	}

	// Add memberID to args for WHERE clause
	args = append(args, memberID)

	query := fmt.Sprintf(`
		UPDATE family_members
		SET %s
		WHERE id = ?
	`, joinStrings(setParts, ", "))

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update family member: %w", err)
	}

	return s.GetFamilyMember(memberID)
}

// DeleteFamilyMember soft deletes a family member (sets is_active = false)
func (s *FamilyMemberService) DeleteFamilyMember(memberID string) error {
	query := `UPDATE family_members SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.Exec(query, memberID)
	if err != nil {
		return fmt.Errorf("failed to delete family member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family member not found")
	}

	return nil
}

// LinkUserToFamilyMember links an existing user account to a family member
func (s *FamilyMemberService) LinkUserToFamilyMember(memberID, userID string) error {
	query := `UPDATE family_members SET user_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.Exec(query, userID, memberID)
	if err != nil {
		return fmt.Errorf("failed to link user to family member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family member not found")
	}

	return nil
}

// UnlinkUserFromFamilyMember removes the user account link from a family member
func (s *FamilyMemberService) UnlinkUserFromFamilyMember(memberID string) error {
	query := `UPDATE family_members SET user_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`

	result, err := s.db.Exec(query, memberID)
	if err != nil {
		return fmt.Errorf("failed to unlink user from family member: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("family member not found")
	}

	return nil
}

// GetFamilyMembersWithStats returns family members with task completion statistics
func (s *FamilyMemberService) GetFamilyMembersWithStats(familyID string) ([]*models.FamilyMemberWithStats, error) {
	query := `
		SELECT fm.id, fm.family_id, fm.name, fm.nickname, fm.member_type, fm.age,
			   fm.avatar_url, fm.email, fm.role, fm.email_verified, fm.last_login_at,
			   fm.display_order, fm.is_active, fm.created_at, fm.updated_at,
			   COUNT(t.id) as total_tasks,
			   COUNT(CASE WHEN t.status = 'completed' THEN 1 END) as completed_tasks,
			   COUNT(CASE WHEN t.status != 'completed' THEN 1 END) as pending_tasks
		FROM family_members fm
		LEFT JOIN tasks t ON t.assigned_to = fm.id
		WHERE fm.family_id = ? AND fm.is_active = true
		GROUP BY fm.id, fm.family_id, fm.name, fm.nickname, fm.member_type, fm.age,
				 fm.avatar_url, fm.email, fm.role, fm.email_verified, fm.last_login_at,
				 fm.display_order, fm.is_active, fm.created_at, fm.updated_at
		ORDER BY fm.display_order ASC, fm.created_at ASC
	`

	rows, err := s.db.Query(query, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to list family members with stats: %w", err)
	}
	defer rows.Close()

	var members []*models.FamilyMemberWithStats
	for rows.Next() {
		member, scanErr := s.scanFamilyMemberWithStats(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan family member with stats: %w", scanErr)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating family members with stats: %w", err)
	}

	return members, nil
}

// Helper functions

func (s *FamilyMemberService) scanFamilyMember(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.FamilyMember, error) {
	var member models.FamilyMember
	var email, role sql.NullString
	var lastLoginAt sql.NullTime

	err := scanner.Scan(
		&member.ID, &member.FamilyID, &member.Name, &member.Nickname, &member.MemberType,
		&member.Age, &member.AvatarURL, &email, &role, &member.EmailVerified, &lastLoginAt,
		&member.DisplayOrder, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if email.Valid {
		member.Email = &email.String
	}
	if role.Valid {
		member.Role = &role.String
	}
	if lastLoginAt.Valid {
		member.LastLoginAt = &lastLoginAt.Time
	}

	return &member, nil
}

func (s *FamilyMemberService) scanFamilyMemberWithStats(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.FamilyMemberWithStats, error) {
	var member models.FamilyMember
	var email, role sql.NullString
	var lastLoginAt sql.NullTime
	var totalTasks, completedTasks, pendingTasks int

	err := scanner.Scan(
		&member.ID, &member.FamilyID, &member.Name, &member.Nickname, &member.MemberType,
		&member.Age, &member.AvatarURL, &email, &role, &member.EmailVerified, &lastLoginAt,
		&member.DisplayOrder, &member.IsActive, &member.CreatedAt, &member.UpdatedAt,
		&totalTasks, &completedTasks, &pendingTasks,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if email.Valid {
		member.Email = &email.String
	}
	if role.Valid {
		member.Role = &role.String
	}
	if lastLoginAt.Valid {
		member.LastLoginAt = &lastLoginAt.Time
	}

	// Calculate completion rate
	var completionRate float64
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	stats := models.TaskStats{
		TotalTasks:     totalTasks,
		CompletedTasks: completedTasks,
		PendingTasks:   pendingTasks,
		CompletionRate: completionRate,
	}

	return &models.FamilyMemberWithStats{
		FamilyMember: member,
		TaskStats:    stats,
	}, nil
}
