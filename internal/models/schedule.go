package models

import "time"

// TaskSchedule represents a recurring task schedule
type TaskSchedule struct {
	ID                string     `json:"id" db:"id"`
	FamilyID          string     `json:"family_id" db:"family_id"`
	CreatedBy         string     `json:"created_by" db:"created_by"`
	Title             string     `json:"title" db:"title"`
	Description       *string    `json:"description" db:"description"`
	TaskType          string     `json:"task_type" db:"task_type"` // 'todo', 'chore', 'appointment'
	AssignedTo        *string    `json:"assigned_to" db:"assigned_to"`
	DaysOfWeek        *string    `json:"days_of_week" db:"days_of_week"` // JSON array: ["tuesday", "thursday"]
	TimeOfDay         *string    `json:"time_of_day" db:"time_of_day"`   // HH:MM format, optional specific time
	Priority          int        `json:"priority" db:"priority"`
	Points            int        `json:"points" db:"points"`
	Active            bool       `json:"active" db:"active"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	LastGeneratedDate *time.Time `json:"last_generated_date" db:"last_generated_date"`
}
