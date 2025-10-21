package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// ChapterService defines the interface for chapter-related business logic.
type ChapterService interface {
	CreateChapter(ctx context.Context, volumeID uuid.UUID, title, content string, chapterNumber int) (*domain.Chapter, error)
	GetChapter(ctx context.Context, id uuid.UUID) (*domain.Chapter, error)
	ListChaptersByVolume(ctx context.Context, volumeID uuid.UUID, page, pageSize int) ([]*domain.Chapter, int64, error)
}
