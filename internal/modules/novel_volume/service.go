// ============================================================================
// Novel Volume Service
// ============================================================================
//
// Service này cung cấp business logic cho NovelVolume module.
// Volume là cấp giữa trong cấu trúc phân cấp: Novel > Volume > Chapter.
//
// CRUD Operations:
//   - CreateVolume: Tạo volume mới với auto-calculated volume number
//   - UpdateVolume: Cập nhật thông tin volume với history tracking
//   - DeleteVolume: Soft delete volume
//   - GetVolumeByID: Lấy chi tiết volume
//   - GetVolumesByNovelID: Lấy danh sách volumes của một novel
//
// State Operations:
//   - PublishVolume: Xuất bản volume với history tracking
//   - UnpublishVolume: Ẩn volume với history tracking
//   - UpdateDisplayOrder: Cập nhật thứ tự hiển thị
//
// History Tracking:
//   - Tất cả các thay đổi quan trọng (update, publish, unpublish) được
//     log vào VolumeHistory để audit trail
//
// ============================================================================

package novel_volume

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	pkgerrors "system/pkg/errors"
)

// volumeServiceImpl implements VolumeService interface
type volumeServiceImpl struct {
	volumeRepo  domain.NovelVolumeRepository
	novelRepo   domain.NovelRepository
	historyRepo VolumeHistoryRepository
}

// NewVolumeService creates a new instance of VolumeService
func NewService(volumeRepo domain.NovelVolumeRepository, novelRepo domain.NovelRepository, historyRepo VolumeHistoryRepository) VolumeService {
	return &volumeServiceImpl{
		volumeRepo:  volumeRepo,
		novelRepo:   novelRepo,
		historyRepo: historyRepo,
	}
}

// CreateVolume creates a new volume with auto-calculated volume number
func (s *volumeServiceImpl) CreateVolume(ctx context.Context, novelID uuid.UUID, title string, description, coverImageURL *string, displayOrder int, isPublished bool, createdBy uuid.UUID) (*domain.NovelVolume, error) {
	// Validate input
	if title == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "title is required")
	}

	// Get all volumes for this novel to calculate next volume number
	existingVolumes, err := s.volumeRepo.GetByNovelID(ctx, novelID, false)
	if err != nil {
		return nil, err
	}

	// Calculate next volume number (max + 1)
	nextVolumeNumber := 1
	if len(existingVolumes) > 0 {
		maxVolNum := 0
		for _, vol := range existingVolumes {
			if vol.VolumeNumber > maxVolNum {
				maxVolNum = vol.VolumeNumber
			}
		}
		nextVolumeNumber = maxVolNum + 1
	}

	// Generate slug from title
	volumeSlug := slug.Make(title)

	// Create volume
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	// Auto-set display_order if not provided
	if displayOrder == 0 {
		displayOrder = nextVolumeNumber
	}

	volume := &domain.NovelVolume{
		ID:            id,
		NovelID:       novelID,
		VolumeNumber:  nextVolumeNumber,
		Title:         title,
		Slug:          volumeSlug,
		Description:   description,
		CoverImageURL: coverImageURL,
		DisplayOrder:  displayOrder,
		IsPublished:   isPublished,
		CreatedBy:     createdBy,
	}

	if err := s.volumeRepo.Create(ctx, volume); err != nil {
		return nil, err
	}

	// Update novel statistics if volume is published
	if isPublished && s.novelRepo != nil {
		_ = s.novelRepo.UpdateContentStatistics(ctx, novelID)
	}

	// Retrieve the created volume to get timestamps
	return s.volumeRepo.GetByID(ctx, id)
}

// UpdateVolume updates volume information with history tracking
func (s *volumeServiceImpl) UpdateVolume(ctx context.Context, id uuid.UUID, volumeNumber int, title string, description, coverImageURL *string, displayOrder int, isPublished bool, changedBy uuid.UUID, requestContext map[string]any) (*domain.NovelVolume, error) {
	// Validate input
	if title == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "title is required")
	}

	if volumeNumber < 1 {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid volume number")
	}

	// Get existing volume
	oldVolume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerrors.NotFound(I18nNotFound, "volume not found")
		}
		return nil, err
	}

	// Check if volume number changed and if the new number is already in use
	if volumeNumber != oldVolume.VolumeNumber {
		existing, err := s.volumeRepo.GetByNovelIDAndNumber(ctx, oldVolume.NovelID, volumeNumber)
		if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, pkgerrors.Conflict(I18nVolumeNumberExists, "volume number already exists")
		}
	}

	// Generate new slug if title changed
	newSlug := slug.Make(title)

	// Update fields
	newVolume := &domain.NovelVolume{
		ID:            oldVolume.ID,
		NovelID:       oldVolume.NovelID,
		VolumeNumber:  volumeNumber,
		Title:         title,
		Slug:          newSlug,
		Description:   description,
		CoverImageURL: coverImageURL,
		DisplayOrder:  displayOrder,
		IsPublished:   isPublished,
		ChapterCount:  oldVolume.ChapterCount,
		WordCount:     oldVolume.WordCount,
		PublishedAt:   oldVolume.PublishedAt,
		CreatedAt:     oldVolume.CreatedAt,
		UpdatedAt:     oldVolume.UpdatedAt,
		DeletedAt:     oldVolume.DeletedAt,
	}

	// Update volume in database
	if err := s.volumeRepo.Update(ctx, newVolume); err != nil {
		return nil, err
	}

	// Log history if history repository is available
	if s.historyRepo != nil {
		// Log the update with old and new values for comparison
		if err := s.historyRepo.LogUpdate(ctx, id, oldVolume.NovelID, oldVolume, newVolume, changedBy, requestContext); err != nil {
			// Log error but don't fail the update
			// TODO: Add proper logging here
			_ = err
		}
	}

	// Retrieve updated volume
	return s.volumeRepo.GetByID(ctx, id)
}

// DeleteVolume deletes a volume (soft delete)
func (s *volumeServiceImpl) DeleteVolume(ctx context.Context, id uuid.UUID) error {
	// Check if volume exists
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerrors.NotFound(I18nNotFound, "volume not found")
		}
		return err
	}

	// TODO: Check if volume has chapters and prevent deletion if needed
	// For now, we allow deletion regardless

	if err := s.volumeRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Update novel statistics after deletion
	if s.novelRepo != nil {
		_ = s.novelRepo.UpdateContentStatistics(ctx, volume.NovelID)
	}

	return nil
}

// GetVolumeByID retrieves a volume by ID
func (s *volumeServiceImpl) GetVolumeByID(ctx context.Context, id uuid.UUID) (*domain.NovelVolume, error) {
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerrors.NotFound(I18nNotFound, "volume not found")
		}
		return nil, err
	}
	return volume, nil
}

// GetVolumesByNovelID retrieves all volumes for a novel
func (s *volumeServiceImpl) GetVolumesByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*domain.NovelVolume, error) {
	return s.volumeRepo.GetByNovelID(ctx, novelID, publishedOnly)
}

// UpdateDisplayOrder updates the display order of a volume
func (s *volumeServiceImpl) UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error {
	// Check if volume exists
	_, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerrors.NotFound(I18nNotFound, "volume not found")
		}
		return err
	}

	return s.volumeRepo.UpdateDisplayOrder(ctx, id, order)
}

// PublishVolume publishes a volume with history tracking
func (s *volumeServiceImpl) PublishVolume(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Check if volume exists
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerrors.NotFound(I18nNotFound, "volume not found")
		}
		return err
	}

	// Publish the volume
	if err := s.volumeRepo.Publish(ctx, id); err != nil {
		return err
	}

	// Log history if history repository is available
	if s.historyRepo != nil {
		if err := s.historyRepo.LogPublish(ctx, id, volume.NovelID, changedBy, requestContext); err != nil {
			// Log error but don't fail the publish
			_ = err
		}
	}

	// Update novel statistics after publish
	if s.novelRepo != nil {
		_ = s.novelRepo.UpdateContentStatistics(ctx, volume.NovelID)
	}

	return nil
}

// UnpublishVolume unpublishes a volume with history tracking
func (s *volumeServiceImpl) UnpublishVolume(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Check if volume exists
	volume, err := s.volumeRepo.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerrors.NotFound(I18nNotFound, "volume not found")
		}
		return err
	}

	// Unpublish the volume
	if err := s.volumeRepo.Unpublish(ctx, id); err != nil {
		return err
	}

	// Log history if history repository is available
	if s.historyRepo != nil {
		if err := s.historyRepo.LogUnpublish(ctx, id, volume.NovelID, changedBy, requestContext); err != nil {
			// Log error but don't fail the unpublish
			_ = err
		}
	}

	// Update novel statistics after unpublish
	if s.novelRepo != nil {
		_ = s.novelRepo.UpdateContentStatistics(ctx, volume.NovelID)
	}

	return nil
}

// Helper function to detect changed fields between old and new volume
func detectVolumeChangedFields(old, new *domain.NovelVolume) []string {
	var changedFields []string

	if old.VolumeNumber != new.VolumeNumber {
		changedFields = append(changedFields, "volume_number")
	}
	if old.Title != new.Title {
		changedFields = append(changedFields, "title")
	}
	if old.Slug != new.Slug {
		changedFields = append(changedFields, "slug")
	}
	if !stringPtrEqual(old.Description, new.Description) {
		changedFields = append(changedFields, "description")
	}
	if !stringPtrEqual(old.CoverImageURL, new.CoverImageURL) {
		changedFields = append(changedFields, "cover_image_url")
	}
	if old.DisplayOrder != new.DisplayOrder {
		changedFields = append(changedFields, "display_order")
	}
	if old.IsPublished != new.IsPublished {
		changedFields = append(changedFields, "is_published")
	}

	return changedFields
}

// Helper function to generate change summary
func generateVolumeChangeSummary(changedFields []string) string {
	if len(changedFields) == 0 {
		return "No changes"
	}

	fieldDescriptions := map[string]string{
		"volume_number":   "volume number",
		"title":           "title",
		"slug":            "slug",
		"description":     "description",
		"cover_image_url": "cover image",
		"display_order":   "display order",
		"is_published":    "publication status",
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

// Helper function to convert changed fields to JSONB
func changedFieldsToJSON(fields []string) json.RawMessage {
	if len(fields) == 0 {
		return json.RawMessage("[]")
	}

	jsonBytes, err := json.Marshal(fields)
	if err != nil {
		return json.RawMessage("[]")
	}

	return json.RawMessage(jsonBytes)
}

// Helper function to compare string pointers
func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
