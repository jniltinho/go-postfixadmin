package dto

import "time"

// --- Role DTOs ---

// RBACPermissionResponse is the read representation of a single permission entry.
type RBACPermissionResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// RBACRoleResponse is the read representation of a role, optionally including
// its assigned permissions when the endpoint requests full detail.
type RBACRoleResponse struct {
	ID          int                      `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	System      bool                     `json:"system"`
	Created     time.Time                `json:"created"`
	Modified    time.Time                `json:"modified"`
	Permissions []RBACPermissionResponse `json:"permissions,omitempty"`
}

// CreateRoleRequest is the body for POST /api/v1/rbac/roles.
type CreateRoleRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	PermissionIDs []int  `json:"permission_ids"`
}

// UpdateRoleRequest is the body for PUT /api/v1/rbac/roles/:id.
// Only Description and PermissionIDs are mutable; Name and System are read-only.
type UpdateRoleRequest struct {
	Description   *string `json:"description"`
	PermissionIDs []int   `json:"permission_ids"`
}

// --- Admin role assignment DTOs ---

// AssignRoleRequest is the body for POST /api/v1/rbac/admins/:username/roles.
// Domain is optional: an empty string means the role applies to all domains
// the admin manages (global scope).
type AssignRoleRequest struct {
	RoleID int    `json:"role_id"`
	Domain string `json:"domain"`
}

// RBACAdminRoleResponse is the read representation of a role assignment.
type RBACAdminRoleResponse struct {
	ID       int              `json:"id"`
	Username string           `json:"username"`
	RoleID   int              `json:"role_id"`
	Domain   string           `json:"domain"`
	Created  time.Time        `json:"created"`
	Role     RBACRoleResponse `json:"role"`
}
