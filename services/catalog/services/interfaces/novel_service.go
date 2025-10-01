package interfaces

import (
	"context"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// NovelServiceInterface defines the contract for novel operations
type NovelServiceInterface interface {
	CreateNovel(ctx context.Context, req d.CreateNovelRequest) (*m.Novel, error)
	ListNovels(ctx context.Context, req d.ListNovelsRequest) (*d.PaginatedNovelsResponse, error)
	GetNovelByID(ctx context.Context, id string, includeTranslations, includeStats bool, language string) (*d.NovelDetailResponse, error)
	UpdateNovel(ctx context.Context, id string, req d.UpdateNovelRequest) (*d.UpdateNovelResponse, error)
	DeleteNovel(ctx context.Context, id string, deletedByUserID string) error
}
