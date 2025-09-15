package auth

import (
	"testing"
	"time"
)

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Test hashing
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	// Test verification with correct password
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}

	if !valid {
		t.Error("Password should be valid")
	}

	// Test verification with wrong password
	valid, err = VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("Failed to verify wrong password: %v", err)
	}

	if valid {
		t.Error("Wrong password should not be valid")
	}
}

func TestJWTTokenCreation(t *testing.T) {
	secretKey, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("Failed to generate secret key: %v", err)
	}

	jwtManager := NewJWTManager(secretKey, "famstack-test")

	// Test token creation
	userID := "test-user"
	familyID := "test-family"
	role := RoleUser
	duration := 1 * time.Hour

	token, err := jwtManager.CreateToken(userID, familyID, role, duration)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Test token validation
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.FamilyID != familyID {
		t.Errorf("Expected FamilyID %s, got %s", familyID, claims.FamilyID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}
}

func TestPermissionSystem(t *testing.T) {
	// Test shared role permissions
	if !HasPermission(RoleShared, EntityTask, ActionRead, ScopeAny) {
		t.Error("Shared role should be able to read tasks")
	}

	if HasPermission(RoleShared, EntityTask, ActionDelete, ScopeAny) {
		t.Error("Shared role should not be able to delete tasks")
	}

	// Test user role permissions
	if !HasPermission(RoleUser, EntityTask, ActionCreate, ScopeAny) {
		t.Error("User role should be able to create tasks")
	}

	if !HasPermission(RoleUser, EntityTask, ActionDelete, ScopeOwn) {
		t.Error("User role should be able to delete own tasks")
	}

	if HasPermission(RoleUser, EntityTask, ActionDelete, ScopeAny) {
		t.Error("User role should not be able to delete any tasks")
	}

	// Test admin role permissions
	if !HasPermission(RoleAdmin, EntityTask, ActionDelete, ScopeAny) {
		t.Error("Admin role should be able to delete any tasks")
	}
}

func TestRoleDowngradeUpgrade(t *testing.T) {
	secretKey, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("Failed to generate secret key: %v", err)
	}

	jwtManager := NewJWTManager(secretKey, "famstack-test")

	// Create initial user token
	userID := "test-user"
	familyID := "test-family"
	role := RoleUser
	duration := 1 * time.Hour

	userToken, err := jwtManager.CreateToken(userID, familyID, role, duration)
	if err != nil {
		t.Fatalf("Failed to create user token: %v", err)
	}

	// Validate initial token
	userClaims, err := jwtManager.ValidateToken(userToken)
	if err != nil {
		t.Fatalf("Failed to validate user token: %v", err)
	}

	// Test downgrade to shared
	sharedToken, err := jwtManager.CreateDowngradedToken(userClaims)
	if err != nil {
		t.Fatalf("Failed to create downgraded token: %v", err)
	}

	sharedClaims, err := jwtManager.ValidateToken(sharedToken)
	if err != nil {
		t.Fatalf("Failed to validate shared token: %v", err)
	}

	if sharedClaims.Role != RoleShared {
		t.Errorf("Expected Role %s, got %s", RoleShared, sharedClaims.Role)
	}

	if sharedClaims.OriginalRole != RoleUser {
		t.Errorf("Expected OriginalRole %s, got %s", RoleUser, sharedClaims.OriginalRole)
	}

	// Test upgrade back to original
	upgradedToken, err := jwtManager.CreateUpgradedToken(sharedClaims)
	if err != nil {
		t.Fatalf("Failed to create upgraded token: %v", err)
	}

	upgradedClaims, err := jwtManager.ValidateToken(upgradedToken)
	if err != nil {
		t.Fatalf("Failed to validate upgraded token: %v", err)
	}

	if upgradedClaims.Role != RoleUser {
		t.Errorf("Expected Role %s, got %s", RoleUser, upgradedClaims.Role)
	}

	if upgradedClaims.OriginalRole != RoleUser {
		t.Errorf("Expected OriginalRole %s, got %s", RoleUser, upgradedClaims.OriginalRole)
	}
}