package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

// ChapterService provides business logic for chapters
type ChapterService struct {
	chapterRepo domain.ChapterRepository
	volumeRepo  domain.VolumeRepository
	historyRepo ChapterHistoryRepository
	creatorRepo domain.CreatorRepository
}

// ChapterHistoryRepository interface for logging chapter history
// TODO: Move to domain package once fully implemented
type ChapterHistoryRepository interface {
	LogUpdate(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, oldChapter, newChapter *domain.Chapter, changedBy uuid.UUID, requestContext map[string]any) error
	LogPublish(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error
	GetLatestVersion(ctx context.Context, chapterID uuid.UUID) (int, error)
}

// NewChapterService creates a new instance of ChapterService
func NewChapterService(chapterRepo domain.ChapterRepository, volumeRepo domain.VolumeRepository, historyRepo ChapterHistoryRepository, creatorRepo domain.CreatorRepository) *ChapterService {
	return &ChapterService{
		chapterRepo: chapterRepo,
		volumeRepo:  volumeRepo,
		historyRepo: historyRepo,
		creatorRepo: creatorRepo,
	}
}

// CreateChapter creates a new chapter
func (s *ChapterService) CreateChapter(
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
) (*domain.Chapter, error) {
	// Validate input
	if title == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Validate content JSON
	if len(content) == 0 || string(content) == "null" {
		return nil, pkgerrors.ErrInvalidInput
	}
	if !json.Valid(content) {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Validate author notes JSON if present
	if len(authorNotes) > 0 && string(authorNotes) != "null" {
		if !json.Valid(authorNotes) {
			return nil, pkgerrors.ErrInvalidInput
		}
	} else {
		authorNotes = nil
	}

	// Validate status
	if !isValidChapterStatus(status) {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Check if chapter number already exists in volume
	// If novelID is nil, we need to get it from volume first
	if novelID == uuid.Nil {
		volume, err := s.volumeRepo.GetByID(ctx, volumeID)
		if err != nil {
			return nil, pkgerrors.ErrVolumeNotFound
		}
		novelID = volume.NovelID
	}

	existing, err := s.chapterRepo.GetByNovelIDAndNumber(ctx, novelID, chapterNumber)
	if err == nil && existing != nil {
		return nil, pkgerrors.ErrChapterNumberExists
	}

	// Generate ID
	id, err := uuid.NewV7()
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
	}

	// Generate slug
	// Assuming slugutil.GenerateUniqueSlug is a helper function that generates a unique slug
	// For this example, I'll use the standard slug.Make and assume uniqueness is handled elsewhere or not strictly required here.
	// If slugutil.GenerateUniqueSlug is a custom function, it needs to be defined or imported.
	// For now, using slug.Make directly.
	chapterSlug := slug.Make(title)

	// Parse scheduled time
	var scheduledTime *time.Time
	if status == "scheduled" {
		if scheduledAt == nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		t, err := time.Parse(time.RFC3339, *scheduledAt)
		if err != nil {
			return nil, pkgerrors.ErrInvalidInput
		}
		scheduledTime = &t
	}

	// Calculate word count (simplified)
	// For accurate word count with PlateJS JSON, we would need to traverse the nodes
	// For now, we'll just count based on the raw JSON string length as a rough proxy or 0
	// TODO: Implement proper word counting for PlateJS content
	// Use the word count and character count provided from frontend
	if wordCount == 0 {
		// Fallback if frontend doesn't provide wordCount
		wordCount = 0
	}
	if characterCount == 0 {
		// Fallback if frontend doesn't provide characterCount
		characterCount = 0
	}

	chapter := &domain.Chapter{
		ID:             id,
		NovelID:        novelID,
		VolumeID:       &volumeID, // volumeID is now uuid.UUID, so it needs to be a pointer here
		ChapterNumber:  chapterNumber,
		Title:          title,
		Slug:           chapterSlug,
		Content:        content,
		WordCount:      wordCount,
		CharacterCount: characterCount,
		IsFree:         isFree,
		Price:          price,
		Currency:       currency,
		Status:         domain.ChapterStatus(status),
		DisplayOrder:   displayOrder,
		AuthorNotes:    authorNotes,
		ScheduledAt:    scheduledTime,
		CreatedBy:      createdBy,
		UpdatedBy:      createdBy,
	}

	if err := s.chapterRepo.Create(ctx, chapter); err != nil {
		return nil, err
	}

	// Update volume statistics (chapter_count và word_count)
	if err := s.volumeRepo.UpdateStatistics(ctx, volumeID); err != nil {
		// Log error but don't fail the creation
		// TODO: Add proper logging
		_ = err
	}

	// Update user's last content updated at
	if err := s.creatorRepo.UpdateLastContentUpdatedAt(ctx, createdBy); err != nil {
		fmt.Printf("Failed to update last content updated at for user %s: %v\n", createdBy, err)
		// Non-blocking error, log and continue
	}

	// Retrieve the created chapter to get timestamps
	return s.chapterRepo.GetByID(ctx, id)
}

// UpdateChapter cập nhật thông tin chapter
func (s *ChapterService) UpdateChapter(
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
) (*domain.Chapter, error) {
	// Validate input
	if title == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Validate status
	if !isValidChapterStatus(status) {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Get existing chapter
	oldChapter, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.ErrChapterNotFound
	}

	// Check if chapter number conflicts (if changed)
	if chapterNumber != oldChapter.ChapterNumber {
		existing, err := s.chapterRepo.GetByNovelIDAndNumber(ctx, oldChapter.NovelID, chapterNumber)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.ErrChapterNumberExists
		}
	}

	// Generate new slug if title changed
	newSlug := oldChapter.Slug
	if title != oldChapter.Title {
		// Assuming slugutil.GenerateUniqueSlug is a helper function that generates a unique slug
		// For now, using slug.Make directly for consistency with CreateChapter.
		// If slugutil.GenerateUniqueSlug is a custom function, it needs to be defined or imported.
		newSlug = slug.Make(title)
	}

	// Validate content JSON
	if len(content) > 0 && string(content) != "null" {
		if !json.Valid(content) {
			return nil, pkgerrors.ErrInvalidInput
		}
	} else {
		// If content is empty or "null", keep the old content.
		// If the client explicitly wants to clear content, they should send a valid empty JSON, e.g., `[]` or `{}`.
		content = oldChapter.Content
	}

	// Validate author notes JSON
	if len(authorNotes) > 0 && string(authorNotes) != "null" {
		if !json.Valid(authorNotes) {
			return nil, pkgerrors.ErrInvalidInput
		}
	} else {
		// If authorNotes is empty or "null", keep the old author notes.
		// If the client explicitly wants to clear author notes, they should send a valid empty JSON.
		authorNotes = oldChapter.AuthorNotes
	}

	// Parse scheduled time
	var scheduledTime *time.Time
	if status == "scheduled" {
		if scheduledAt != nil {
			t, err := time.Parse(time.RFC3339, *scheduledAt)
			if err != nil {
				return nil, pkgerrors.ErrInvalidInput
			}
			scheduledTime = &t
		} else {
			scheduledTime = oldChapter.ScheduledAt
		}
	}

	// Calculate word count (simplified)
	// TODO: Implement proper word counting for PlateJS content
	// Use the word count and character count provided from frontend
	if wordCount > 0 {
		// Use the provided values from frontend
		// wordCount and characterCount already assigned from params
	} else {
		// If frontend doesn't provide, keep old values
		wordCount = oldChapter.WordCount
		characterCount = oldChapter.CharacterCount
	}

	// Determine VolumeID
	newVolumeID := oldChapter.VolumeID
	if volumeID != nil {
		newVolumeID = volumeID
	}

	newChapter := &domain.Chapter{
		ID:             id,
		NovelID:        oldChapter.NovelID,
		VolumeID:       newVolumeID,
		ChapterNumber:  chapterNumber,
		Title:          title,
		Slug:           newSlug,
		Content:        content,
		WordCount:      wordCount,
		CharacterCount: characterCount,
		IsFree:         isFree,
		Price:          price,
		Currency:       currency,
		Status:         domain.ChapterStatus(status),
		ViewCount:      oldChapter.ViewCount,
		LikeCount:      oldChapter.LikeCount,
		CommentCount:   oldChapter.CommentCount,
		DisplayOrder:   displayOrder,
		AuthorNotes:    authorNotes,
		PublishedAt:    oldChapter.PublishedAt,
		ScheduledAt:    scheduledTime,
		CreatedAt:      oldChapter.CreatedAt,
		UpdatedAt:      oldChapter.UpdatedAt,
		DeletedAt:      oldChapter.DeletedAt,
		CreatedBy:      oldChapter.CreatedBy,
		UpdatedBy:      changedBy,
	}

	// Update chapter in database
	if err := s.chapterRepo.Update(ctx, newChapter); err != nil {
		return nil, err
	}

	// Update volume statistics if volume changed or word count changed
	if newVolumeID != nil {
		if err := s.volumeRepo.UpdateStatistics(ctx, *newVolumeID); err != nil {
			// Log error but don't fail the update
			_ = err
		}
		// If volume changed, also update the old volume
		if oldChapter.VolumeID != nil && *oldChapter.VolumeID != *newVolumeID {
			if err := s.volumeRepo.UpdateStatistics(ctx, *oldChapter.VolumeID); err != nil {
				_ = err
			}
		}
	} else if oldChapter.WordCount != newChapter.WordCount {
		// Even if volume didn't change, update stats if word count changed
		if oldChapter.VolumeID != nil {
			if err := s.volumeRepo.UpdateStatistics(ctx, *oldChapter.VolumeID); err != nil {
				_ = err
			}
		}
	}

	// Log history if history repository is available
	if s.historyRepo != nil {
		// Get volume and novel IDs for history
		var volID uuid.UUID
		if volumeID != nil {
			volID = *volumeID
		} else if oldChapter.VolumeID != nil {
			volID = *oldChapter.VolumeID
		}

		if err := s.historyRepo.LogUpdate(ctx, id, volID, oldChapter.NovelID, oldChapter, newChapter, changedBy, requestContext); err != nil {
			// Log error but don't fail the update
			_ = err
		}
	}

	// Retrieve updated chapter
	return s.chapterRepo.GetByID(ctx, id)
}

// DeleteChapter deletes a chapter (soft delete)
func (s *ChapterService) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	// Check if chapter exists
	chapter, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrChapterNotFound
		}
		return err
	}

	// Delete the chapter
	if err := s.chapterRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Update volume statistics after deletion
	if chapter.VolumeID != nil {
		if err := s.volumeRepo.UpdateStatistics(ctx, *chapter.VolumeID); err != nil {
			// Log error but don't fail the deletion
			_ = err
		}
	}

	return nil
}

// GetChapterByID retrieves a chapter by ID
func (s *ChapterService) GetChapterByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error) {
	chapter, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgerrors.ErrChapterNotFound
		}
		return nil, err
	}
	return chapter, nil
}

// GetChaptersByNovelID retrieves chapters for a novel with filters
func (s *ChapterService) GetChaptersByNovelID(ctx context.Context, novelID uuid.UUID, filter domain.ChapterFilter) ([]*domain.Chapter, error) {
	return s.chapterRepo.GetByNovelID(ctx, novelID, filter)
}

// GetChaptersByVolumeID retrieves all chapters for a volume
func (s *ChapterService) GetChaptersByVolumeID(ctx context.Context, volumeID uuid.UUID, publishedOnly bool) ([]*domain.Chapter, error) {
	return s.chapterRepo.GetByVolumeID(ctx, volumeID, publishedOnly)
}

// PublishChapter publishes a chapter immediately with history tracking
func (s *ChapterService) PublishChapter(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Check if chapter exists
	chapter, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrChapterNotFound
		}
		return err
	}

	// Publish the chapter
	if err := s.chapterRepo.Publish(ctx, id); err != nil {
		return err
	}

	// Log history if history repository is available
	if s.historyRepo != nil {
		var volID uuid.UUID
		if chapter.VolumeID != nil {
			volID = *chapter.VolumeID
		}

		if err := s.historyRepo.LogPublish(ctx, id, volID, chapter.NovelID, changedBy, requestContext); err != nil {
			// Log error but don't fail the publish
			_ = err
		}
	}

	return nil
}

// ScheduleChapter schedules a chapter for publication
func (s *ChapterService) ScheduleChapter(ctx context.Context, id uuid.UUID, scheduledAt time.Time) error {
	// Check if chapter exists
	_, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrChapterNotFound
		}
		return err
	}

	return s.chapterRepo.Schedule(ctx, id, scheduledAt)
}

// GetScheduledChapters retrieves chapters scheduled for publication before a specific time
func (s *ChapterService) GetScheduledChapters(ctx context.Context, before time.Time) ([]*domain.Chapter, error) {
	return s.chapterRepo.GetScheduledChapters(ctx, before)
}

// IncrementViewCount increments the view count of a chapter
func (s *ChapterService) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.chapterRepo.IncrementViewCount(ctx, id)
}

// UpdateStatistics updates chapter statistics
func (s *ChapterService) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.ChapterStatistics) error {
	// Check if chapter exists
	_, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrChapterNotFound
		}
		return err
	}

	return s.chapterRepo.UpdateStatistics(ctx, id, stats)
}

// Helper function to validate chapter status
func isValidChapterStatus(status string) bool {
	validStatuses := map[string]bool{
		"draft":     true,
		"published": true,
		"scheduled": true,
	}
	return validStatuses[status]
}

// Helper function to count words in text
func countWords(text string) int {
	// Simple word counting - split by whitespace
	words := strings.Fields(text)
	return len(words)
}

// Helper function to extract content from JSONB
func extractContentFromJSON(contentJSON json.RawMessage) string {
	if len(contentJSON) == 0 {
		return ""
	}

	var contentData map[string]any
	if err := json.Unmarshal(contentJSON, &contentData); err == nil {
		if content, ok := contentData["content"].(string); ok && content != "" {
			return content
		}
	}

	return ""
}

// Helper function to detect changed fields between old and new chapter
func detectChapterChangedFields(old, new *domain.Chapter) ([]string, bool) {
	var changedFields []string
	contentChanged := false

	if old.ChapterNumber != new.ChapterNumber {
		changedFields = append(changedFields, "chapter_number")
	}
	if old.Title != new.Title {
		changedFields = append(changedFields, "title")
	}
	if old.Slug != new.Slug {
		changedFields = append(changedFields, "slug")
	}
	if string(old.Content) != string(new.Content) {
		changedFields = append(changedFields, "content")
		contentChanged = true
	}
	if old.IsFree != new.IsFree {
		changedFields = append(changedFields, "is_free")
	}
	if !float64PtrEqual(old.Price, new.Price) {
		changedFields = append(changedFields, "price")
	}
	if !stringPtrEqualChapter(old.Currency, new.Currency) {
		changedFields = append(changedFields, "currency")
	}
	if old.Status != new.Status {
		changedFields = append(changedFields, "status")
	}
	if old.DisplayOrder != new.DisplayOrder {
		changedFields = append(changedFields, "display_order")
	}
	if string(old.AuthorNotes) != string(new.AuthorNotes) {
		changedFields = append(changedFields, "author_notes")
	}

	return changedFields, contentChanged
}

// Helper function to generate change summary for chapters
func generateChapterChangeSummary(changedFields []string, contentChanged bool) string {
	if len(changedFields) == 0 {
		return "No changes"
	}

	fieldDescriptions := map[string]string{
		"chapter_number": "chapter number",
		"title":          "title",
		"slug":           "slug",
		"content":        "content",
		"is_free":        "pricing model",
		"price":          "price",
		"currency":       "currency",
		"status":         "status",
		"display_order":  "display order",
		"author_notes":   "author notes",
	}

	var descriptions []string
	for _, field := range changedFields {
		if desc, ok := fieldDescriptions[field]; ok {
			descriptions = append(descriptions, desc)
		} else {
			descriptions = append(descriptions, field)
		}
	}

	if len(descriptions) == 1 {
		return fmt.Sprintf("Updated %s", descriptions[0])
	}

	return fmt.Sprintf("Updated %s", strings.Join(descriptions, ", "))
}

// Helper function to compare float64 pointers
func float64PtrEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Helper function to compare string pointers
func stringPtrEqualChapter(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
