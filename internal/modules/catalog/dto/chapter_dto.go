package dto

import "github.com/google/uuid"

// CreateChapterRequest defines the structure for creating a new chapter.
type CreateChapterRequest struct {
	Title         string    `json:"title" validate:"required,min=1,max=255"`
	Content       string    `json:"content" validate:"required"`
	VolumeID      uuid.UUID `json:"volume_id" validate:"required"`
	ChapterNumber int       `json:"chapter_number" validate:"required,gte=1"`
}
