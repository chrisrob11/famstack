package models

import (
	"fmt"
	"time"
)

// MemberType represents the type of family member
type MemberType string

const (
	MemberTypeAdult MemberType = "adult" // Parents, older teens with accounts
	MemberTypeChild MemberType = "child" // Kids, younger family members
	MemberTypePet   MemberType = "pet"   // Family pets
)

// FamilyMember represents a member of a family with optional authentication info
type FamilyMember struct {
	ID            string     `json:"id" db:"id"`
	FamilyID      string     `json:"family_id" db:"family_id"`
	Name          string     `json:"name" db:"name"`
	Nickname      *string    `json:"nickname,omitempty" db:"nickname"`
	MemberType    MemberType `json:"member_type" db:"member_type"`
	Age           *int       `json:"age,omitempty" db:"age"`
	AvatarURL     *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	Email         *string    `json:"email,omitempty" db:"email"`
	PasswordHash  *string    `json:"-" db:"password_hash"` // Never expose password hash in JSON
	Role          *string    `json:"role,omitempty" db:"role"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	DisplayOrder  int        `json:"display_order" db:"display_order"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// DisplayName returns the preferred display name for the family member
func (fm *FamilyMember) DisplayName() string {
	if fm.Nickname != nil && *fm.Nickname != "" {
		return *fm.Nickname
	}
	return fm.Name
}

// HasAccount returns true if this family member has authentication credentials
func (fm *FamilyMember) HasAccount() bool {
	return fm.Email != nil && *fm.Email != "" && fm.PasswordHash != nil && *fm.PasswordHash != ""
}

// CanLogin returns true if this family member can login (has account and is active)
func (fm *FamilyMember) CanLogin() bool {
	return fm.HasAccount() && fm.IsActive
}

// IsAdult returns true if this is an adult family member
func (fm *FamilyMember) IsAdult() bool {
	return fm.MemberType == MemberTypeAdult
}

// IsChild returns true if this is a child family member
func (fm *FamilyMember) IsChild() bool {
	return fm.MemberType == MemberTypeChild
}

// IsPet returns true if this is a pet family member
func (fm *FamilyMember) IsPet() bool {
	return fm.MemberType == MemberTypePet
}

// GetAgeDisplay returns a display-friendly age string
func (fm *FamilyMember) GetAgeDisplay() string {
	if fm.Age == nil {
		return ""
	}

	age := *fm.Age
	switch fm.MemberType {
	case MemberTypePet:
		if age == 1 {
			return "1 year old"
		}
		return fmt.Sprintf("%d years old", age)
	case MemberTypeChild:
		if age == 1 {
			return "1 year old"
		}
		return fmt.Sprintf("%d years old", age)
	case MemberTypeAdult:
		// Adults might not want age displayed
		return ""
	default:
		return ""
	}
}

// CreateFamilyMemberRequest represents a request to create a new family member
type CreateFamilyMemberRequest struct {
	Name         string     `json:"name" validate:"required,min=1,max=100"`
	Nickname     *string    `json:"nickname,omitempty" validate:"omitempty,max=50"`
	MemberType   MemberType `json:"member_type" validate:"required,oneof=adult child pet"`
	Age          *int       `json:"age,omitempty" validate:"omitempty,min=0,max=150"`
	AvatarURL    *string    `json:"avatar_url,omitempty" validate:"omitempty,url"`
	DisplayOrder *int       `json:"display_order,omitempty"`
}

// UpdateFamilyMemberRequest represents a request to update a family member
type UpdateFamilyMemberRequest struct {
	Name         *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Nickname     *string     `json:"nickname,omitempty" validate:"omitempty,max=50"`
	MemberType   *MemberType `json:"member_type,omitempty" validate:"omitempty,oneof=adult child pet"`
	Age          *int        `json:"age,omitempty" validate:"omitempty,min=0,max=150"`
	AvatarURL    *string     `json:"avatar_url,omitempty" validate:"omitempty,url"`
	DisplayOrder *int        `json:"display_order,omitempty"`
	IsActive     *bool       `json:"is_active,omitempty"`
}

// FamilyMemberWithStats represents a family member with additional statistics
type FamilyMemberWithStats struct {
	FamilyMember
	TaskStats TaskStats `json:"task_stats"`
}

// TaskStats represents task completion statistics for a family member
type TaskStats struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	PendingTasks   int     `json:"pending_tasks"`
	CompletionRate float64 `json:"completion_rate"` // Percentage
}
