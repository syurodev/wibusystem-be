package novel_chapter

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ChapterHistoryRepository interface for logging chapter history
type ChapterHistoryRepository interface {
	LogUpdate(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, oldChapter, newChapter *domain.NovelChapter, changedBy uuid.UUID, requestContext map[string]any) error
	LogPublish(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error
	GetLatestVersion(ctx context.Context, chapterID uuid.UUID) (int, error)
}
