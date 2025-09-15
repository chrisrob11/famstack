package auth

// AuthorizationService handles permission checking
type AuthorizationService struct {
	session *Session
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(session *Session) *AuthorizationService {
	return &AuthorizationService{
		session: session,
	}
}

// HasPermission checks if the current session has permission for an entity/action
func (a *AuthorizationService) HasPermission(entity Entity, action Action, resourceOwnerID *string) bool {
	// Check different scopes in order of preference

	// 1. Check for "any" scope permission
	if a.hasExactPermission(entity, action, ScopeAny) {
		return true
	}

	// 2. Check for "own" scope permission
	if a.hasExactPermission(entity, action, ScopeOwn) {
		return a.isOwner(resourceOwnerID)
	}

	// 3. Default to no access
	return false
}

// hasExactPermission checks for a specific permission with exact scope
func (a *AuthorizationService) hasExactPermission(entity Entity, action Action, scope PermissionScope) bool {
	return HasPermission(a.session.Role, entity, action, scope)
}

// isOwner checks if the current user owns the resource
func (a *AuthorizationService) isOwner(resourceOwnerID *string) bool {
	// Shared sessions can't own anything
	if a.session.Role == RoleShared {
		return false
	}

	if resourceOwnerID == nil {
		return false // No owner specified
	}

	return a.session.UserID == *resourceOwnerID
}

// CanUpgradeToAccess checks if upgrading to original role would grant access
func (a *AuthorizationService) CanUpgradeToAccess(entity Entity, action Action, resourceOwnerID *string) bool {
	// Only shared sessions can upgrade
	if a.session.Role != RoleShared {
		return false
	}

	// Check if the original role would have this permission
	originalPermissions := RolePermissions[a.session.OriginalRole]

	// Try "any" scope first
	anyPerm := MakePermission(entity, action, ScopeAny)
	if originalPermissions[anyPerm] {
		return true
	}

	// Try "own" scope
	ownPerm := MakePermission(entity, action, ScopeOwn)
	if originalPermissions[ownPerm] {
		return true // Would work if they own the resource after auth
	}

	return false
}

// WouldHavePermissionAfterUpgrade checks what permission would be available after upgrade
func (a *AuthorizationService) WouldHavePermissionAfterUpgrade(entity Entity, action Action, resourceOwnerID *string) bool {
	if a.session.Role != RoleShared {
		return false
	}

	// Temporarily check as if we had the original role
	tempAuth := &AuthorizationService{
		session: &Session{
			UserID:       a.session.UserID,
			FamilyID:     a.session.FamilyID,
			Role:         a.session.OriginalRole,
			OriginalRole: a.session.OriginalRole,
		},
	}

	return tempAuth.HasPermission(entity, action, resourceOwnerID)
}

// GetCurrentPermissions returns all permissions for the current session
func (a *AuthorizationService) GetCurrentPermissions() []string {
	return GetPermissionList(a.session.Role)
}

// GetOriginalPermissions returns what permissions would be available after upgrade
func (a *AuthorizationService) GetOriginalPermissions() []string {
	if a.session.Role == RoleShared {
		return GetPermissionList(a.session.OriginalRole)
	}
	return a.GetCurrentPermissions()
}

// CanUpdateTaskField checks if the current role can update a specific task field
func (a *AuthorizationService) CanUpdateTaskField(field string) bool {
	return CanUpdateTaskField(a.session.Role, field)
}

// GetAllowedTaskFields returns all task fields the current role can update
func (a *AuthorizationService) GetAllowedTaskFields() []string {
	return GetAllowedTaskFields(a.session.Role)
}

// FilterTaskUpdateData filters task update data to only include allowed fields
func (a *AuthorizationService) FilterTaskUpdateData(updateData map[string]interface{}) (map[string]interface{}, []string) {
	allowedFields := a.GetAllowedTaskFields()
	filtered := make(map[string]interface{})
	var denied []string

	for field, value := range updateData {
		allowed := false
		for _, allowedField := range allowedFields {
			if field == allowedField {
				allowed = true
				break
			}
		}

		if allowed {
			filtered[field] = value
		} else {
			denied = append(denied, field)
		}
	}

	return filtered, denied
}
