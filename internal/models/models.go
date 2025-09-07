package models

import (
	"time"
)

// Family represents a family unit
type Family struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
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
	Frequency   *string    `json:"frequency" db:"frequency"` // For recurring tasks
	Points      int        `json:"points" db:"points"`       // Gamification
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
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
