/*
Media Progress Service - Business Logic Layer
==============================================

ARCHITECTURE:
─────────────
Service layer chứa toàn bộ business logic cho media progress.
Service này orchestrate các operations giữa repository và external dependencies.

FLOW: Khi User Đọc Chapter
━━━━━━━━━━━━━━━━━━━━━━━━━
1. Frontend POST /api/v1/history với {content_id, unit_id, position}
   │
   ▼
2. Handler parse request → gọi Service.UpdateProgress()
   │
   ▼
3. Service.UpdateProgress():
   a. Get novel info để lấy total_chapters (cho tính %)
   b. Upsert UnitProgress (track chapter này đang đọc)
   c. Upsert MediaProgress (update current position)
   │
   ▼
4. Return MediaProgress với đầy đủ thông tin

FLOW: Đánh dấu Chapter Đã Đọc Xong
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. User scroll tới cuối chapter
   │
   ▼
2. Frontend POST /api/v1/progress/{type}/{id}/units/{unit_id}/complete
   │
   ▼
3. Handler → Service.MarkUnitComplete():
   a. Update unit_progress.status = 'completed'
   b. Recalculate completed_units trong media_progress
   c. Update progress_percentage

DEPENDENCIES:
─────────────
- MediaProgressRepository: Data access
- NovelRepository: Get novel info (total chapters)
- ChapterRepository: Get chapter info
*/

package media_progress

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/domain"
)

// service implements the Service interface
type service struct {
	repo       domain.MediaProgressRepository
	novelRepo  domain.NovelRepository
	chapterRepo domain.NovelChapterRepository
	logger     *zap.Logger
}

// NewService creates a new media progress service.
//
// Parameters:
//   - repo: Media progress repository
//   - novelRepo: Novel repository (để get total_chapters)
//   - chapterRepo: Chapter repository (để validate chapter)
//   - logger: Zap logger
//
// Returns:
//   - Service: Service interface implementation
func NewService(
	repo domain.MediaProgressRepository,
	novelRepo domain.NovelRepository,
	chapterRepo domain.NovelChapterRepository,
	logger *zap.Logger,
) Service {
	return &service{
		repo:        repo,
		novelRepo:   novelRepo,
		chapterRepo: chapterRepo,
		logger:      logger,
	}
}

// =============================================================================
// SERVICE METHODS
// =============================================================================

// UpdateProgress updates reading/watching progress for a user.
//
// DETAILED FLOW:
// ──────────────
// 1. Parse position data thành JSON
// 2. Get novel info để lấy total_chapters
// 3. Tạo UnitProgress record (track chapter này)
// 4. Tạo MediaProgress record (update current position)
// 5. Return updated MediaProgress
func (s *service) UpdateProgress(ctx context.Context, input UpdateProgressInput) (*domain.MediaProgress, error) {
	s.logger.Debug("Updating media progress",
		zap.String("user_id", input.UserID.String()),
		zap.String("media_type", input.MediaType),
		zap.String("media_id", input.MediaID.String()),
		zap.String("unit_id", input.UnitID.String()),
	)

	// 1. Parse position to JSON
	var positionJSON json.RawMessage
	if input.Position != nil {
		posBytes, err := json.Marshal(input.Position)
		if err != nil {
			s.logger.Error("Failed to marshal position", zap.Error(err))
			positionJSON = json.RawMessage("{}")
		} else {
			positionJSON = posBytes
		}
	} else {
		positionJSON = json.RawMessage("{}")
	}

	// 2. Get total units (chapters) for this media
	var totalUnits int
	if input.MediaType == domain.MediaTypeNovel {
		novel, err := s.novelRepo.GetByID(ctx, input.MediaID)
		if err != nil {
			s.logger.Warn("Could not get novel info, using 0 for total_chapters",
				zap.String("novel_id", input.MediaID.String()),
				zap.Error(err),
			)
		} else {
			totalUnits = novel.TotalChapters
		}
	}

	// 3. Upsert UnitProgress - track this specific chapter
	unitProgress := &domain.UnitProgress{
		UserID:    input.UserID,
		MediaType: input.MediaType,
		MediaID:   input.MediaID,
		UnitID:    input.UnitID,
		Status:    domain.UnitStatusInProgress,
		Position:  positionJSON,
	}

	if err := s.repo.UpsertUnitProgress(ctx, unitProgress); err != nil {
		s.logger.Error("Failed to upsert unit progress", zap.Error(err))
		return nil, err
	}

	// 4. Upsert MediaProgress - update overall progress
	mediaProgress := &domain.MediaProgress{
		UserID:        input.UserID,
		MediaType:     input.MediaType,
		MediaID:       input.MediaID,
		CurrentUnitID: input.UnitID,
		Position:      positionJSON,
		TotalUnits:    totalUnits,
	}

	if err := s.repo.UpsertMediaProgress(ctx, mediaProgress); err != nil {
		s.logger.Error("Failed to upsert media progress", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Media progress updated successfully",
		zap.String("progress_id", mediaProgress.ID.String()),
	)

	return mediaProgress, nil
}

// GetRecentProgress retrieves N most recent progress entries for "Continue" section.
//
// FLOW:
// 1. Validate limit (default 12, max 50)
// 2. Query repository với JOINs để lấy đầy đủ info
// 3. Return list
func (s *service) GetRecentProgress(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.MediaProgress, error) {
	// Validate limit
	if limit <= 0 {
		limit = 12 // default
	}
	if limit > 50 {
		limit = 50 // max
	}

	s.logger.Debug("Getting recent progress",
		zap.String("user_id", userID.String()),
		zap.Int("limit", limit),
	)

	result, err := s.repo.GetRecentMediaProgress(ctx, userID, limit)
	if err != nil {
		s.logger.Error("Failed to get recent media progress", zap.Error(err))
		return nil, err
	}

	s.logger.Debug("Got recent progress",
		zap.Int("count", len(result)),
	)

	return result, nil
}

// GetProgressByMedia retrieves progress for a specific media.
func (s *service) GetProgressByMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) (*domain.MediaProgress, error) {
	return s.repo.GetMediaProgressByUserAndMedia(ctx, userID, mediaType, mediaID)
}

// GetProgressList retrieves paginated progress for a user.
func (s *service) GetProgressList(ctx context.Context, userID uuid.UUID, filter domain.MediaProgressFilter) ([]*domain.MediaProgress, int64, error) {
	// Apply defaults
	if filter.Limit <= 0 {
		filter.Limit = 15
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.repo.GetMediaProgressByUserID(ctx, userID, filter)
}

// GetUnitProgress retrieves read status for all chapters in a media.
//
// FLOW:
// 1. Query unit_progress for this media
// 2. Return list of UnitProgress
//
// Use case: Chapter list với icon "đã đọc"
func (s *service) GetUnitProgress(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) ([]*domain.UnitProgress, error) {
	s.logger.Debug("Getting unit progress",
		zap.String("user_id", userID.String()),
		zap.String("media_type", mediaType),
		zap.String("media_id", mediaID.String()),
	)

	return s.repo.GetUnitProgressByMedia(ctx, userID, mediaType, mediaID)
}

// MarkUnitComplete marks a chapter/episode as completed.
//
// DETAILED FLOW:
// ──────────────
// 1. Update unit_progress.status = 'completed'
// 2. Recalculate completed_units trong media_progress
// 3. Update progress_percentage
func (s *service) MarkUnitComplete(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID, unitID uuid.UUID) error {
	s.logger.Info("Marking unit as complete",
		zap.String("user_id", userID.String()),
		zap.String("unit_id", unitID.String()),
	)

	// 1. Mark the unit as complete
	if err := s.repo.MarkUnitComplete(ctx, userID, unitID); err != nil {
		s.logger.Error("Failed to mark unit complete", zap.Error(err))
		return err
	}

	// 2. Update completed count in media_progress
	if err := s.repo.UpdateCompletedUnitsCount(ctx, userID, mediaType, mediaID); err != nil {
		s.logger.Error("Failed to update completed count", zap.Error(err))
		return err
	}

	return nil
}

// DeleteProgress deletes progress for a specific media.
func (s *service) DeleteProgress(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("Deleting media progress", zap.String("id", id.String()))
	return s.repo.DeleteMediaProgress(ctx, id)
}

// ClearAllProgress clears all progress for a user.
func (s *service) ClearAllProgress(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("Clearing all progress for user", zap.String("user_id", userID.String()))
	return s.repo.DeleteAllMediaProgressByUser(ctx, userID)
}
