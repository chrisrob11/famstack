package services

import (
	"famstack/internal/database"
	"famstack/internal/encryption"
)

// Registry provides centralized access to all services
type Registry struct {
	// Database services
	Tasks         *TasksService
	Families      *FamiliesService
	FamilyMembers *FamilyMemberService
	Calendar      *CalendarService
	Schedules     *SchedulesService
	OAuth         *OAuthService
	Jobs          *JobsService
	Integrations  *IntegrationsService

	// Internal references
	db            *database.Fascade
	encryptionSvc *encryption.Service
}

// NewRegistry creates a new service registry with all services initialized
func NewRegistry(db *database.Fascade, encryptionSvc *encryption.Service) *Registry {
	return &Registry{
		// Database services (using database facade)
		Tasks:         NewTasksService(db),
		Families:      NewFamiliesService(db),
		FamilyMembers: NewFamilyMemberService(db),
		Calendar:      NewCalendarService(db),
		Schedules:     NewSchedulesService(db),
		OAuth:         NewOAuthService(db),
		Jobs:          NewJobsService(db),

		// External services (using database facade)
		Integrations: NewIntegrationsService(db, encryptionSvc),

		// Keep references for legacy access
		db:            db,
		encryptionSvc: encryptionSvc,
	}
}

// GetDB returns the database facade for legacy handlers
func (r *Registry) GetDB() *database.Fascade {
	return r.db
}

// GetEncryptionService returns the encryption service
func (r *Registry) GetEncryptionService() *encryption.Service {
	return r.encryptionSvc
}
