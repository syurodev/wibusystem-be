package model

import (
	"time"

	"github.com/google/uuid"
)

// Genre represents a content genre that can be applied to anime, manga, or novels
type Genre struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}