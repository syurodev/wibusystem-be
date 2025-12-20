package novel_chapter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ChapterService interface định nghĩa business logic cho chapters
type ChapterService interface {
	// CreateChapter creates a new chapter
	CreateChapter(
		ctx context.Context,
		novelID, volumeID uuid.UUID,
		chapterNumber int,
		title string,
		content json.RawMessage,
		wordCount int,
		characterCount int,
		authorNotes json.RawMessage,
		isFree bool,
		price *float64,
		currency *string,
		status string,
		displayOrder int,
		scheduledAt *string,
		createdBy uuid.UUID,
	) (*domain.NovelChapter, error)

	// UpdateChapter updates chapter information
	UpdateChapter(
		ctx context.Context,
		id uuid.UUID,
		volumeID *uuid.UUID,
		chapterNumber int,
		title string,
		content json.RawMessage,
		wordCount int,
		characterCount int,
		authorNotes json.RawMessage,
		isFree bool,
		price *float64,
		currency *string,
		status string,
		displayOrder int,
		scheduledAt *string,
		changedBy uuid.UUID,
		requestContext map[string]any,
	) (*domain.NovelChapter, error)

	// DeleteChapter deletes a chapter (soft delete)
	DeleteChapter(ctx context.Context, id uuid.UUID) error

	// GetChapterByID retrieves a chapter by ID
	GetChapterByID(ctx context.Context, id uuid.UUID) (*domain.NovelChapter, error)

	// GetChaptersByNovelID retrieves chapters for a novel with filters
	GetChaptersByNovelID(ctx context.Context, novelID uuid.UUID, filter domain.NovelChapterFilter) ([]*domain.NovelChapter, error)

	// GetChaptersByVolumeID retrieves all chapters for a volume
	GetChaptersByVolumeID(ctx context.Context, volumeID uuid.UUID, publishedOnly bool) ([]*domain.NovelChapter, error)

	// PublishChapter publishes a chapter immediately
	PublishChapter(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error

	// ScheduleChapter schedules a chapter for publication
	ScheduleChapter(ctx context.Context, id uuid.UUID, scheduledAt time.Time) error

	// GetScheduledChapters retrieves chapters scheduled for publication
	GetScheduledChapters(ctx context.Context, before time.Time) ([]*domain.NovelChapter, error)

	// IncrementViewCount increments the view count of a chapter
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// UpdateStatistics updates chapter statistics
	// UpdateStatistics updates chapter statistics
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.NovelChapterStatistics) error

	// GetRecentChapters retrieves recently published chapters
	GetRecentChapters(ctx context.Context, limit int) ([]*domain.NovelChapterSummary, error)
}
