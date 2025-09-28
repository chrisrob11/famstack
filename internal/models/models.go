package models

import (
	"html"
	"strings"
	"time"

	"famstack/internal/validation"
)

// Family represents a family unit
type Family struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Timezone  string    `json:"timezone" db:"timezone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// User represents a family member
type User struct {
	ID           string    `json:"id" db:"id"`
	FamilyID     string    `json:"family_id" db:"family_id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Task represents a unified task (todo, chore, appointment)
type Task struct {
	ID          string     `json:"id" db:"id"`
	FamilyID    string     `json:"family_id" db:"family_id"`
	AssignedTo  *string    `json:"assigned_to" db:"assigned_to"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	TaskType    string     `json:"task_type" db:"task_type"` // 'todo', 'chore', 'appointment'
	Status      string     `json:"status" db:"status"`       // 'pending', 'completed'
	Priority    int        `json:"priority" db:"priority"`
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	FamilyID  string    `json:"family_id" db:"family_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TaskType constants
const (
	TaskTypeTodo        = "todo"
	TaskTypeChore       = "chore"
	TaskTypeAppointment = "appointment"
)

// TaskStatus constants
const (
	TaskStatusPending   = "pending"
	TaskStatusCompleted = "completed"
)

// UserRole constants
const (
	RoleParent = "parent"
	RoleChild  = "child"
	RoleAdmin  = "admin"
)

// IsValidTaskType checks if a task type is valid
func IsValidTaskType(taskType string) bool {
	switch taskType {
	case TaskTypeTodo, TaskTypeChore, TaskTypeAppointment:
		return true
	default:
		return false
	}
}

// IsValidTaskStatus checks if a task status is valid
func IsValidTaskStatus(status string) bool {
	switch status {
	case TaskStatusPending, TaskStatusCompleted:
		return true
	default:
		return false
	}
}

// IsValidUserRole checks if a user role is valid
func IsValidUserRole(role string) bool {
	switch role {
	case RoleParent, RoleChild, RoleAdmin:
		return true
	default:
		return false
	}
}

// IsValidTimezone checks if a timezone is valid using Go's time.LoadLocation
func IsValidTimezone(timezone string) bool {
	if timezone == "" {
		return false
	}
	_, err := time.LoadLocation(timezone)
	return err == nil
}

// Validate validates the task and returns validation errors
func (t *Task) Validate() error {
	validator := validation.NewValidator()

	// Validate title
	validator.Required("title", t.Title)
	validator.MinLength("title", t.Title, 1)
	validator.MaxLength("title", t.Title, 255)

	// Validate description (optional but if provided, has max length)
	validator.MaxLength("description", t.Description, 1000)

	// Validate status
	validator.Required("status", t.Status)
	if !IsValidTaskStatus(t.Status) {
		validator.AddError("status", "Status must be 'pending' or 'completed'")
	}

	// Validate task type
	validator.Required("task_type", t.TaskType)
	if !IsValidTaskType(t.TaskType) {
		validator.AddError("task_type", "Task type must be 'todo', 'chore', or 'appointment'")
	}

	// Validate family ID
	validator.Required("family_id", t.FamilyID)
	validator.MaxLength("family_id", t.FamilyID, 50)

	// Validate created by
	validator.Required("created_by", t.CreatedBy)

	return validator.ToError()
}

// Sanitize cleans and sanitizes the task input
func (t *Task) Sanitize() {
	// Trim and escape HTML in title and description
	t.Title = html.EscapeString(strings.TrimSpace(t.Title))
	t.Description = html.EscapeString(strings.TrimSpace(t.Description))

	// Normalize status and task type to lowercase
	t.Status = strings.ToLower(strings.TrimSpace(t.Status))
	t.TaskType = strings.ToLower(strings.TrimSpace(t.TaskType))

	// Trim other string fields
	t.FamilyID = strings.TrimSpace(t.FamilyID)
	t.CreatedBy = strings.TrimSpace(t.CreatedBy)
	if t.AssignedTo != nil {
		assigned := strings.TrimSpace(*t.AssignedTo)
		t.AssignedTo = &assigned
	}
}

// SetDefaults sets default values for optional fields
func (t *Task) SetDefaults() {
	if t.Status == "" {
		t.Status = TaskStatusPending
	}

	if t.TaskType == "" {
		t.TaskType = TaskTypeTodo
	}

	// Note: FamilyID should be set explicitly by the caller, not defaulted

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}

	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now().UTC()
	}

	if t.Priority == 0 {
		t.Priority = 1 // Default priority
	}
}

// PrepareForCreate sanitizes, sets defaults, and validates a task for creation
func (t *Task) PrepareForCreate() error {
	t.Sanitize()
	t.SetDefaults()
	return t.Validate()
}

// CanComplete returns true if the task can be completed
func (t *Task) CanComplete() bool {
	return t.Status == TaskStatusPending
}

// CanReopen returns true if the task can be reopened
func (t *Task) CanReopen() bool {
	return t.Status == TaskStatusCompleted
}

// Complete marks the task as completed
func (t *Task) Complete() error {
	if !t.CanComplete() {
		return &validation.ValidationErrors{
			{Field: "status", Message: "Task is already completed or cannot be completed"},
		}
	}

	t.Status = TaskStatusCompleted
	now := time.Now().UTC()
	t.CompletedAt = &now
	t.UpdatedAt = now

	return nil
}

// Reopen marks the task as pending
func (t *Task) Reopen() error {
	if !t.CanReopen() {
		return &validation.ValidationErrors{
			{Field: "status", Message: "Task is not completed or cannot be reopened"},
		}
	}

	t.Status = TaskStatusPending
	t.CompletedAt = nil
	t.UpdatedAt = time.Now().UTC()

	return nil
}

// Validate validates the family and returns validation errors
func (f *Family) Validate() error {
	validator := validation.NewValidator()

	// Validate name
	validator.Required("name", f.Name)
	validator.MinLength("name", f.Name, 1)
	validator.MaxLength("name", f.Name, 255)

	// Validate timezone
	validator.Required("timezone", f.Timezone)
	if !IsValidTimezone(f.Timezone) {
		validator.AddError("timezone", "Invalid timezone identifier")
	}

	return validator.ToError()
}

// SetDefaults sets default values for optional fields
func (f *Family) SetDefaults() {
	if f.Timezone == "" {
		f.Timezone = "UTC"
	}

	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
}

// GetLocation returns the time.Location for this family's timezone
func (f *Family) GetLocation() (*time.Location, error) {
	return time.LoadLocation(f.Timezone)
}
