package auth

import "fmt"

// PermissionSet represents a set of permissions
type PermissionSet map[Permission]bool

// MakePermission creates a permission string from entity, action, and scope
func MakePermission(entity Entity, action Action, scope PermissionScope) Permission {
	return Permission(fmt.Sprintf("%s:%s:%s", entity, action, scope))
}

// RolePermissions defines the permissions for each role
var RolePermissions = map[Role]PermissionSet{
	RoleShared: {
		// Tasks - can read and update status only
		MakePermission(EntityTask, ActionRead, ScopeAny):   true,
		MakePermission(EntityTask, ActionUpdate, ScopeAny): true, // Limited to status updates in handler

		// Calendar - read only
		MakePermission(EntityCalendar, ActionRead, ScopeAny): true,

		// No access to other entities
	},

	RoleUser: {
		// Tasks - full management for own items, read/create for any
		MakePermission(EntityTask, ActionRead, ScopeAny):   true,
		MakePermission(EntityTask, ActionCreate, ScopeAny): true,
		MakePermission(EntityTask, ActionUpdate, ScopeAny): true, // Can update any task
		MakePermission(EntityTask, ActionDelete, ScopeOwn): true, // Can only delete own tasks

		// Calendar - full management for own items, read/create for any
		MakePermission(EntityCalendar, ActionRead, ScopeAny):   true,
		MakePermission(EntityCalendar, ActionCreate, ScopeAny): true,
		MakePermission(EntityCalendar, ActionUpdate, ScopeOwn): true,
		MakePermission(EntityCalendar, ActionDelete, ScopeOwn): true,

		// Schedules - can create and manage own
		MakePermission(EntitySchedule, ActionRead, ScopeAny):   true,
		MakePermission(EntitySchedule, ActionCreate, ScopeAny): true,
		MakePermission(EntitySchedule, ActionUpdate, ScopeOwn): true,
		MakePermission(EntitySchedule, ActionDelete, ScopeOwn): true,

		// Family - can view and update own family
		MakePermission(EntityFamily, ActionRead, ScopeAny):   true,
		MakePermission(EntityFamily, ActionUpdate, ScopeOwn): true,

		// Users - can view family members
		MakePermission(EntityUser, ActionRead, ScopeAny): true,

		// No access to settings
	},

	RoleAdmin: {
		// Tasks - full access to all
		MakePermission(EntityTask, ActionRead, ScopeAny):   true,
		MakePermission(EntityTask, ActionCreate, ScopeAny): true,
		MakePermission(EntityTask, ActionUpdate, ScopeAny): true,
		MakePermission(EntityTask, ActionDelete, ScopeAny): true,

		// Calendar - full access to all
		MakePermission(EntityCalendar, ActionRead, ScopeAny):   true,
		MakePermission(EntityCalendar, ActionCreate, ScopeAny): true,
		MakePermission(EntityCalendar, ActionUpdate, ScopeAny): true,
		MakePermission(EntityCalendar, ActionDelete, ScopeAny): true,

		// Schedules - full access to all
		MakePermission(EntitySchedule, ActionRead, ScopeAny):   true,
		MakePermission(EntitySchedule, ActionCreate, ScopeAny): true,
		MakePermission(EntitySchedule, ActionUpdate, ScopeAny): true,
		MakePermission(EntitySchedule, ActionDelete, ScopeAny): true,

		// Family - full management
		MakePermission(EntityFamily, ActionRead, ScopeAny):   true,
		MakePermission(EntityFamily, ActionUpdate, ScopeAny): true,

		// Users - full user management
		MakePermission(EntityUser, ActionRead, ScopeAny):   true,
		MakePermission(EntityUser, ActionCreate, ScopeAny): true,
		MakePermission(EntityUser, ActionUpdate, ScopeAny): true,
		MakePermission(EntityUser, ActionDelete, ScopeAny): true,

		// Settings - full access
		MakePermission(EntitySetting, ActionRead, ScopeAny):   true,
		MakePermission(EntitySetting, ActionUpdate, ScopeAny): true,
	},
}

// GetPermissionList returns a list of permission strings for a role
func GetPermissionList(role Role) []string {
	permissions := RolePermissions[role]
	var list []string

	for permission := range permissions {
		if permissions[permission] {
			list = append(list, string(permission))
		}
	}

	return list
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, entity Entity, action Action, scope PermissionScope) bool {
	permissions := RolePermissions[role]
	permission := MakePermission(entity, action, scope)
	return permissions[permission]
}
