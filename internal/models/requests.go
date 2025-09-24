package models

import "time"

// Request/Response models for APIs

// Task request models
type CreateTaskRequest struct {
	Title       string     `json:"title" validate:"required,min=1,max=255"`
	Description string     `json:"description" validate:"max=1000"`
	TaskType    string     `json:"task_type" validate:"required,oneof=todo chore appointment"`
	AssignedTo  *string    `json:"assigned_to"`
	Priority    int        `json:"priority" validate:"min=0,max=10"`
	DueDate     *time.Time `json:"due_date"`
	Points      int        `json:"points" validate:"min=0"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	Status      *string    `json:"status,omitempty" validate:"omitempty,oneof=pending completed"`
	AssignedTo  *string    `json:"assigned_to,omitempty"`
	Priority    *int       `json:"priority,omitempty" validate:"omitempty,min=0,max=10"`
	DueDate     *time.Time `json:"due_date,omitempty"`
}

// Family request models
type UpdateFamilyRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Timezone *string `json:"timezone,omitempty" validate:"omitempty"`
}

// Calendar event request models
type CreateCalendarEventRequest struct {
	Title       string    `json:"title" validate:"required,min=1,max=255"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     time.Time `json:"end_time" validate:"required"`
	Location    *string   `json:"location,omitempty" validate:"omitempty,max=255"`
	EventType   string    `json:"event_type" validate:"required,oneof=appointment event reminder"`
	AssignedTo  *string   `json:"assigned_to,omitempty"`
}

type UpdateCalendarEventRequest struct {
	Title       *string    `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Location    *string    `json:"location,omitempty" validate:"omitempty,max=255"`
	EventType   *string    `json:"event_type,omitempty" validate:"omitempty,oneof=appointment event reminder"`
	AssignedTo  *string    `json:"assigned_to,omitempty"`
}

// Unified calendar event request models
type CreateUnifiedCalendarEventRequest struct {
	FamilyID        string    `json:"family_id" validate:"required"`
	IntegrationID   string    `json:"integration_id" validate:"required"`
	ExternalEventID string    `json:"external_event_id" validate:"required"`
	Title           string    `json:"title" validate:"required,min=1,max=255"`
	Description     *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	StartTime       time.Time `json:"start_time" validate:"required"`
	EndTime         time.Time `json:"end_time" validate:"required"`
	Location        *string   `json:"location,omitempty" validate:"omitempty,max=255"`
	Organizer       *string   `json:"organizer,omitempty" validate:"omitempty,max=255"`
	Attendees       *string   `json:"attendees,omitempty" validate:"omitempty,max=1000"`
}

// Task schedule request models
type CreateTaskScheduleRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=255"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	TaskType    string   `json:"task_type" validate:"required,oneof=todo chore appointment"`
	AssignedTo  *string  `json:"assigned_to,omitempty"`
	DaysOfWeek  []string `json:"days_of_week" validate:"required,min=1"`
	TimeOfDay   *string  `json:"time_of_day,omitempty"`
	Priority    int      `json:"priority" validate:"min=0,max=3"`
	FamilyID    *string  `json:"family_id,omitempty"`
}

type UpdateTaskScheduleRequest struct {
	Title       *string   `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=1000"`
	TaskType    *string   `json:"task_type,omitempty" validate:"omitempty,oneof=todo chore appointment"`
	AssignedTo  *string   `json:"assigned_to,omitempty"`
	DaysOfWeek  *[]string `json:"days_of_week,omitempty"`
	TimeOfDay   *string   `json:"time_of_day,omitempty"`
	Priority    *int      `json:"priority,omitempty" validate:"omitempty,min=0,max=3"`
	Active      *bool     `json:"active,omitempty"`
}
