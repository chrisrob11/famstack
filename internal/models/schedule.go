package models

import "time"

// TaskSchedule represents a recurring task schedule
type TaskSchedule struct {
	ID              string     `json:"id" db:"id"`
	FamilyID        string     `json:"family_id" db:"family_id"`
	Name            string     `json:"name" db:"name"`
	Description     *string    `json:"description" db:"description"`
	TaskTemplate    string     `json:"task_template" db:"task_template"` // JSON template for task creation
	AssignedTo      *string    `json:"assigned_to" db:"assigned_to"`
	ScheduleType    string     `json:"schedule_type" db:"schedule_type"` // 'cron', 'interval', 'weekly', 'daily'
	CronExpression  *string    `json:"cron_expression" db:"cron_expression"`
	IntervalMinutes *int       `json:"interval_minutes" db:"interval_minutes"`
	DaysOfWeek      *string    `json:"days_of_week" db:"days_of_week"` // Comma-separated: "mon,wed,fri"
	StartDate       time.Time  `json:"start_date" db:"start_date"`
	EndDate         *time.Time `json:"end_date" db:"end_date"`
	NextRunAt       time.Time  `json:"next_run_at" db:"next_run_at"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedBy       string     `json:"created_by" db:"created_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// ScheduleType constants
const (
	ScheduleTypeCron     = "cron"
	ScheduleTypeInterval = "interval"
	ScheduleTypeWeekly   = "weekly"
	ScheduleTypeDaily    = "daily"
)

// IsValidScheduleType checks if a schedule type is valid
func IsValidScheduleType(scheduleType string) bool {
	switch scheduleType {
	case ScheduleTypeCron, ScheduleTypeInterval, ScheduleTypeWeekly, ScheduleTypeDaily:
		return true
	default:
		return false
	}
}
