package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a global user identify and profile info
type User struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Email       string     `json:"email" db:"email" validate:"required,email"`
	Username    string     `json:"username,omitempty" db:"username"`
	DisplayName string     `json:"display_name,omitempty" db:"display_name"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}
