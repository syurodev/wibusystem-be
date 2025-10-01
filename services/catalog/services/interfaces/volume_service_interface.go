package interfaces

import (
	"context"

	d "wibusystem/pkg/common/dto"
)

// VolumeServiceInterface defines business logic operations for volume management
// This service layer sits between HTTP handlers and the repository layer,
// implementing business rules and validation logic for novel volumes.
type VolumeServiceInterface interface {
	// CreateVolume creates a new volume for a specific novel
	// Validates novel existence and volume number uniqueness before creation
	// Returns the created volume response or error
	CreateVolume(ctx context.Context, novelID string, req d.CreateVolumeRequest) (*d.CreateVolumeResponse, error)

	// GetVolumeByID retrieves detailed information about a specific volume
	// Returns error if volume is not found or has been deleted
	GetVolumeByID(ctx context.Context, id string) (*d.VolumeResponse, error)

	// ListVolumesByNovelID retrieves all volumes for a specific novel with pagination
	// Returns paginated list of volumes ordered by volume number
	ListVolumesByNovelID(ctx context.Context, novelID string, req d.ListVolumesRequest) (*d.PaginatedVolumesResponse, error)

	// UpdateVolume updates an existing volume
	// Only updates non-nil fields from the request
	// Returns updated volume response or error
	UpdateVolume(ctx context.Context, id string, req d.UpdateVolumeRequest) (*d.UpdateVolumeResponse, error)

	// DeleteVolume soft-deletes a volume
	// Validates that no users have purchased content from this volume before deletion
	// Returns error if volume has purchases or is not found
	DeleteVolume(ctx context.Context, id string) error
}