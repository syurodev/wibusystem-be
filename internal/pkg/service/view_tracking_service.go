package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/configs"
	"system/internal/domain"
)

// ViewTrackingService handles view tracking business logic.
// Service này orchestrate tất cả operations cho view tracking system:
// - Deduplication checking
// - Redis buffering
// - Event queuing cho ClickHouse
// - Sync operations từ Redis → PostgreSQL/ClickHouse
type ViewTrackingService struct {
	viewTrackingRepo  domain.ViewTrackingRepository
	viewAnalyticsRepo domain.ViewAnalyticsRepository
	chapterRepo       domain.ChapterRepository
	novelRepo         domain.NovelRepository
	logger            *zap.Logger
	config            *configs.ViewTrackingConfig
}

// NewViewTrackingService creates a new view tracking service instance.
//
// Parameters:
//   - viewTrackingRepo: Redis repository cho deduplication + buffering
//   - viewAnalyticsRepo: ClickHouse repository cho analytics
//   - chapterRepo: Chapter repository cho batch update
//   - novelRepo: Novel repository cho batch update
//   - logger: Zap logger instance
//   - config: ViewTracking configuration
//
// Returns:
//   - *ViewTrackingService: Service instance
func NewViewTrackingService(
	viewTrackingRepo domain.ViewTrackingRepository,
	viewAnalyticsRepo domain.ViewAnalyticsRepository,
	chapterRepo domain.ChapterRepository,
	novelRepo domain.NovelRepository,
	logger *zap.Logger,
	config *configs.ViewTrackingConfig,
) *ViewTrackingService {
	return &ViewTrackingService{
		viewTrackingRepo:  viewTrackingRepo,
		viewAnalyticsRepo: viewAnalyticsRepo,
		chapterRepo:       chapterRepo,
		novelRepo:         novelRepo,
		logger:            logger,
		config:            config,
	}
}

// TrackChapterView tracks một chapter view với deduplication.
// Method này implement toàn bộ flow:
//  1. Get chapter để lấy novel_id (dual tracking)
//  2. Generate deduplication identifier (user_id hoặc IP)
//  3. Check deduplication cho chapter
//  4. Record deduplication marker cho chapter + novel
//  5. Increment Redis buffers cho chapter + novel
//  6. Enqueue 2 events (chapter + novel) vào ClickHouse queue
//
// Parameters:
//   - ctx: Context
//   - chapterID: ID của chapter được xem
//   - userID: ID của user (nil nếu anonymous)
//   - ipAddress: IP address của viewer
//
// Returns:
//   - bool: true nếu view được counted, false nếu duplicate
//   - error: Lỗi nếu có
func (s *ViewTrackingService) TrackChapterView(ctx context.Context, chapterID uuid.UUID, userID *uuid.UUID, ipAddress string) (bool, error) {
	// 1. Get chapter để lấy novel_id cho dual tracking
	chapter, err := s.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return false, fmt.Errorf("failed to get chapter: %w", err)
	}

	// 2. Generate deduplication identifier
	// Ưu tiên user_id, fallback sang IP nếu anonymous
	identifier := ipAddress
	if userID != nil {
		identifier = userID.String()
	}

	// 3. Check deduplication cho chapter
	isDuplicate, err := s.viewTrackingRepo.CheckDeduplication(ctx, "chapter", chapterID, identifier)
	if err != nil {
		s.logger.Warn("Failed to check deduplication, counting view anyway",
			zap.Error(err),
			zap.String("chapter_id", chapterID.String()))
	} else if isDuplicate {
		// View is within deduplication window, skip
		s.logger.Debug("Duplicate view detected, skipping",
			zap.String("chapter_id", chapterID.String()),
			zap.String("identifier", identifier))
		return false, nil
	}

	// 4. Record deduplication marker cho chapter
	if err := s.viewTrackingRepo.RecordDeduplication(ctx, "chapter", chapterID, identifier, s.config.DedupWindowSeconds); err != nil {
		s.logger.Warn("Failed to record chapter deduplication marker",
			zap.Error(err),
			zap.String("chapter_id", chapterID.String()))
	}

	// Also record deduplication marker cho novel (dual tracking)
	if err := s.viewTrackingRepo.RecordDeduplication(ctx, "novel", chapter.NovelID, identifier, s.config.DedupWindowSeconds); err != nil {
		s.logger.Warn("Failed to record novel deduplication marker",
			zap.Error(err),
			zap.String("novel_id", chapter.NovelID.String()))
	}

	// 5. Increment Redis buffers (CHAPTER + NOVEL)
	if err := s.viewTrackingRepo.IncrementBuffer(ctx, "chapter", chapterID); err != nil {
		s.logger.Error("Failed to increment chapter buffer",
			zap.Error(err),
			zap.String("chapter_id", chapterID.String()))
		return false, fmt.Errorf("failed to increment chapter buffer: %w", err)
	}

	if err := s.viewTrackingRepo.IncrementBuffer(ctx, "novel", chapter.NovelID); err != nil {
		s.logger.Error("Failed to increment novel buffer",
			zap.Error(err),
			zap.String("novel_id", chapter.NovelID.String()))
		// Continue even if novel increment fails - không block user request
	}

	// 6. Enqueue event for ClickHouse analytics
	// Fetch enriched metadata
	novel, err := s.novelRepo.GetByID(ctx, chapter.NovelID)
	if err != nil {
		s.logger.Error("Failed to get novel for analytics", zap.Error(err))
		// Continue with minimal data
	}

	var authorIDs, genreIDs, artistIDs, groupIDs []uuid.UUID
	var originalLanguage string
	var ownerID *uuid.UUID
	var studioID *uuid.UUID // Anime only, null for novel

	if novel != nil {
		originalLanguage = *novel.OriginalLanguage
		ownerID = &novel.OwnerID

		// Fetch relations in parallel or sequence
		// Ignoring errors to ensure view is counted even if metadata fails
		authorIDs, _ = s.novelRepo.GetAuthors(ctx, novel.ID)
		genreIDs, _ = s.novelRepo.GetGenres(ctx, novel.ID)
		artistIDs, _ = s.novelRepo.GetArtists(ctx, novel.ID)
		groupIDs, _ = s.novelRepo.GetOrganizationAssignments(ctx, novel.ID)
	}

	// Determine GroupID (Translation Group)
	var groupID *uuid.UUID
	if len(groupIDs) > 0 {
		groupID = &groupIDs[0] // Pick first active group for now
	} else if novel != nil && novel.OwnerType == "tenant" {
		// If owner is a tenant (group), use it as group_id
		groupID = &novel.OwnerID
	}

	// Parse User Agent (Simple implementation)
	// In a real app, use a library like uasurfer or user_agent
	// Here we assume these are passed via context or we parse raw string if available
	// Since TrackChapterView signature only has ipAddress, we can't get UA easily unless we change signature
	// For now, we'll leave them empty or "unknown"
	platform := "web"
	os := "unknown"
	browser := "unknown"

	// Determine User Status
	isPremium := false
	userRole := "guest"
	if userID != nil {
		userRole = "user"
		// TODO: Check if user is premium via UserRepo or cached session
	}

	eventTime := time.Now()
	event := &domain.ViewEvent{
		EventTime: eventTime,
		MediaType: domain.MediaTypeNovel,
		MediaID:   chapter.NovelID,
		UnitID:    chapterID,
		UserID:    userID,
		IPAddress: ipAddress,
		ViewCount: 1,

		// Metadata
		AuthorIDs:        authorIDs,
		GenreIDs:         genreIDs,
		ArtistIDs:        artistIDs,
		TagIDs:           []uuid.UUID{}, // Future proof
		GroupID:          groupID,
		OwnerID:          ownerID,
		StudioID:         studioID,
		OriginalLanguage: originalLanguage,

		// Context
		Platform:    platform,
		OS:          os,
		Browser:     browser,
		CountryCode: "", // Requires GeoIP
		City:        "", // Requires GeoIP
		Referrer:    "", // Requires context

		// Status
		IsPremium: isPremium,
		UserRole:  userRole,
	}

	if err := s.viewTrackingRepo.EnqueueEvent(ctx, event); err != nil {
		s.logger.Warn("Failed to enqueue ClickHouse event",
			zap.Error(err),
			zap.String("chapter_id", chapterID.String()))
	}

	s.logger.Debug("Chapter view tracked successfully",
		zap.String("chapter_id", chapterID.String()),
		zap.String("novel_id", chapter.NovelID.String()),
		zap.String("identifier", identifier))

	return true, nil
}

// SyncBuffersToPostgreSQL syncs Redis buffers sang PostgreSQL.
// Method này được gọi bởi background worker mỗi 5-10 phút.
//
// Flow:
//  1. Get all buffers from Redis (atomic clear)
//  2. Separate chapter vs novel increments
//  3. Batch update chapters using BatchIncrementViewCount
//  4. Batch update novels using BatchIncrementViewCount
//
// Parameters:
//   - ctx: Context
//
// Returns:
//   - error: Lỗi nếu có
func (s *ViewTrackingService) SyncBuffersToPostgreSQL(ctx context.Context) error {
	s.logger.Info("Starting buffer sync to PostgreSQL...")

	// 1. Get all buffers from Redis (atomically clears them)
	buffers, err := s.viewTrackingRepo.GetAllBuffers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get buffers: %w", err)
	}

	if len(buffers) == 0 {
		s.logger.Debug("No buffers to sync")
		return nil
	}

	s.logger.Info("Syncing buffers to PostgreSQL",
		zap.Int("buffer_count", len(buffers)))

	// 2. Separate chapter and novel increments
	chapterIncrements := make(map[uuid.UUID]int64)
	novelIncrements := make(map[uuid.UUID]int64)

	for _, buffer := range buffers {
		if buffer.EntityType == "chapter" {
			chapterIncrements[buffer.EntityID] = buffer.Count
		} else if buffer.EntityType == "novel" {
			novelIncrements[buffer.EntityID] = buffer.Count
		}
	}

	// 3. Batch update chapters
	if len(chapterIncrements) > 0 {
		if err := s.chapterRepo.BatchIncrementViewCount(ctx, chapterIncrements); err != nil {
			s.logger.Error("Failed to batch update chapter view counts", zap.Error(err))
			// TODO: Implement retry/dead-letter-queue logic
			return fmt.Errorf("failed to update chapters: %w", err)
		}
		s.logger.Info("Updated chapter view counts",
			zap.Int("chapter_count", len(chapterIncrements)))
	}

	// 4. Batch update novels
	if len(novelIncrements) > 0 {
		if err := s.novelRepo.BatchIncrementViewCount(ctx, novelIncrements); err != nil {
			s.logger.Error("Failed to batch update novel view counts", zap.Error(err))
			return fmt.Errorf("failed to update novels: %w", err)
		}
		s.logger.Info("Updated novel view counts",
			zap.Int("novel_count", len(novelIncrements)))
	}

	s.logger.Info("Buffer sync to PostgreSQL completed successfully",
		zap.Int("total_entities", len(chapterIncrements)+len(novelIncrements)))

	return nil
}

// SyncEventsToClickHouse syncs events from Redis queue to ClickHouse.
// Method này được gọi bởi background worker.
//
// Flow:
//  1. Dequeue events batch from Redis queue
//  2. Batch insert to ClickHouse
//
// Parameters:
//   - ctx: Context
//
// Returns:
//   - error: Lỗi nếu có
func (s *ViewTrackingService) SyncEventsToClickHouse(ctx context.Context) error {
	s.logger.Debug("Starting event sync to ClickHouse...")

	// 1. Dequeue events from Redis
	events, err := s.viewTrackingRepo.DequeueEvents(ctx, s.config.ClickHouseBatchSize)
	if err != nil {
		return fmt.Errorf("failed to dequeue events: %w", err)
	}

	if len(events) == 0 {
		s.logger.Debug("No events to sync")
		return nil
	}

	s.logger.Info("Syncing events to ClickHouse",
		zap.Int("event_count", len(events)))

	// 2. Batch insert to ClickHouse
	if err := s.viewAnalyticsRepo.BatchInsertEvents(ctx, events); err != nil {
		s.logger.Error("Failed to batch insert events to ClickHouse", zap.Error(err))
		// TODO: Re-enqueue events or send to DLQ
		return fmt.Errorf("failed to insert events: %w", err)
	}

	s.logger.Info("Successfully synced events to ClickHouse",
		zap.Int("event_count", len(events)))

	return nil
}
