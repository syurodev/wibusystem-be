package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserBio represents structured bio information stored as JSONB
type UserBio struct {
	Text     string            `json:"text,omitempty"`
	Links    []UserBioLink     `json:"links,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UserBioLink represents a link in user bio
type UserBioLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type,omitempty"` // website, social, portfolio, etc.
}

// Value implements driver.Valuer interface for database storage
func (ub UserBio) Value() (driver.Value, error) {
	return json.Marshal(ub)
}

// Scan implements sql.Scanner interface for database retrieval
func (ub *UserBio) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into UserBio", value)
	}

	return json.Unmarshal(bytes, ub)
}

// User represents a global user identify and profile info
type User struct {
	ID            uuid.UUID        `json:"id" db:"id"`
	Email         string           `json:"email" db:"email" validate:"required,email"`
	Username      string           `json:"username,omitempty" db:"username"`
	DisplayName   string           `json:"display_name,omitempty" db:"display_name"`
	AvatarURL     *string          `json:"avatar_url,omitempty" db:"avatar_url"`
	CoverImageURL *string          `json:"cover_image_url,omitempty" db:"cover_image_url"`
	Bio           *json.RawMessage `json:"bio,omitempty" db:"bio"` // Plate editor JSON data
	IsBlocked     bool             `json:"is_blocked" db:"is_blocked"`
	CreatedAt     time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" db:"updated_at"`
	LastLoginAt   *time.Time       `json:"last_login_at,omitempty" db:"last_login_at"`
}
