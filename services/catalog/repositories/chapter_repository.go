package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// ChapterRepository defines CRUD and listing operations for novel chapters
// This repository handles all database operations related to novel chapters,
// following the repository pattern established in the catalog service.
type ChapterRepository interface {
	// CreateChapter inserts a new chapter for a specific volume
	// Automatically calculates word count, character count, and reading time from content
	// Returns the created chapter with generated ID, timestamps, and calculated metadata
	CreateChapter(ctx context.Context, volumeID uuid.UUID, req d.CreateChapterRequest) (*m.NovelChapter, error)

	// GetChapterByID retrieves a single chapter by its ID
	// Returns error if chapter not found or is deleted
	GetChapterByID(ctx context.Context, id uuid.UUID, includeContent bool) (*m.NovelChapter, error)

	// ListChaptersByVolumeID retrieves all chapters for a specific volume with pagination
	// Returns paginated list of chapters ordered by chapter_number
	ListChaptersByVolumeID(ctx context.Context, volumeID uuid.UUID, req d.ListChaptersRequest) (*d.PaginatedChaptersResponse, error)

	// UpdateChapter modifies an existing chapter
	// Recalculates word count, character count, and reading time if content is updated
	// Increments version number on each update
	// Returns the updated chapter with new timestamps and metadata
	UpdateChapter(ctx context.Context, id uuid.UUID, req d.UpdateChapterRequest) (*m.NovelChapter, error)

	// DeleteChapter performs soft delete on a chapter
	// Sets is_deleted=true and records deletion timestamp
	DeleteChapter(ctx context.Context, id uuid.UUID) error

	// PublishChapter publishes a chapter by setting is_public=true and is_draft=false
	// Sets published_at timestamp if not already published
	PublishChapter(ctx context.Context, id uuid.UUID, publishAt *time.Time) (*m.NovelChapter, error)

	// UnpublishChapter unpublishes a chapter by setting is_public=false
	// Clears published_at timestamp
	UnpublishChapter(ctx context.Context, id uuid.UUID) (*m.NovelChapter, error)

	// CheckChapterPurchases checks if any users have purchased this specific chapter
	// Used to prevent deletion of chapters that users have paid for
	CheckChapterPurchases(ctx context.Context, chapterID uuid.UUID) (bool, error)
}

// chapterRepository implements ChapterRepository interface
type chapterRepository struct {
	pool *pgxpool.Pool
}

// NewChapterRepository creates a new chapter repository instance
// Takes a postgres connection pool for database operations
func NewChapterRepository(pool *pgxpool.Pool) ChapterRepository {
	return &chapterRepository{pool: pool}
}

// calculateContentMetadata calculates word count, character count, and reading time from chapter content
// This is a helper function used during create and update operations
// Returns word count, character count, and reading time in minutes
func calculateContentMetadata(content *json.RawMessage) (wordCount, charCount, readingTime int) {
	if content == nil {
		return 0, 0, 0
	}

	// Convert JSONB content to string for counting
	contentStr := string(*content)

	// Count characters (UTF-8 aware)
	charCount = utf8.RuneCountInString(contentStr)

	// Count words (simple split by whitespace)
	words := strings.Fields(contentStr)
	wordCount = len(words)

	// Calculate reading time (assuming 200 words per minute average reading speed)
	if wordCount > 0 {
		readingTime = (wordCount + 199) / 200 // Round up
	}

	return wordCount, charCount, readingTime
}

// CreateChapter inserts a new chapter for a specific volume
// This method validates the volume exists and chapter number is unique before creating
func (r *chapterRepository) CreateChapter(ctx context.Context, volumeID uuid.UUID, req d.CreateChapterRequest) (*m.NovelChapter, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify volume exists and is not deleted
	var volumeExists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM novel_volume WHERE id = $1 AND is_deleted = FALSE)", volumeID).Scan(&volumeExists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify volume existence: %w", err)
	}
	if !volumeExists {
		return nil, fmt.Errorf("volume not found or already deleted")
	}

	// Check if chapter number already exists for this volume
	var chapterExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM novel_chapter WHERE volume_id = $1 AND chapter_number = $2 AND is_deleted = FALSE)",
		volumeID, req.ChapterNumber,
	).Scan(&chapterExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check chapter number: %w", err)
	}
	if chapterExists {
		return nil, fmt.Errorf("chapter number %d already exists for this volume", req.ChapterNumber)
	}

	// Calculate content metadata
	wordCount, charCount, readingTime := calculateContentMetadata(req.Content)

	// Insert chapter record
	chapterID := uuid.New()
	var chapter m.NovelChapter

	// Set published_at if not draft and is_public
	var publishedAt *time.Time
	if !req.IsDraft && req.IsPublic {
		now := time.Now()
		publishedAt = &now
	}

	query := `
		INSERT INTO novel_chapter (
			id, volume_id, chapter_number, title, content,
			published_at, scheduled_publish_at, is_draft, is_public,
			content_warnings, has_mature_content, price_coins,
			word_count, character_count, reading_time_minutes,
			version, view_count, like_count, comment_count,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 1, 0, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING
			id, volume_id, chapter_number, title, content,
			created_by_user_id, updated_by_user_id,
			published_at, scheduled_publish_at, is_draft, is_public, is_deleted, deleted_at,
			version, content_warnings, has_mature_content, price_coins,
			word_count, character_count, reading_time_minutes,
			view_count, like_count, comment_count,
			created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		chapterID, volumeID, req.ChapterNumber, req.Title, req.Content,
		publishedAt, req.ScheduledPublishAt, req.IsDraft, req.IsPublic,
		req.ContentWarnings, req.HasMatureContent, req.PriceCoins,
		wordCount, charCount, readingTime,
	).Scan(
		&chapter.ID, &chapter.VolumeID, &chapter.ChapterNumber, &chapter.Title, &chapter.Content,
		&chapter.CreatedByUserID, &chapter.UpdatedByUserID,
		&chapter.PublishedAt, &chapter.ScheduledPublishAt, &chapter.IsDraft, &chapter.IsPublic, &chapter.IsDeleted, &chapter.DeletedAt,
		&chapter.Version, &chapter.ContentWarnings, &chapter.HasMatureContent, &chapter.PriceCoins,
		&chapter.WordCount, &chapter.CharacterCount, &chapter.ReadingTimeMinutes,
		&chapter.ViewCount, &chapter.LikeCount, &chapter.CommentCount,
		&chapter.CreatedAt, &chapter.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chapter: %w", err)
	}

	// Update chapter count in volume
	_, err = tx.Exec(ctx, `
		UPDATE novel_volume
		SET chapter_count = (SELECT COUNT(*) FROM novel_chapter WHERE volume_id = $1 AND is_deleted = FALSE),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, volumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to update volume chapter count: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &chapter, nil
}

// GetChapterByID retrieves a single chapter by its ID
// Returns error if chapter is not found or has been soft-deleted
// Content field is only populated if includeContent is true
func (r *chapterRepository) GetChapterByID(ctx context.Context, id uuid.UUID, includeContent bool) (*m.NovelChapter, error) {
	var chapter m.NovelChapter

	// Build query dynamically based on includeContent flag
	contentField := "NULL::jsonb as content"
	if includeContent {
		contentField = "content"
	}

	query := fmt.Sprintf(`
		SELECT
			id, volume_id, chapter_number, title, %s,
			created_by_user_id, updated_by_user_id,
			published_at, scheduled_publish_at, is_draft, is_public, is_deleted, deleted_at,
			version, content_warnings, has_mature_content, price_coins,
			word_count, character_count, reading_time_minutes,
			view_count, like_count, comment_count,
			created_at, updated_at
		FROM novel_chapter
		WHERE id = $1 AND is_deleted = FALSE
	`, contentField)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&chapter.ID, &chapter.VolumeID, &chapter.ChapterNumber, &chapter.Title, &chapter.Content,
		&chapter.CreatedByUserID, &chapter.UpdatedByUserID,
		&chapter.PublishedAt, &chapter.ScheduledPublishAt, &chapter.IsDraft, &chapter.IsPublic, &chapter.IsDeleted, &chapter.DeletedAt,
		&chapter.Version, &chapter.ContentWarnings, &chapter.HasMatureContent, &chapter.PriceCoins,
		&chapter.WordCount, &chapter.CharacterCount, &chapter.ReadingTimeMinutes,
		&chapter.ViewCount, &chapter.LikeCount, &chapter.CommentCount,
		&chapter.CreatedAt, &chapter.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("chapter not found")
		}
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	return &chapter, nil
}

// ListChaptersByVolumeID retrieves all chapters for a specific volume with pagination
// Results are ordered by chapter_number ascending
// Content field is only populated if IncludeContent is true
func (r *chapterRepository) ListChaptersByVolumeID(ctx context.Context, volumeID uuid.UUID, req d.ListChaptersRequest) (*d.PaginatedChaptersResponse, error) {
	// Set pagination defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Build query dynamically based on IncludeContent flag
	contentField := "NULL::jsonb as content"
	if req.IncludeContent {
		contentField = "content"
	}

	query := fmt.Sprintf(`
		SELECT
			id, volume_id, chapter_number, title, %s,
			published_at, is_draft, is_public,
			price_coins, word_count, character_count, reading_time_minutes,
			view_count, like_count, comment_count,
			content_warnings, has_mature_content, version,
			created_at, updated_at
		FROM novel_chapter
		WHERE volume_id = $1 AND is_deleted = FALSE
		ORDER BY chapter_number ASC
		LIMIT $2 OFFSET $3
	`, contentField)

	offset := (req.Page - 1) * req.Limit
	rows, err := r.pool.Query(ctx, query, volumeID, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapters: %w", err)
	}
	defer rows.Close()

	var chapters []d.ChapterResponse
	for rows.Next() {
		var chapter d.ChapterResponse
		var id, volumeIDStr string
		var contentWarningsBytes *json.RawMessage
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &volumeIDStr, &chapter.ChapterNumber, &chapter.Title, &chapter.Content,
			&chapter.PublishedAt, &chapter.IsDraft, &chapter.IsPublic,
			&chapter.PriceCoins, &chapter.WordCount, &chapter.CharacterCount, &chapter.ReadingTimeMinutes,
			&chapter.ViewCount, &chapter.LikeCount, &chapter.CommentCount,
			&contentWarningsBytes, &chapter.HasMatureContent, &chapter.Version,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chapter: %w", err)
		}

		chapter.ID = id
		chapter.VolumeID = volumeIDStr
		chapter.CreatedAt = createdAt
		chapter.UpdatedAt = updatedAt

		// Parse content warnings JSON
		if contentWarningsBytes != nil {
			var warnings []string
			if err := json.Unmarshal(*contentWarningsBytes, &warnings); err == nil {
				chapter.ContentWarnings = warnings
			} else {
				chapter.ContentWarnings = []string{}
			}
		} else {
			chapter.ContentWarnings = []string{}
		}

		chapters = append(chapters, chapter)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate chapters: %w", rows.Err())
	}

	// Get total count for pagination
	var total int64
	countQuery := `SELECT COUNT(*) FROM novel_chapter WHERE volume_id = $1 AND is_deleted = FALSE`
	err = r.pool.QueryRow(ctx, countQuery, volumeID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrevious := req.Page > 1

	return &d.PaginatedChaptersResponse{
		Chapters: chapters,
		Pagination: d.PaginationMeta{
			Page:        req.Page,
			PageSize:    req.Limit,
			Total:       total,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// UpdateChapter modifies an existing chapter
// Only updates fields that are provided in the request (non-nil values)
// Recalculates content metadata if content is updated
// Increments version number on each update
func (r *chapterRepository) UpdateChapter(ctx context.Context, id uuid.UUID, req d.UpdateChapterRequest) (*m.NovelChapter, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	// Always update updated_at and increment version
	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	updateFields = append(updateFields, fmt.Sprintf("version = version + 1"))

	// Update fields if provided
	if req.Title != nil {
		updateFields = append(updateFields, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Content != nil {
		updateFields = append(updateFields, fmt.Sprintf("content = $%d", argIndex))
		args = append(args, *req.Content)
		argIndex++

		// Recalculate content metadata
		wordCount, charCount, readingTime := calculateContentMetadata(req.Content)
		updateFields = append(updateFields, fmt.Sprintf("word_count = $%d", argIndex))
		args = append(args, wordCount)
		argIndex++

		updateFields = append(updateFields, fmt.Sprintf("character_count = $%d", argIndex))
		args = append(args, charCount)
		argIndex++

		updateFields = append(updateFields, fmt.Sprintf("reading_time_minutes = $%d", argIndex))
		args = append(args, readingTime)
		argIndex++
	}

	if req.IsDraft != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_draft = $%d", argIndex))
		args = append(args, *req.IsDraft)
		argIndex++
	}

	if req.IsPublic != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if req.PriceCoins != nil {
		updateFields = append(updateFields, fmt.Sprintf("price_coins = $%d", argIndex))
		args = append(args, *req.PriceCoins)
		argIndex++
	}

	if req.ContentWarnings != nil {
		updateFields = append(updateFields, fmt.Sprintf("content_warnings = $%d", argIndex))
		args = append(args, *req.ContentWarnings)
		argIndex++
	}

	if req.HasMatureContent != nil {
		updateFields = append(updateFields, fmt.Sprintf("has_mature_content = $%d", argIndex))
		args = append(args, *req.HasMatureContent)
		argIndex++
	}

	if len(updateFields) == 2 { // Only updated_at and version fields
		return nil, fmt.Errorf("no fields to update")
	}

	// Build and execute update query
	query := fmt.Sprintf(`
		UPDATE novel_chapter
		SET %s
		WHERE id = $%d AND is_deleted = FALSE
		RETURNING
			id, volume_id, chapter_number, title, content,
			created_by_user_id, updated_by_user_id,
			published_at, scheduled_publish_at, is_draft, is_public, is_deleted, deleted_at,
			version, content_warnings, has_mature_content, price_coins,
			word_count, character_count, reading_time_minutes,
			view_count, like_count, comment_count,
			created_at, updated_at
	`, strings.Join(updateFields, ", "), argIndex)

	args = append(args, id)

	var chapter m.NovelChapter
	err = tx.QueryRow(ctx, query, args...).Scan(
		&chapter.ID, &chapter.VolumeID, &chapter.ChapterNumber, &chapter.Title, &chapter.Content,
		&chapter.CreatedByUserID, &chapter.UpdatedByUserID,
		&chapter.PublishedAt, &chapter.ScheduledPublishAt, &chapter.IsDraft, &chapter.IsPublic, &chapter.IsDeleted, &chapter.DeletedAt,
		&chapter.Version, &chapter.ContentWarnings, &chapter.HasMatureContent, &chapter.PriceCoins,
		&chapter.WordCount, &chapter.CharacterCount, &chapter.ReadingTimeMinutes,
		&chapter.ViewCount, &chapter.LikeCount, &chapter.CommentCount,
		&chapter.CreatedAt, &chapter.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("chapter not found or already deleted")
		}
		return nil, fmt.Errorf("failed to update chapter: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &chapter, nil
}

// CheckChapterPurchases checks if any users have purchased this specific chapter
// Returns true if there are any purchase records for this chapter
func (r *chapterRepository) CheckChapterPurchases(ctx context.Context, chapterID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_content_purchases ucp
			WHERE ucp.item_type = 'NOVEL_CHAPTER' AND ucp.item_id = $1
		)
	`

	var hasPurchases bool
	err := r.pool.QueryRow(ctx, query, chapterID).Scan(&hasPurchases)
	if err != nil {
		return false, fmt.Errorf("failed to check chapter purchases: %w", err)
	}

	return hasPurchases, nil
}

// DeleteChapter performs soft delete on a chapter
// Fails if any users have purchased this chapter
func (r *chapterRepository) DeleteChapter(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get volume_id before deletion
	var volumeID uuid.UUID
	var exists bool
	err = tx.QueryRow(ctx, `
		SELECT volume_id, EXISTS(SELECT 1 FROM novel_chapter WHERE id = $1 AND is_deleted = FALSE)
		FROM novel_chapter
		WHERE id = $1 AND is_deleted = FALSE
	`, id).Scan(&volumeID, &exists)
	if err != nil {
		return fmt.Errorf("failed to check chapter existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("chapter not found or already deleted")
	}

	// Check if any users have purchased this chapter
	hasPurchases, err := r.CheckChapterPurchases(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check purchases: %w", err)
	}
	if hasPurchases {
		return fmt.Errorf("cannot delete chapter: users have purchased this chapter")
	}

	// Perform soft delete on the chapter
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE novel_chapter
		SET is_deleted = TRUE, deleted_at = $2, updated_at = $2
		WHERE id = $1
	`, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}

	// Update chapter count in volume
	_, err = tx.Exec(ctx, `
		UPDATE novel_volume
		SET chapter_count = (SELECT COUNT(*) FROM novel_chapter WHERE volume_id = $1 AND is_deleted = FALSE),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, volumeID)
	if err != nil {
		return fmt.Errorf("failed to update volume chapter count: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PublishChapter publishes a chapter by setting is_public=true, is_draft=false and published_at
// If publishAt is provided, uses that timestamp; otherwise uses current time
func (r *chapterRepository) PublishChapter(ctx context.Context, id uuid.UUID, publishAt *time.Time) (*m.NovelChapter, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use provided publishAt or default to now
	var pubTime time.Time
	if publishAt != nil {
		pubTime = *publishAt
	} else {
		pubTime = time.Now()
	}

	var chapter m.NovelChapter
	query := `
		UPDATE novel_chapter
		SET is_public = TRUE, is_draft = FALSE, published_at = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_deleted = FALSE
		RETURNING
			id, volume_id, chapter_number, title, content,
			created_by_user_id, updated_by_user_id,
			published_at, scheduled_publish_at, is_draft, is_public, is_deleted, deleted_at,
			version, content_warnings, has_mature_content, price_coins,
			word_count, character_count, reading_time_minutes,
			view_count, like_count, comment_count,
			created_at, updated_at
	`

	err = tx.QueryRow(ctx, query, id, pubTime).Scan(
		&chapter.ID, &chapter.VolumeID, &chapter.ChapterNumber, &chapter.Title, &chapter.Content,
		&chapter.CreatedByUserID, &chapter.UpdatedByUserID,
		&chapter.PublishedAt, &chapter.ScheduledPublishAt, &chapter.IsDraft, &chapter.IsPublic, &chapter.IsDeleted, &chapter.DeletedAt,
		&chapter.Version, &chapter.ContentWarnings, &chapter.HasMatureContent, &chapter.PriceCoins,
		&chapter.WordCount, &chapter.CharacterCount, &chapter.ReadingTimeMinutes,
		&chapter.ViewCount, &chapter.LikeCount, &chapter.CommentCount,
		&chapter.CreatedAt, &chapter.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("chapter not found or already deleted")
		}
		return nil, fmt.Errorf("failed to publish chapter: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &chapter, nil
}

// UnpublishChapter unpublishes a chapter by setting is_public=false and clearing published_at
func (r *chapterRepository) UnpublishChapter(ctx context.Context, id uuid.UUID) (*m.NovelChapter, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var chapter m.NovelChapter
	query := `
		UPDATE novel_chapter
		SET is_public = FALSE, published_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_deleted = FALSE
		RETURNING
			id, volume_id, chapter_number, title, content,
			created_by_user_id, updated_by_user_id,
			published_at, scheduled_publish_at, is_draft, is_public, is_deleted, deleted_at,
			version, content_warnings, has_mature_content, price_coins,
			word_count, character_count, reading_time_minutes,
			view_count, like_count, comment_count,
			created_at, updated_at
	`

	err = tx.QueryRow(ctx, query, id).Scan(
		&chapter.ID, &chapter.VolumeID, &chapter.ChapterNumber, &chapter.Title, &chapter.Content,
		&chapter.CreatedByUserID, &chapter.UpdatedByUserID,
		&chapter.PublishedAt, &chapter.ScheduledPublishAt, &chapter.IsDraft, &chapter.IsPublic, &chapter.IsDeleted, &chapter.DeletedAt,
		&chapter.Version, &chapter.ContentWarnings, &chapter.HasMatureContent, &chapter.PriceCoins,
		&chapter.WordCount, &chapter.CharacterCount, &chapter.ReadingTimeMinutes,
		&chapter.ViewCount, &chapter.LikeCount, &chapter.CommentCount,
		&chapter.CreatedAt, &chapter.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("chapter not found or already deleted")
		}
		return nil, fmt.Errorf("failed to unpublish chapter: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &chapter, nil
}