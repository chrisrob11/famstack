package auth

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"famstack/internal/database"
	"famstack/internal/encryption"
	"famstack/internal/models"
)

// Service handles authentication operations
type Service struct {
	db         *database.Fascade
	jwtManager *JWTManager

	// Rate limiting for password attempts
	upgradeAttempts map[string][]time.Time
	upgradeMutex    sync.RWMutex
}

// NewService creates a new authentication service using encryption service for JWT signing
func NewService(db *database.Fascade, encryptionService *encryption.Service, issuer string) *Service {
	// Get JWT signing key from encryption service
	jwtKey, err := encryptionService.GetJWTSigningKey()
	if err != nil {
		// This is a critical error - we can't operate without JWT signing capability
		panic(fmt.Sprintf("Failed to get JWT signing key: %v", err))
	}

	return &Service{
		db:              db,
		jwtManager:      NewJWTManager(jwtKey, issuer),
		upgradeAttempts: make(map[string][]time.Time),
	}
}

// Login authenticates a user with email and password
func (s *Service) Login(email, password string) (*AuthResponse, error) {
	// Get user by email
	user, err := s.getFamilyMemberByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user has auth info
	if user.PasswordHash == nil || user.Role == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	valid, err := VerifyPassword(password, *user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login time
	if updateErr := s.updateLastLogin(user.ID); updateErr != nil {
		// Log error but don't fail authentication
		fmt.Printf("Failed to update last login for user %s: %v\n", user.ID, updateErr)
	}

	// Create JWT token (4 hours expiration for full sessions)

	token, err := s.jwtManager.CreateToken(user.ID, user.FamilyID, Role(*user.Role), 4*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	// Create session from token

	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created token: %w", err)
	}
	session := SessionFromJWTClaims(claims)

	return &AuthResponse{
		User:        user,
		Session:     session,
		Token:       token,
		Permissions: GetPermissionList(Role(*user.Role)),
	}, nil
}

// DowngradeToShared downgrades a user session to shared mode
func (s *Service) DowngradeToShared(token string) (*TokenResponse, error) {
	// Validate current token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Create downgraded token
	sharedToken, err := s.jwtManager.CreateDowngradedToken(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to downgrade: %w", err)
	}

	// Create session from new token
	sharedClaims, err := s.jwtManager.ValidateToken(sharedToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse downgraded token: %w", err)
	}
	session := SessionFromJWTClaims(sharedClaims)

	return &TokenResponse{
		Token:       sharedToken,
		Session:     session,
		Permissions: GetPermissionList(RoleShared),
	}, nil
}

// UpgradeWithPassword upgrades a shared session back to original permissions
func (s *Service) UpgradeWithPassword(token, password string) (*TokenResponse, error) {
	// Validate current token
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims.Role != RoleShared {
		return nil, fmt.Errorf("not in shared mode")
	}

	// Check rate limiting
	if !s.checkUpgradeRateLimit(claims.UserID) {
		return nil, fmt.Errorf("too many upgrade attempts, please try again later")
	}

	// Get user and verify password
	user, err := s.getFamilyMemberByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user has auth info
	if user.PasswordHash == nil {
		return nil, fmt.Errorf("user cannot authenticate")
	}

	valid, err := VerifyPassword(password, *user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if !valid {
		return nil, fmt.Errorf("invalid password")
	}

	// Create upgraded token
	originalToken, err := s.jwtManager.CreateUpgradedToken(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade: %w", err)
	}

	// Create session from new token
	originalClaims, err := s.jwtManager.ValidateToken(originalToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upgraded token: %w", err)
	}
	session := SessionFromJWTClaims(originalClaims)

	return &TokenResponse{
		Token:       originalToken,
		Session:     session,
		Permissions: GetPermissionList(session.Role),
	}, nil
}

// ValidateToken validates a JWT token and returns session info
func (s *Service) ValidateToken(token string) (*Session, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return SessionFromJWTClaims(claims), nil
}

// RefreshToken creates a new token with extended expiration
func (s *Service) RefreshToken(token string) (*TokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Create new token with 4 hours expiration
	newToken, err := s.jwtManager.RefreshToken(claims, 4*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Create session from new token
	newClaims, tokenErr := s.jwtManager.ValidateToken(newToken)
	if tokenErr != nil {
		return nil, fmt.Errorf("failed to validate token: %v", tokenErr)
	}
	session := SessionFromJWTClaims(newClaims)

	return &TokenResponse{
		Token:       newToken,
		Session:     session,
		Permissions: GetPermissionList(session.Role),
	}, nil
}

// GetFamilyMemberByToken gets user info from a valid token
func (s *Service) GetFamilyMemberByToken(token string) (*models.FamilyMember, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return s.getFamilyMemberByID(claims.UserID)
}

// checkUpgradeRateLimit implements rate limiting for password upgrade attempts
func (s *Service) checkUpgradeRateLimit(userID string) bool {
	s.upgradeMutex.Lock()
	defer s.upgradeMutex.Unlock()

	now := time.Now().UTC()
	attempts := s.upgradeAttempts[userID]

	// Remove attempts older than 15 minutes
	var recentAttempts []time.Time
	for _, attempt := range attempts {
		if now.Sub(attempt) < 15*time.Minute {
			recentAttempts = append(recentAttempts, attempt)
		}
	}

	// Allow max 5 attempts per 15 minutes
	if len(recentAttempts) >= 5 {
		return false
	}

	// Add current attempt
	recentAttempts = append(recentAttempts, now)
	s.upgradeAttempts[userID] = recentAttempts

	return true
}

// getFamilyMemberByEmail fetches a family member by email address
func (s *Service) getFamilyMemberByEmail(email string) (*models.FamilyMember, error) {
	query := `
		SELECT id, family_id, first_name, last_name, member_type, avatar_url, email, password_hash,
			   role, email_verified, last_login_at, display_order, is_active, created_at, updated_at
		FROM family_members
		WHERE email = ? AND email_verified = true AND password_hash IS NOT NULL
	`

	var user models.FamilyMember
	var userEmail, passwordHash, avatarURL sql.NullString
	var role sql.NullString
	var lastLoginAt sql.NullTime

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.FamilyID, &user.FirstName, &user.LastName, &user.MemberType, &avatarURL,
		&userEmail, &passwordHash, &role, &user.EmailVerified,
		&lastLoginAt, &user.DisplayOrder, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Handle nullable fields
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if userEmail.Valid {
		user.Email = &userEmail.String
	}
	if passwordHash.Valid {
		user.PasswordHash = &passwordHash.String
	}
	if role.Valid {
		user.Role = &role.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// getFamilyMemberByID fetches a family member by ID
func (s *Service) getFamilyMemberByID(userID string) (*models.FamilyMember, error) {
	query := `
		SELECT id, family_id, first_name, last_name, member_type, avatar_url, email, password_hash,
			   role, email_verified, last_login_at, display_order, is_active, created_at, updated_at
		FROM family_members
		WHERE id = ?
	`

	var user models.FamilyMember
	var userEmail, passwordHash, avatarURL sql.NullString
	var role sql.NullString
	var lastLoginAt sql.NullTime

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.FamilyID, &user.FirstName, &user.LastName, &user.MemberType, &avatarURL,
		&userEmail, &passwordHash, &role, &user.EmailVerified,
		&lastLoginAt, &user.DisplayOrder, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Handle nullable fields
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if userEmail.Valid {
		user.Email = &userEmail.String
	}
	if passwordHash.Valid {
		user.PasswordHash = &passwordHash.String
	}
	if role.Valid {
		user.Role = &role.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// GetFamilyMemberByID is a public wrapper for getFamilyMemberByID
func (s *Service) GetFamilyMemberByID(userID string) (*models.FamilyMember, error) {
	return s.getFamilyMemberByID(userID)
}

// updateLastLogin updates the family member's last login timestamp
func (s *Service) updateLastLogin(userID string) error {
	query := `UPDATE family_members SET last_login_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, userID)
	return err
}

// CreateFamilyMember creates a new family member with auth details
func (s *Service) CreateFamilyMember(req *CreateUserRequest) (*models.FamilyMember, error) {
	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert family member into database with auth fields
	query := `
		INSERT INTO family_members (family_id, first_name, last_name, member_type, email, password_hash, role, email_verified, is_active, created_at, updated_at)
		VALUES (?, ?, ?, 'adult', ?, ?, ?, true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err = s.db.Exec(query, req.FamilyID, req.FirstName, req.LastName, req.Email, hashedPassword, req.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create family member: %w", err)
	}

	// For SQLite, we can't get the UUID directly, so we need to query by email
	// This is safe because email should be unique per family
	return s.getFamilyMemberByEmail(req.Email)
}
