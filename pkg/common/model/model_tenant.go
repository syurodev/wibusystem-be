package model

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents an organization or team in multi-tenant system
type Tenant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required,max=150"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Membership links users to tenants (many-to-many)
type Membership struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	UserID   uuid.UUID  `json:"user_id" db:"user_id"`
	TenantID uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RoleID   *uuid.UUID `json:"role_id,omitempty" db:"role_id"`
	Status   string     `json:"status" db:"status"`
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
}
