package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// CreatorService defines the interface for creator-related business logic.
type CreatorService interface {
	CreateCreator(ctx context.Context, name, bio string) (*domain.Creator, error)
	GetCreator(ctx context.Context, id uuid.UUID) (*domain.Creator, error)
	ListCreators(ctx context.Context, page, pageSize int) ([]*domain.Creator, int64, error)
}
