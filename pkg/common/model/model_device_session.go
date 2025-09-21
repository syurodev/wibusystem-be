package model

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// Device records trusted or known devices used by users
type Device struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	Name         *string    `json:"name,omitempty" db:"name"`
	DeviceType   *string    `json:"device_type,omitempty" db:"device_type"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty" db:"last_seen_at"`
	RegisteredAt time.Time  `json:"registered_at" db:"registered_at"`
}

// Session tracks active login sessions or tokens
type Session struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	DeviceID     *uuid.UUID `json:"device_id,omitempty" db:"device_id"`
	TokenHash    *string    `json:"-" db:"token_hash"` // Never expose in JSON
	IPAddress    *net.IP    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty" db:"last_active_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}
