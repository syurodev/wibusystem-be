package dto

import "github.com/google/uuid"

// GlobalRole represents a global role and its permissions returned from the API.
type GlobalRole struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Permissions []string  `json:"permissions"`
}

// TenantRole represents a tenant-scoped role definition.
type TenantRole struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsDefault   bool      `json:"is_default"`
	Permissions []string  `json:"permissions"`
}

// Permission represents a basic permission descriptor.
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description,omitempty"`
}

// TenantRolePayload captures role data used during creation/update.
type TenantRolePayload struct {
	Name           string   `json:"name" binding:"required"`
	Description    *string  `json:"description,omitempty"`
	IsDefault      bool     `json:"is_default"`
	PermissionKeys []string `json:"permission_keys"`
}

// AssignGlobalRoleRequest is used to assign a global role to a user.
type AssignGlobalRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
}

// AssignTenantRoleRequest assigns a tenant role to a user.
type AssignTenantRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
}
