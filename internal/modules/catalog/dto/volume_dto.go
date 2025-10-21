package dto

import "github.com/google/uuid"

// CreateVolumeRequest defines the structure for creating a new volume.
type CreateVolumeRequest struct {
	Title        string    `json:"title" validate:"required,min=1,max=255"`
	NovelID      uuid.UUID `json:"novel_id" validate:"required"`
	VolumeNumber int       `json:"volume_number" validate:"required,gte=1"`
}
