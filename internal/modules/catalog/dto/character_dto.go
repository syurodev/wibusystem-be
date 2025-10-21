package dto

import "github.com/google/uuid"

// CreateCharacterRequest defines the structure for creating a new character.
type CreateCharacterRequest struct {
	Name        string    `json:"name" validate:"required,min=1,max=150"`
	Description string    `json:"description"`
	NovelID     uuid.UUID `json:"novel_id" validate:"required"`
}
