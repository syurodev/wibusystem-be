package model

import "github.com/google/uuid"

// Permission represents a global fine-grained permission
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key" validate:"required,max=100"`
	Description string    `json:"description,omitempty" db:"description"`
}

// Role represents a tenant-scoped role (collection of permissions)
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name" validate:"required,max=100"`
	Description string    `json:"description,omitempty" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
}

// RolePermission represents the many-to-many mapping between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
}

// RoleAssignment links user memberships to roles
type RoleAssignment struct {
	MembershipID uuid.UUID `json:"membership_id" db:"membership_id"`
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
}
