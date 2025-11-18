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
	historyRepo ChapterHistoryRepository
}

// ChapterHistoryRepository interface for logging chapter history
// TODO: Move to domain package once fully implemented
type ChapterHistoryRepository interface {
	LogUpdate(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, oldChapter, newChapter *domain.Chapter, changedBy uuid.UUID, requestContext map[string]interface{}) error
	LogPublish(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]interface{}) error
	GetLatestVersion(ctx context.Context, chapterID uuid.UUID) (int, error)
}

// NewChapterService creates a new instance of ChapterService
func NewChapterService(chapterRepo domain.ChapterRepository, historyRepo ChapterHistoryRepository) *ChapterService {
	return &ChapterService{
		chapterRepo: chapterRepo,
		historyRepo: historyRepo,
	}
}

// CreateChapter creates a new chapter
func (s *ChapterService) CreateChapter(ctx context.Context, novelID uuid.UUID, volumeID *uuid.UUID, chapterNumber int, title, content string, authorNotes *string, isFree bool, price *float64, currency *string, status string, displayOrder int, scheduledAt *time.Time) (*domain.Chapter, error) {
	// Validate input
	if title == "" || content == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	if chapterNumber < 1 {
		return nil, pkgerrors.ErrInvalidChapterNumber
	}

	// Validate status
	if !isValidChapterStatus(status) {
		return nil, pkgerrors.ErrInvalidChapterStatus
	}

	// Check if chapter number already exists for this novel
	existing, err := s.chapterRepo.GetByNovelIDAndNumber(ctx, novelID, chapterNumber)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	if existing != nil {
		return nil, pkgerrors.ErrChapterNumberExists
	}

	// Generate slug from title
	chapterSlug := slug.Make(title)

	// Prepare content JSON
	contentJSON := prepareContentJSON(content)

	// Prepare author notes JSON
	var authorNotesJSON json.RawMessage
	if authorNotes != nil && *authorNotes != "" {
		authorNotesJSON = prepareContentJSON(*authorNotes)
	} else {
		authorNotesJSON = json.RawMessage("{}")
	}

	// Calculate word and character counts
	wordCount := countWords(content)
	charCount := len(content)

	// Create chapter
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	chapter := &domain.Chapter{
		ID:             id,
		NovelID:        novelID,
		VolumeID:       volumeID,
		ChapterNumber:  chapterNumber,
		Title:          title,
		Slug:           chapterSlug,
		Content:        contentJSON,
		WordCount:      wordCount,
		CharacterCount: charCount,
		IsFree:         isFree,
		Price:          price,
		Currency:       currency,
		Status:         domain.ChapterStatus(status),
		DisplayOrder:   displayOrder,
		AuthorNotes:    authorNotesJSON,
		ScheduledAt:    scheduledAt,
	}

	if err := s.chapterRepo.Create(ctx, chapter); err != nil {
		return nil, err
	}

	// Retrieve the created chapter to get timestamps
	return s.chapterRepo.GetByID(ctx, id)
}

// UpdateChapter updates chapter information with history tracking
func (s *ChapterService) UpdateChapter(ctx context.Context, id uuid.UUID, volumeID *uuid.UUID, chapterNumber int, title, content string, authorNotes *string, isFree bool, price *float64, currency *string, status string, displayOrder int, scheduledAt *time.Time, changedBy uuid.UUID, requestContext map[string]interface{}) (*domain.Chapter, error) {
	// Validate input
	if title == "" || content == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	if chapterNumber < 1 {
		return nil, pkgerrors.ErrInvalidChapterNumber
	}

	// Validate status
	if !isValidChapterStatus(status) {
		return nil, pkgerrors.ErrInvalidChapterStatus
	}

	// Get existing chapter
	oldChapter, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgerrors.ErrChapterNotFound
		}
		return nil, err
	}

	// Check if chapter number changed and if the new number is already in use
	if chapterNumber != oldChapter.ChapterNumber {
		existing, err := s.chapterRepo.GetByNovelIDAndNumber(ctx, oldChapter.NovelID, chapterNumber)
		if err != nil && err != pgx.ErrNoRows {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, pkgerrors.ErrChapterNumberExists
		}
	}

	// Generate new slug if title changed
	newSlug := slug.Make(title)

	// Prepare content JSON
	contentJSON := prepareContentJSON(content)

	// Prepare author notes JSON
	var authorNotesJSON json.RawMessage
	if authorNotes != nil && *authorNotes != "" {
		authorNotesJSON = prepareContentJSON(*authorNotes)
	} else {
		authorNotesJSON = json.RawMessage("{}")
	}

	// Calculate word and character counts
	wordCount := countWords(content)
	charCount := len(content)

	// Update fields
	newChapter := &domain.Chapter{
		ID:             oldChapter.ID,
		NovelID:        oldChapter.NovelID,
		VolumeID:       volumeID,
		ChapterNumber:  chapterNumber,
		Title:          title,
		Slug:           newSlug,
		Content:        contentJSON,
		WordCount:      wordCount,
		CharacterCount: charCount,
		IsFree:         isFree,
		Price:          price,
		Currency:       currency,
		Status:         domain.ChapterStatus(status),
		ViewCount:      oldChapter.ViewCount,
		LikeCount:      oldChapter.LikeCount,
		CommentCount:   oldChapter.CommentCount,
		DisplayOrder:   displayOrder,
		AuthorNotes:    authorNotesJSON,
		PublishedAt:    oldChapter.PublishedAt,
		ScheduledAt:    scheduledAt,
		CreatedAt:      oldChapter.CreatedAt,
		UpdatedAt:      oldChapter.UpdatedAt,
		DeletedAt:      oldChapter.DeletedAt,
	}

	// Update chapter in database
	if err := s.chapterRepo.Update(ctx, newChapter); err != nil {
		return nil, err
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
	_, err := s.chapterRepo.GetByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrChapterNotFound
		}
		return err
	}

	return s.chapterRepo.Delete(ctx, id)
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
func (s *ChapterService) PublishChapter(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]interface{}) error {
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

// Helper function to prepare content JSON
func prepareContentJSON(content string) json.RawMessage {
	// Store as structured JSON
	// Escape quotes in content
	escapedContent := strings.ReplaceAll(content, `"`, `\"`)
	escapedContent = strings.ReplaceAll(escapedContent, "\n", "\\n")
	escapedContent = strings.ReplaceAll(escapedContent, "\r", "\\r")
	escapedContent = strings.ReplaceAll(escapedContent, "\t", "\\t")

	return json.RawMessage(fmt.Sprintf(`{"content": "%s"}`, escapedContent))
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

	var contentData map[string]interface{}
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
