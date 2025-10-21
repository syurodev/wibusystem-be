package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// NovelService defines the interface for novel-related business logic.
type NovelService interface {
	CreateNovel(ctx context.Context, title, summary string, authorID uuid.UUID) (*domain.Novel, error)
	GetNovel(ctx context.Context, id uuid.UUID) (*domain.Novel, error)
	UpdateNovel(ctx context.Context, id uuid.UUID, title, summary string) (*domain.Novel, error)
	DeleteNovel(ctx context.Context, id uuid.UUID) error
	ListNovels(ctx context.Context, page, pageSize int) ([]*domain.Novel, int64, error)
}
