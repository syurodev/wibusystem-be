package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a global user identify and profile info
type User struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	Email         string           `json:"email" db:"email" validate:"required,email"`
	Username      string           `json:"username,omitempty" db:"username"`
	DisplayName   string           `json:"display_name,omitempty" db:"display_name"`
	AvatarURL     *string          `json:"avatar_url" db:"avatar_url"`
	CoverImageURL *string          `json:"cover_image_url,omitempty" db:"cover_image_url"`
	Bio           *json.RawMessage `json:"bio,omitempty" db:"bio"` // Plate editor JSON data
	IsBlocked     bool             `json:"is_blocked" db:"is_blocked"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	LastLoginAt   *time.Time       `json:"last_login_at,omitempty" db:"last_login_at"`
}
