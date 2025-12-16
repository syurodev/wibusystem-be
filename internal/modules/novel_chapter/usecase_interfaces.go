package novel_chapter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// CreateChapterInput represents input for creating a chapter
type CreateChapterInput struct {
	NovelID        uuid.UUID
	VolumeID       uuid.UUID
	ChapterNumber  int
	Title          string
	Content        json.RawMessage
	WordCount      int
	CharacterCount int
	AuthorNotes    json.RawMessage
	IsFree         bool
	Price          *float64
	Currency       *string
	Status         string
	DisplayOrder   int
	ScheduledAt    *string
	CreatedBy      uuid.UUID
}

// UpdateChapterInput represents input for updating a chapter
type UpdateChapterInput struct {
	ID             uuid.UUID
	VolumeID       *uuid.UUID
	ChapterNumber  int
	Title          string
	Content        json.RawMessage
	WordCount      int
	CharacterCount int
	AuthorNotes    json.RawMessage
	IsFree         bool
	Price          *float64
	Currency       *string
	Status         string
	DisplayOrder   int
	ScheduledAt    *string
	ChangedBy      uuid.UUID
	RequestContext map[string]any
}

// DeleteChapterInput represents input for deleting a chapter
type DeleteChapterInput struct {
	ID uuid.UUID
}

// GetChapterInput represents input for getting a chapter
type GetChapterInput struct {
	ID uuid.UUID
}

// ListChaptersByNovelInput represents input for listing chapters by novel
type ListChaptersByNovelInput struct {
	NovelID       uuid.UUID
	Filter        domain.NovelChapterFilter
}

// ListChaptersByVolumeInput represents input for listing chapters by volume
type ListChaptersByVolumeInput struct {
	VolumeID      uuid.UUID
	PublishedOnly bool
}

// PublishChapterInput represents input for publishing a chapter
type PublishChapterInput struct {
	ID             uuid.UUID
	ChangedBy      uuid.UUID
	RequestContext map[string]any
}

// ScheduleChapterInput represents input for scheduling a chapter
type ScheduleChapterInput struct {
	ID          uuid.UUID
	ScheduledAt time.Time
}

// IncrementViewCountInput represents input for incrementing view count
type IncrementViewCountInput struct {
	ID uuid.UUID
}

// UpdateStatisticsInput represents input for updating statistics
type UpdateStatisticsInput struct {
	ID    uuid.UUID
	Stats domain.NovelChapterStatistics
}

// CreateChapterUseCase interface
type CreateChapterUseCase interface {
	Execute(ctx context.Context, input CreateChapterInput) (*domain.NovelChapter, error)
}

// UpdateChapterUseCase interface
type UpdateChapterUseCase interface {
	Execute(ctx context.Context, input UpdateChapterInput) (*domain.NovelChapter, error)
}

// DeleteChapterUseCase interface
type DeleteChapterUseCase interface {
	Execute(ctx context.Context, input DeleteChapterInput) error
}

// GetChapterUseCase interface
type GetChapterUseCase interface {
	Execute(ctx context.Context, input GetChapterInput) (*domain.NovelChapter, error)
}

// ListChaptersByNovelUseCase interface
type ListChaptersByNovelUseCase interface {
	Execute(ctx context.Context, input ListChaptersByNovelInput) ([]*domain.NovelChapter, error)
}

// ListChaptersByVolumeUseCase interface
type ListChaptersByVolumeUseCase interface {
	Execute(ctx context.Context, input ListChaptersByVolumeInput) ([]*domain.NovelChapter, error)
}

// PublishChapterUseCase interface
type PublishChapterUseCase interface {
	Execute(ctx context.Context, input PublishChapterInput) error
}

// ScheduleChapterUseCase interface
type ScheduleChapterUseCase interface {
	Execute(ctx context.Context, input ScheduleChapterInput) error
}

// IncrementViewCountUseCase interface
type IncrementViewCountUseCase interface {
	Execute(ctx context.Context, input IncrementViewCountInput) error
}

// UpdateStatisticsUseCase interface
type UpdateStatisticsUseCase interface {
	Execute(ctx context.Context, input UpdateStatisticsInput) error
}
