package model

import (
	"time"

	"github.com/google/uuid"
)

// Creator represents a content creator (author, artist, studio, voice actor, etc.)
type Creator struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreatorWithRole represents a creator with their role in a specific content
type CreatorWithRole struct {
	Creator
	Role string `json:"role" db:"role"` // AUTHOR, ILLUSTRATOR, ARTIST, STUDIO, VOICE_ACTOR
}