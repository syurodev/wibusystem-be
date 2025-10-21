package dto

import "github.com/google/uuid"

// CreateNovelRequest defines the structure for creating a new novel.
type CreateNovelRequest struct {
	Title    string    `json:"title" validate:"required,min=1,max=255"`
	Summary  string    `json:"summary" validate:"required"`
	AuthorID uuid.UUID `json:"author_id" validate:"required"`
}
