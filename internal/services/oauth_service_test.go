package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"famstack/internal/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOAuthTestDB(t *testing.T) *database.Fascade {
	dbFile := fmt.Sprintf("test_oauth_%d.db", time.Now().UnixNano())
	db, err := database.New(dbFile)
	require.NoError(t, err)

	err = db.MigrateUp()
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
		os.Remove(dbFile)
	})

	return db
}

func setupTestFamilyAndUser(t *testing.T, db *database.Fascade) (familyID, userID string) {
	familyID = fmt.Sprintf("fam_oauth_test_%d", time.Now().UnixNano())
	userID = fmt.Sprintf("user_oauth_test_%d", time.Now().UnixNano())

	// Create test family
	_, err := db.Exec(`INSERT INTO families (id, name, timezone) VALUES (?, ?, ?)`,
		familyID, "OAuth Test Family", "UTC")
	require.NoError(t, err)

	// Create test user
	_, err = db.Exec(`INSERT INTO family_members (id, family_id, first_name, last_name, member_type, is_active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, familyID, "OAuth", "User", "child", true, time.Now(), time.Now())
	require.NoError(t, err)

	return familyID, userID
}

func createTestOAuthToken(familyID, userID string) *OAuthToken {
	return &OAuthToken{
		ID:           fmt.Sprintf("token_%d", time.Now().UnixNano()),
		FamilyID:     familyID,
		UserID:       userID,
		Provider:     "google",
		AccessToken:  "test_access_token_123",
		RefreshToken: "test_refresh_token_456",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
		Scope:        "https://www.googleapis.com/auth/calendar.readonly",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func createTestOAuthState(userID string) *OAuthState {
	return &OAuthState{
		State:     fmt.Sprintf("state_%d", time.Now().UnixNano()),
		UserID:    userID,
		Provider:  "google",
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}
}

func TestOAuthService_SaveAndGetToken(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	familyID, userID := setupTestFamilyAndUser(t, db)

	// Create test token
	originalToken := createTestOAuthToken(familyID, userID)

	t.Run("save new token", func(t *testing.T) {
		err := service.SaveToken(originalToken)
		assert.NoError(t, err)
	})

	t.Run("get saved token", func(t *testing.T) {
		retrievedToken, err := service.GetToken(userID, "google")
		assert.NoError(t, err)
		assert.NotNil(t, retrievedToken)

		// Verify all fields match
		assert.Equal(t, originalToken.ID, retrievedToken.ID)
		assert.Equal(t, originalToken.FamilyID, retrievedToken.FamilyID)
		assert.Equal(t, originalToken.UserID, retrievedToken.UserID)
		assert.Equal(t, originalToken.Provider, retrievedToken.Provider)
		assert.Equal(t, originalToken.AccessToken, retrievedToken.AccessToken)
		assert.Equal(t, originalToken.RefreshToken, retrievedToken.RefreshToken)
		assert.Equal(t, originalToken.TokenType, retrievedToken.TokenType)
		assert.Equal(t, originalToken.Scope, retrievedToken.Scope)

		// Time fields should be close (allowing for small database precision differences)
		assert.WithinDuration(t, originalToken.ExpiresAt, retrievedToken.ExpiresAt, time.Second)
		assert.WithinDuration(t, originalToken.CreatedAt, retrievedToken.CreatedAt, time.Second)
		assert.WithinDuration(t, originalToken.UpdatedAt, retrievedToken.UpdatedAt, time.Second)
	})

	t.Run("get non-existent token", func(t *testing.T) {
		_, err := service.GetToken("non_existent_user", "google")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token not found")
	})

	t.Run("get token for different provider", func(t *testing.T) {
		_, err := service.GetToken(userID, "microsoft")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token not found")
	})
}

func TestOAuthService_UpdateToken(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	familyID, userID := setupTestFamilyAndUser(t, db)

	// Save initial token
	originalToken := createTestOAuthToken(familyID, userID)
	err := service.SaveToken(originalToken)
	require.NoError(t, err)

	// Update token with new values
	updatedToken := &OAuthToken{
		ID:           originalToken.ID, // Same ID for replacement
		FamilyID:     originalToken.FamilyID,
		UserID:       originalToken.UserID,
		Provider:     originalToken.Provider,
		AccessToken:  "updated_access_token_789",
		RefreshToken: "updated_refresh_token_101112",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().UTC().Add(2 * time.Hour),
		Scope:        "https://www.googleapis.com/auth/calendar",
		CreatedAt:    originalToken.CreatedAt,
		UpdatedAt:    time.Now().UTC(),
	}

	err = service.SaveToken(updatedToken)
	assert.NoError(t, err)

	// Retrieve and verify update
	retrievedToken, err := service.GetToken(userID, "google")
	assert.NoError(t, err)
	assert.Equal(t, updatedToken.AccessToken, retrievedToken.AccessToken)
	assert.Equal(t, updatedToken.RefreshToken, retrievedToken.RefreshToken)
	assert.Equal(t, updatedToken.Scope, retrievedToken.Scope)
	assert.WithinDuration(t, updatedToken.ExpiresAt, retrievedToken.ExpiresAt, time.Second)
}

func TestOAuthService_DeleteToken(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	familyID, userID := setupTestFamilyAndUser(t, db)

	// Save token
	token := createTestOAuthToken(familyID, userID)
	err := service.SaveToken(token)
	require.NoError(t, err)

	// Verify token exists
	_, err = service.GetToken(userID, "google")
	assert.NoError(t, err)

	// Delete token
	err = service.DeleteToken(userID, "google")
	assert.NoError(t, err)

	// Verify token is gone
	_, err = service.GetToken(userID, "google")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token not found")

	// Delete non-existent token should not error
	err = service.DeleteToken("non_existent_user", "google")
	assert.NoError(t, err)
}

func TestOAuthService_SaveAndGetState(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	_, userID := setupTestFamilyAndUser(t, db)

	// Create test state
	originalState := createTestOAuthState(userID)

	t.Run("save new state", func(t *testing.T) {
		err := service.SaveState(originalState)
		assert.NoError(t, err)
	})

	t.Run("get saved state", func(t *testing.T) {
		retrievedState, err := service.GetState(originalState.State)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedState)

		// Verify all fields match
		assert.Equal(t, originalState.State, retrievedState.State)
		assert.Equal(t, originalState.UserID, retrievedState.UserID)
		assert.Equal(t, originalState.Provider, retrievedState.Provider)
		assert.WithinDuration(t, originalState.ExpiresAt, retrievedState.ExpiresAt, time.Second)
		assert.WithinDuration(t, originalState.CreatedAt, retrievedState.CreatedAt, time.Second)
	})

	t.Run("get non-existent state", func(t *testing.T) {
		_, err := service.GetState("non_existent_state")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state not found")
	})
}

func TestOAuthService_StateExpiration(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	_, userID := setupTestFamilyAndUser(t, db)

	// Create expired state
	expiredState := &OAuthState{
		State:     "expired_state_123",
		UserID:    userID,
		Provider:  "google",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().UTC().Add(-2 * time.Hour),
	}

	// Save expired state
	err := service.SaveState(expiredState)
	require.NoError(t, err)

	// Try to get expired state
	_, err = service.GetState(expiredState.State)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state expired")
}

func TestOAuthService_DeleteState(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	_, userID := setupTestFamilyAndUser(t, db)

	// Save state
	state := createTestOAuthState(userID)
	err := service.SaveState(state)
	require.NoError(t, err)

	// Verify state exists
	_, err = service.GetState(state.State)
	assert.NoError(t, err)

	// Delete state
	err = service.DeleteState(state.State)
	assert.NoError(t, err)

	// Verify state is gone
	_, err = service.GetState(state.State)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state not found")

	// Delete non-existent state should not error
	err = service.DeleteState("non_existent_state")
	assert.NoError(t, err)
}

func TestOAuthService_GetUserFamilyID(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	familyID, userID := setupTestFamilyAndUser(t, db)

	t.Run("get existing user family ID", func(t *testing.T) {
		retrievedFamilyID, err := service.GetUserFamilyID(userID)
		assert.NoError(t, err)
		assert.Equal(t, familyID, retrievedFamilyID)
	})

	t.Run("get non-existent user family ID", func(t *testing.T) {
		_, err := service.GetUserFamilyID("non_existent_user")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestOAuthService_GetUsersWithTokens(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)

	// Create multiple families and users
	familyID1, userID1 := setupTestFamilyAndUser(t, db)
	familyID2, userID2 := setupTestFamilyAndUser(t, db)

	// Create tokens for different users and providers
	token1 := createTestOAuthToken(familyID1, userID1)
	token1.Provider = "google"
	token1.ExpiresAt = time.Now().UTC().Add(1 * time.Hour) // Valid token

	token2 := createTestOAuthToken(familyID2, userID2)
	token2.Provider = "google"
	token2.ExpiresAt = time.Now().UTC().Add(2 * time.Hour) // Valid token

	expiredToken := createTestOAuthToken(familyID1, userID1)
	expiredToken.ID = "expired_token_123"
	expiredToken.Provider = "microsoft"
	expiredToken.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour) // Expired token

	differentProviderToken := createTestOAuthToken(familyID2, userID2)
	differentProviderToken.ID = "diff_provider_token_456"
	differentProviderToken.Provider = "microsoft"
	differentProviderToken.ExpiresAt = time.Now().UTC().Add(1 * time.Hour) // Valid token

	// Save all tokens
	require.NoError(t, service.SaveToken(token1))
	require.NoError(t, service.SaveToken(token2))
	require.NoError(t, service.SaveToken(expiredToken))
	require.NoError(t, service.SaveToken(differentProviderToken))

	t.Run("get users with google tokens", func(t *testing.T) {
		tokens, err := service.GetUsersWithTokens("google")
		assert.NoError(t, err)
		assert.Len(t, tokens, 2) // Only non-expired Google tokens

		// Verify we got the right tokens
		tokenIDs := make(map[string]bool)
		for _, token := range tokens {
			tokenIDs[token.ID] = true
			assert.Equal(t, "google", token.Provider)
			assert.True(t, token.ExpiresAt.After(time.Now().UTC()))
		}
		assert.True(t, tokenIDs[token1.ID])
		assert.True(t, tokenIDs[token2.ID])
	})

	t.Run("get users with microsoft tokens", func(t *testing.T) {
		tokens, err := service.GetUsersWithTokens("microsoft")
		assert.NoError(t, err)
		assert.Len(t, tokens, 1) // Only non-expired Microsoft token

		assert.Equal(t, differentProviderToken.ID, tokens[0].ID)
		assert.Equal(t, "microsoft", tokens[0].Provider)
	})

	t.Run("get users with non-existent provider", func(t *testing.T) {
		tokens, err := service.GetUsersWithTokens("apple")
		assert.NoError(t, err)
		assert.Len(t, tokens, 0)
	})
}

func TestOAuthService_MultipleProvidersPerUser(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	familyID, userID := setupTestFamilyAndUser(t, db)

	// Create tokens for same user but different providers
	googleToken := createTestOAuthToken(familyID, userID)
	googleToken.Provider = "google"
	googleToken.AccessToken = "google_access_token"

	microsoftToken := createTestOAuthToken(familyID, userID)
	microsoftToken.ID = "microsoft_token_456"
	microsoftToken.Provider = "microsoft"
	microsoftToken.AccessToken = "microsoft_access_token"

	// Save both tokens
	require.NoError(t, service.SaveToken(googleToken))
	require.NoError(t, service.SaveToken(microsoftToken))

	// Retrieve each token independently
	retrievedGoogle, err := service.GetToken(userID, "google")
	assert.NoError(t, err)
	assert.Equal(t, "google_access_token", retrievedGoogle.AccessToken)

	retrievedMicrosoft, err := service.GetToken(userID, "microsoft")
	assert.NoError(t, err)
	assert.Equal(t, "microsoft_access_token", retrievedMicrosoft.AccessToken)

	// Delete one provider should not affect the other
	err = service.DeleteToken(userID, "google")
	assert.NoError(t, err)

	_, err = service.GetToken(userID, "google")
	assert.Error(t, err) // Google token should be gone

	_, err = service.GetToken(userID, "microsoft")
	assert.NoError(t, err) // Microsoft token should still exist
}

func TestOAuthService_StateReplacement(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	_, userID := setupTestFamilyAndUser(t, db)

	// Create initial state
	state1 := &OAuthState{
		State:     "same_state_key",
		UserID:    userID,
		Provider:  "google",
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	err := service.SaveState(state1)
	require.NoError(t, err)

	// Create replacement state with same key but different data
	state2 := &OAuthState{
		State:     "same_state_key", // Same state key
		UserID:    userID,
		Provider:  "microsoft", // Different provider
		ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}

	err = service.SaveState(state2)
	require.NoError(t, err)

	// Retrieve state should get the latest one
	retrievedState, err := service.GetState("same_state_key")
	assert.NoError(t, err)
	assert.Equal(t, "microsoft", retrievedState.Provider) // Should be the updated provider
	assert.WithinDuration(t, state2.ExpiresAt, retrievedState.ExpiresAt, time.Second)
}

func TestOAuthService_DatabaseOperations(t *testing.T) {
	db := setupOAuthTestDB(t)
	service := NewOAuthService(db)
	familyID, userID := setupTestFamilyAndUser(t, db)

	t.Run("token with valid family and user IDs should succeed", func(t *testing.T) {
		validToken := &OAuthToken{
			ID:           "valid_token_123",
			FamilyID:     familyID,
			UserID:       userID,
			Provider:     "google",
			AccessToken:  "test_token",
			RefreshToken: "test_refresh",
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
			Scope:        "test_scope",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := service.SaveToken(validToken)
		assert.NoError(t, err)

		// Verify we can retrieve it
		retrievedToken, err := service.GetToken(userID, "google")
		assert.NoError(t, err)
		assert.Equal(t, validToken.ID, retrievedToken.ID)
	})

	t.Run("state with valid user ID should succeed", func(t *testing.T) {
		validState := &OAuthState{
			State:     "valid_state_123",
			UserID:    userID,
			Provider:  "google",
			ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
			CreatedAt: time.Now().UTC(),
		}

		err := service.SaveState(validState)
		assert.NoError(t, err)

		// Verify we can retrieve it
		retrievedState, err := service.GetState(validState.State)
		assert.NoError(t, err)
		assert.Equal(t, validState.State, retrievedState.State)
	})

	t.Run("empty token fields should be handled gracefully", func(t *testing.T) {
		tokenWithEmptyRefresh := &OAuthToken{
			ID:           "empty_refresh_token_123",
			FamilyID:     familyID,
			UserID:       userID,
			Provider:     "microsoft",
			AccessToken:  "test_access_token",
			RefreshToken: "", // Empty refresh token
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().UTC().Add(1 * time.Hour),
			Scope:        "test_scope",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := service.SaveToken(tokenWithEmptyRefresh)
		assert.NoError(t, err)

		retrievedToken, err := service.GetToken(userID, "microsoft")
		assert.NoError(t, err)
		assert.Equal(t, "", retrievedToken.RefreshToken)
	})
}
