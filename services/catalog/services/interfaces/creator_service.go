package interfaces

import (
	"context"

	"github.com/google/uuid"
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// CreatorServiceInterface defines the contract for creator operations
type CreatorServiceInterface interface {
	// CreateCreator creates a new creator with validation
	CreateCreator(ctx context.Context, req d.CreateCreatorRequest) (*m.Creator, error)

	// GetCreatorByID retrieves a creator by ID
	GetCreatorByID(ctx context.Context, creatorID uuid.UUID) (*m.Creator, error)

	// ListCreators retrieves paginated list of creators with optional search
	ListCreators(ctx context.Context, req d.ListCreatorsRequest) ([]*m.Creator, int64, error)

	// UpdateCreator updates creator information
	UpdateCreator(ctx context.Context, creatorID uuid.UUID, req d.UpdateCreatorRequest) (*m.Creator, error)

	// DeleteCreator removes a creator
	DeleteCreator(ctx context.Context, creatorID uuid.UUID) error

	// ValidateCreatorName validates creator name before creation/update
	ValidateCreatorName(name string) error

	// CheckCreatorExists checks if a creator name already exists
	CheckCreatorExists(ctx context.Context, name string) (bool, error)
}