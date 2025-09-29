package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"famstack/internal/config"
	"famstack/internal/database"
	"famstack/internal/encryption"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationTestDB(t *testing.T) (*database.Fascade, *encryption.Service) {
	dbFile := fmt.Sprintf("test_integrations_%d.db", time.Now().UnixNano())
	db, err := database.New(dbFile)
	require.NoError(t, err)

	err = db.MigrateUp()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
		os.Remove(dbFile)
	})

	// Create a test encryption service
	encryptionConfig := config.EncryptionSettings{
		FixedKey: &config.FixedKeyConfig{
			Value: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
	}
	encryptionSvc, err := encryption.NewService(encryptionConfig)
	require.NoError(t, err)

	return db, encryptionSvc
}

func setupTestFamily(t *testing.T, db *database.Fascade) (familyID, userID string) {
	familyID = fmt.Sprintf("fam_test_%d", time.Now().UnixNano())
	userID = fmt.Sprintf("user_test_%d", time.Now().UnixNano())

	// Create test family
	_, err := db.Exec(`INSERT INTO families (id, name, timezone) VALUES (?, ?, ?)`,
		familyID, "Test Family", "UTC")
	require.NoError(t, err)

	// Create test user
	_, err = db.Exec(`INSERT INTO family_members (id, family_id, first_name, last_name, member_type, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, familyID, "Test", "User", "child", true, time.Now(), time.Now())
	require.NoError(t, err)

	return familyID, userID
}

func TestIntegrationsService_CreateIntegration(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	tests := []struct {
		name    string
		request *CreateIntegrationRequest
	}{
		{
			name: "successful calendar integration creation",
			request: &CreateIntegrationRequest{
				IntegrationType: TypeCalendar,
				Provider:        ProviderGoogle,
				AuthMethod:      AuthOAuth2,
				DisplayName:     "My Google Calendar",
				Description:     "Personal Google Calendar",
				Settings: map[string]any{
					"sync_frequency_minutes": 30,
					"sync_range_days":        7,
					"sync_all_day_events":    true,
				},
				SettingsType: "CalendarSyncConfig",
			},
		},
		{
			name: "integration with custom type",
			request: &CreateIntegrationRequest{
				IntegrationType: "custom_type",
				Provider:        ProviderGoogle,
				AuthMethod:      AuthOAuth2,
				DisplayName:     "Custom Integration",
			},
		},
		{
			name: "integration with empty display name",
			request: &CreateIntegrationRequest{
				IntegrationType: TypeCalendar,
				Provider:        ProviderGoogle,
				AuthMethod:      AuthOAuth2,
				DisplayName:     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration, err := service.CreateIntegration(familyID, userID, tt.request)

			assert.NoError(t, err)
			assert.NotNil(t, integration)
			assert.NotEmpty(t, integration.ID)
			assert.Equal(t, familyID, integration.FamilyID)
			assert.Equal(t, userID, integration.CreatedBy)
			assert.Equal(t, tt.request.IntegrationType, integration.IntegrationType)
			assert.Equal(t, tt.request.Provider, integration.Provider)
			assert.Equal(t, tt.request.DisplayName, integration.DisplayName)
			assert.Equal(t, StatusPending, integration.Status)
			assert.True(t, integration.Enabled)
		})
	}
}

func TestIntegrationsService_GetIntegration(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create a test integration
	request := &CreateIntegrationRequest{
		IntegrationType: TypeCalendar,
		Provider:        ProviderGoogle,
		AuthMethod:      AuthOAuth2,
		DisplayName:     "Test Integration",
		Description:     "Test Description",
	}

	created, err := service.CreateIntegration(familyID, userID, request)
	require.NoError(t, err)

	tests := []struct {
		name           string
		integrationID  string
		expectedError  bool
		expectedResult bool
	}{
		{
			name:           "get existing integration",
			integrationID:  created.ID,
			expectedError:  false,
			expectedResult: true,
		},
		{
			name:           "get non-existent integration",
			integrationID:  "non_existent_id",
			expectedError:  true,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration, err := service.GetIntegration(tt.integrationID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, integration)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult {
					assert.NotNil(t, integration)
					assert.Equal(t, tt.integrationID, integration.ID)
					assert.Equal(t, familyID, integration.FamilyID)
					assert.Equal(t, userID, integration.CreatedBy)
				}
			}
		})
	}
}

func TestIntegrationsService_ListIntegrations(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create multiple test integrations
	integrations := []struct {
		integrationType IntegrationType
		provider        Provider
		displayName     string
	}{
		{TypeCalendar, ProviderGoogle, "Google Calendar"},
		{TypeCalendar, ProviderGoogle, "Work Calendar"},
	}

	createdIntegrations := make([]*Integration, 0, len(integrations))
	for _, integ := range integrations {
		request := &CreateIntegrationRequest{
			IntegrationType: integ.integrationType,
			Provider:        integ.provider,
			AuthMethod:      AuthOAuth2,
			DisplayName:     integ.displayName,
		}

		created, err := service.CreateIntegration(familyID, userID, request)
		require.NoError(t, err)
		createdIntegrations = append(createdIntegrations, created)
	}

	// Test listing integrations
	result, err := service.ListIntegrations(familyID, &ListIntegrationsQuery{})
	assert.NoError(t, err)
	assert.Len(t, result, len(integrations))

	// Verify all created integrations are in the result
	integrationIDs := make(map[string]bool)
	for _, integration := range result {
		integrationIDs[integration.ID] = true
		assert.Equal(t, familyID, integration.FamilyID)
	}

	for _, created := range createdIntegrations {
		assert.True(t, integrationIDs[created.ID], "Integration %s not found in list", created.ID)
	}
}

func TestIntegrationsService_UpdateIntegration(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create a test integration
	request := &CreateIntegrationRequest{
		IntegrationType: TypeCalendar,
		Provider:        ProviderGoogle,
		AuthMethod:      AuthOAuth2,
		DisplayName:     "Original Name",
		Description:     "Original Description",
	}

	created, err := service.CreateIntegration(familyID, userID, request)
	require.NoError(t, err)

	// Test update
	updateRequest := &UpdateIntegrationRequest{
		DisplayName: "Updated Name",
		Description: "Updated Description",
		Settings: map[string]any{
			"sync_frequency_minutes": 60,
			"sync_range_days":        14,
		},
	}

	updated, err := service.UpdateIntegration(created.ID, updateRequest)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated Name", updated.DisplayName)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.NotNil(t, updated.Settings)

	// Test update non-existent integration
	_, err = service.UpdateIntegration("non_existent", updateRequest)
	assert.Error(t, err)
}

func TestIntegrationsService_DeleteIntegration(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create a test integration
	request := &CreateIntegrationRequest{
		IntegrationType: TypeCalendar,
		Provider:        ProviderGoogle,
		AuthMethod:      AuthOAuth2,
		DisplayName:     "Test Integration",
	}

	created, err := service.CreateIntegration(familyID, userID, request)
	require.NoError(t, err)

	// Test successful deletion
	err = service.DeleteIntegration(created.ID)
	assert.NoError(t, err)

	// Verify integration is deleted
	_, err = service.GetIntegration(created.ID)
	assert.Error(t, err)

	// Test delete non-existent integration (should not return error)
	err = service.DeleteIntegration("non_existent")
	assert.NoError(t, err) // Delete is idempotent
}

func TestIntegrationsService_StoreOAuthCredentials(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create a test integration
	request := &CreateIntegrationRequest{
		IntegrationType: TypeCalendar,
		Provider:        ProviderGoogle,
		AuthMethod:      AuthOAuth2,
		DisplayName:     "Test Integration",
	}

	created, err := service.CreateIntegration(familyID, userID, request)
	require.NoError(t, err)

	// Store OAuth credentials
	accessToken := "test_access_token"
	refreshToken := "test_refresh_token"
	tokenType := "Bearer"
	scope := "https://www.googleapis.com/auth/calendar.readonly"
	expiresAt := time.Now().Add(1 * time.Hour)

	err = service.StoreOAuthCredentials(created.ID, accessToken, refreshToken, tokenType, scope, &expiresAt)
	assert.NoError(t, err)

	// Just verify the method doesn't return an error
	// The actual storage mechanism might use a different approach
}

func TestIntegrationsService_GetOAuthCredentials(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create a test integration
	request := &CreateIntegrationRequest{
		IntegrationType: TypeCalendar,
		Provider:        ProviderGoogle,
		AuthMethod:      AuthOAuth2,
		DisplayName:     "Test Integration",
	}

	created, err := service.CreateIntegration(familyID, userID, request)
	require.NoError(t, err)

	// Test getting credentials for integration without stored credentials
	_, err = service.getOAuthCredentials(created.ID)
	assert.Error(t, err) // Should error since no credentials stored

	// Test getting credentials for non-existent integration
	_, err = service.getOAuthCredentials("non_existent")
	assert.Error(t, err)
}

func TestIntegrationsService_GetRecentSyncHistory(t *testing.T) {
	db, encryptionSvc := setupIntegrationTestDB(t)
	service := NewIntegrationsService(db, encryptionSvc)
	familyID, userID := setupTestFamily(t, db)

	// Create a test integration
	request := &CreateIntegrationRequest{
		IntegrationType: TypeCalendar,
		Provider:        ProviderGoogle,
		AuthMethod:      AuthOAuth2,
		DisplayName:     "Test Integration",
	}

	created, err := service.CreateIntegration(familyID, userID, request)
	require.NoError(t, err)

	// Test getting sync history (should be empty initially)
	history, err := service.getRecentSyncHistory(created.ID, 10)
	assert.NoError(t, err)
	assert.Len(t, history, 0)

	// Test with non-existent integration
	_, err = service.getRecentSyncHistory("non_existent", 10)
	assert.NoError(t, err) // Should return empty list, not error
}

// Helper functions for pointer creation
func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int) *int {
	return &i
}

func TimePtr(t time.Time) *time.Time {
	return &t
}
