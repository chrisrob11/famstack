package services

import (
	"famstack/internal/database"
	"famstack/internal/encryption"
	"famstack/internal/integrations"
)

// Registry provides centralized access to all services
type Registry struct {
	// Database services
	Tasks         *TasksService
	Families      *FamiliesService
	FamilyMembers *FamilyMemberService
	Calendar      *CalendarService
	Schedules     *SchedulesService

	// External services (existing)
	Integrations *integrations.Service
}

// NewRegistry creates a new service registry with all services initialized
func NewRegistry(db *database.DB, encryptionSvc *encryption.Service) *Registry {
	return &Registry{
		// Database services (using raw sql.DB)
		Tasks:         NewTasksService(db.DB),
		Families:      NewFamiliesService(db.DB),
		FamilyMembers: NewFamilyMemberService(db.DB),
		Calendar:      NewCalendarService(db.DB),
		Schedules:     NewSchedulesService(db.DB),

		// External services (using wrapped database.DB)
		Integrations: integrations.NewService(db, encryptionSvc),
	}
}
