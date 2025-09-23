package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tenant status constants
const (
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusInactive  = "inactive"
)

// Membership status constants
const (
	MembershipStatusActive    = "active"
	MembershipStatusInactive  = "inactive"
	MembershipStatusPending   = "pending"
	MembershipStatusSuspended = "suspended"
)

// TenantSettings represents tenant settings as JSONB
type TenantSettings map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (ts TenantSettings) Value() (driver.Value, error) {
	if ts == nil {
		return nil, nil
	}
	return json.Marshal(ts)
}

// Scan implements sql.Scanner interface for database retrieval
func (ts *TenantSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into TenantSettings", value)
	}

	return json.Unmarshal(bytes, ts)
}

// Tenant represents an organization or team in multi-tenant system
type Tenant struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Name        string          `json:"name" db:"name" validate:"required,max=150"`
	Slug        string          `json:"slug,omitempty" db:"slug"`
	Description *string         `json:"description,omitempty" db:"description"`
	Settings    *TenantSettings `json:"settings,omitempty" db:"settings"`
	Status      string          `json:"status" db:"status"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// Membership links users to tenants (many-to-many)
type Membership struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	UserID   uuid.UUID  `json:"user_id" db:"user_id"`
	TenantID uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RoleID   *uuid.UUID `json:"role_id,omitempty" db:"role_id"`
	Role     string     `json:"role" db:"role"`
	Status   string     `json:"status" db:"status"`
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
}
