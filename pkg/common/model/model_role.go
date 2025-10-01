package model

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a tenant-scoped fine-grained permission.
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key" validate:"required,max=100"`
	Description string    `json:"description,omitempty" db:"description"`
}

// Role represents a tenant-scoped role (collection of permissions)
type Role struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    *uuid.UUID `json:"tenant_id,omitempty" db:"tenant_id"`
	Name        string     `json:"name" db:"name" validate:"required,max=100"`
	Description string     `json:"description,omitempty" db:"description"`
	IsDefault   bool       `json:"is_default" db:"is_default"`
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

// GlobalPermission represents a platform-level permission that is not tenant scoped.
type GlobalPermission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key" validate:"required,max=100"`
	Description string    `json:"description,omitempty" db:"description"`
}

// GlobalRole represents a collection of global permissions.
type GlobalRole struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,max=100"`
	Description string    `json:"description,omitempty" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
}

// GlobalRolePermission links a global role to global permissions.
type GlobalRolePermission struct {
	RoleID       uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID uuid.UUID `json:"permission_id" db:"permission_id"`
}

// UserGlobalRole records which global roles are assigned to a user.
type UserGlobalRole struct {
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	RoleID     uuid.UUID `json:"role_id" db:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}
