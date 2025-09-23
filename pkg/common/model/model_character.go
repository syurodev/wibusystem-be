package model

import (
	"time"

	"github.com/google/uuid"
)

// Character represents a fictional character that can appear in multiple content (anime/manga/novel)
type Character struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	ImageURL    *string   `json:"image_url,omitempty" db:"image_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}