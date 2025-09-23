package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey []byte
	issuer    string
}

// JWTClaims represents the claims in our JWT tokens
type JWTClaims struct {
	UserID       string `json:"user_id"`
	FamilyID     string `json:"family_id"`
	Role         Role   `json:"role"`
	OriginalRole Role   `json:"original_role"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager with a secret key
func NewJWTManager(secretKey []byte, issuer string) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		issuer:    issuer,
	}
}

// GenerateSecretKey generates a new random secret key
func GenerateSecretKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}
	return key, nil
}

// GenerateSecretKeyHex generates a new random secret key as a hex string
func GenerateSecretKeyHex() (string, error) {
	key, err := GenerateSecretKey()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// CreateToken creates a new JWT token for a user session
func (j *JWTManager) CreateToken(userID, familyID string, role Role, duration time.Duration) (string, error) {
	now := time.Now().UTC()

	claims := &JWTClaims{
		UserID:       userID,
		FamilyID:     familyID,
		Role:         role,
		OriginalRole: role, // Same as role initially
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			Audience:  []string{familyID},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// CreateDowngradedToken creates a new JWT token with downgraded permissions
func (j *JWTManager) CreateDowngradedToken(originalClaims *JWTClaims) (string, error) {
	if originalClaims.Role == RoleShared {
		return "", fmt.Errorf("cannot downgrade: already in shared mode")
	}

	now := time.Now().UTC()

	claims := &JWTClaims{
		UserID:       originalClaims.UserID,
		FamilyID:     originalClaims.FamilyID,
		Role:         RoleShared,                  // Downgrade to shared
		OriginalRole: originalClaims.OriginalRole, // Keep original role for upgrade
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   originalClaims.UserID,
			Audience:  []string{originalClaims.FamilyID},
			ExpiresAt: originalClaims.ExpiresAt, // Keep same expiration
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// CreateUpgradedToken creates a new JWT token with restored permissions
func (j *JWTManager) CreateUpgradedToken(sharedClaims *JWTClaims) (string, error) {
	if sharedClaims.Role != RoleShared {
		return "", fmt.Errorf("cannot upgrade: not in shared mode")
	}

	now := time.Now().UTC()

	claims := &JWTClaims{
		UserID:       sharedClaims.UserID,
		FamilyID:     sharedClaims.FamilyID,
		Role:         sharedClaims.OriginalRole, // Restore original role
		OriginalRole: sharedClaims.OriginalRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   sharedClaims.UserID,
			Audience:  []string{sharedClaims.FamilyID},
			ExpiresAt: sharedClaims.ExpiresAt, // Keep same expiration
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is expired
	if time.Now().UTC().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token has expired")
	}

	// Check if token is not yet valid
	if time.Now().UTC().Before(claims.NotBefore.Time) {
		return nil, fmt.Errorf("token not yet valid")
	}

	return claims, nil
}

// RefreshToken creates a new token with extended expiration
func (j *JWTManager) RefreshToken(claims *JWTClaims, duration time.Duration) (string, error) {
	now := time.Now().UTC()

	newClaims := &JWTClaims{
		UserID:       claims.UserID,
		FamilyID:     claims.FamilyID,
		Role:         claims.Role,
		OriginalRole: claims.OriginalRole,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   claims.UserID,
			Audience:  []string{claims.FamilyID},
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(j.secretKey)
}

// GetTokenExpiration returns the expiration time of a token
func (j *JWTManager) GetTokenExpiration(tokenString string) (*time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	expiration := claims.ExpiresAt.Time
	return &expiration, nil
}

// IsTokenExpired checks if a token is expired without full validation
func (j *JWTManager) IsTokenExpired(tokenString string) (bool, error) {
	// Parse without verification to check expiration
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})

	if err != nil {
		return true, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return time.Now().UTC().After(claims.ExpiresAt.Time), nil
	}

	return true, nil // Assume expired if we can't parse
}
