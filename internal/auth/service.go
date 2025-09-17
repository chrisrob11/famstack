package auth

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"famstack/internal/encryption"
)

// Service handles authentication operations
type Service struct {
	db         *sql.DB
	jwtManager *JWTManager

	// Rate limiting for password attempts
	upgradeAttempts map[string][]time.Time
	upgradeMutex    sync.RWMutex
}

// NewService creates a new authentication service using encryption service for JWT signing
func NewService(db *sql.DB, encryptionService *encryption.Service, issuer string) *Service {
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
	user, err := s.getUserByEmail(email)
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
	token, err := s.jwtManager.CreateToken(user.ID, user.FamilyID, *user.Role, 4*time.Hour)
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
		Permissions: GetPermissionList(*user.Role),
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
	user, err := s.getUserByID(claims.UserID)
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

// GetUserByToken gets user info from a valid token
func (s *Service) GetUserByToken(token string) (*User, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return s.getUserByID(claims.UserID)
}

// checkUpgradeRateLimit implements rate limiting for password upgrade attempts
func (s *Service) checkUpgradeRateLimit(userID string) bool {
	s.upgradeMutex.Lock()
	defer s.upgradeMutex.Unlock()

	now := time.Now()
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

// getUserByEmail fetches a family member by email address
func (s *Service) getUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, family_id, name, nickname, member_type, age, avatar_url, email, password_hash,
			   role, email_verified, last_login_at, display_order, is_active, created_at, updated_at
		FROM family_members
		WHERE email = ? AND email_verified = true AND password_hash IS NOT NULL
	`

	var user User
	var nickname, userEmail, passwordHash, avatarURL sql.NullString
	var age sql.NullInt64
	var role sql.NullString
	var lastLoginAt sql.NullTime

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.FamilyID, &user.Name, &nickname, &user.MemberType, &age, &avatarURL,
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
	if nickname.Valid {
		user.Nickname = &nickname.String
	}
	if age.Valid {
		ageInt := int(age.Int64)
		user.Age = &ageInt
	}
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
		roleEnum := Role(role.String)
		user.Role = &roleEnum
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// getUserByID fetches a family member by ID
func (s *Service) getUserByID(userID string) (*User, error) {
	query := `
		SELECT id, family_id, name, nickname, member_type, age, avatar_url, email, password_hash,
			   role, email_verified, last_login_at, display_order, is_active, created_at, updated_at
		FROM family_members
		WHERE id = ?
	`

	var user User
	var nickname, userEmail, passwordHash, avatarURL sql.NullString
	var age sql.NullInt64
	var role sql.NullString
	var lastLoginAt sql.NullTime

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.FamilyID, &user.Name, &nickname, &user.MemberType, &age, &avatarURL,
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
	if nickname.Valid {
		user.Nickname = &nickname.String
	}
	if age.Valid {
		ageInt := int(age.Int64)
		user.Age = &ageInt
	}
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
		roleEnum := Role(role.String)
		user.Role = &roleEnum
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

// updateLastLogin updates the family member's last login timestamp
func (s *Service) updateLastLogin(userID string) error {
	query := `UPDATE family_members SET last_login_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.db.Exec(query, userID)
	return err
}

// CreateUser creates a new family member with auth details
func (s *Service) CreateUser(req *CreateUserRequest) (*User, error) {
	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Build the name from first and last name
	name := req.FirstName
	if req.LastName != "" {
		name = req.FirstName + " " + req.LastName
	}

	// Insert family member into database with auth fields
	query := `
		INSERT INTO family_members (family_id, name, member_type, email, password_hash, role, email_verified, is_active, created_at, updated_at)
		VALUES (?, ?, 'adult', ?, ?, ?, true, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err = s.db.Exec(query, req.FamilyID, name, req.Email, hashedPassword, req.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create family member: %w", err)
	}

	// Add a small delay to ensure the transaction is committed
	time.Sleep(10 * time.Millisecond)

	// Fetch and return the created user by email (since UUID is auto-generated)
	return s.getUserByEmail(req.Email)
}
