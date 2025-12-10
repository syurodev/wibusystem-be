package novel_chapter

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"system/internal/domain"
)

// chapterHistoryRepository implements ChapterHistoryRepository
type chapterHistoryRepository struct {
	pool *pgxpool.Pool
}

// NewChapterHistoryRepository creates a new instance of chapterHistoryRepository
func NewChapterHistoryRepository(pool *pgxpool.Pool) ChapterHistoryRepository {
	return &chapterHistoryRepository{pool: pool}
}

// GetLatestVersion retrieves the latest version number for a chapter
func (r *chapterHistoryRepository) GetLatestVersion(ctx context.Context, chapterID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(version_number), 0) as version
		FROM catalog.novel_chapter_histories
		WHERE chapter_id = $1
	`

	var version int
	err := r.pool.QueryRow(ctx, query, chapterID).Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

// LogUpdate logs a chapter update to history
func (r *chapterHistoryRepository) LogUpdate(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, oldChapter, newChapter *domain.NovelChapter, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, chapterID)
	if err != nil {
		return err
	}
	nextVersion := latestVersion + 1

	// Detect changed fields
	changedFields, contentChanged := detectChapterChangedFields(oldChapter, newChapter)
	if len(changedFields) == 0 {
		return nil // No changes to log
	}

	// Generate change summary
	changeSummary := generateChapterChangeSummary(changedFields, contentChanged)

	// Convert changed fields to JSONB
	changedFieldsJSON, err := json.Marshal(changedFields)
	if err != nil {
		return err
	}

	// Extract request context
	var requestID *string
	var ipAddress *string
	var userAgent *string

	if requestContext != nil {
		if rid, ok := requestContext["request_id"].(string); ok {
			requestID = &rid
		}
		if ip, ok := requestContext["ip_address"].(string); ok {
			ipAddress = &ip
		}
		if ua, ok := requestContext["user_agent"].(string); ok {
			userAgent = &ua
		}
	}

	// Prepare volume_id (may be NULL)
	var volumeIDPtr *uuid.UUID
	if volumeID != uuid.Nil {
		volumeIDPtr = &volumeID
	}

	// Insert history record
	query := `
		INSERT INTO catalog.novel_chapter_histories (
			chapter_id, volume_id, novel_id, version_number, action,
			title, slug, chapter_number, status,
			word_count, character_count,
			changed_fields, change_summary, content_changed,
			changed_by, request_id, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err = r.pool.Exec(ctx, query,
		chapterID,
		volumeIDPtr,
		novelID,
		nextVersion,
		"updated",
		newChapter.Title,
		newChapter.Slug,
		newChapter.ChapterNumber,
		newChapter.Status,
		newChapter.WordCount,
		newChapter.CharacterCount,
		changedFieldsJSON,
		changeSummary,
		contentChanged,
		changedBy,
		requestID,
		ipAddress,
		userAgent,
	)

	return err
}

// LogPublish logs a chapter publish action to history
func (r *chapterHistoryRepository) LogPublish(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, chapterID)
	if err != nil {
		return err
	}
	nextVersion := latestVersion + 1

	// Extract request context
	var requestID *string
	var ipAddress *string
	var userAgent *string

	if requestContext != nil {
		if rid, ok := requestContext["request_id"].(string); ok {
			requestID = &rid
		}
		if ip, ok := requestContext["ip_address"].(string); ok {
			ipAddress = &ip
		}
		if ua, ok := requestContext["user_agent"].(string); ok {
			userAgent = &ua
		}
	}

	// Prepare volume_id (may be NULL)
	var volumeIDPtr *uuid.UUID
	if volumeID != uuid.Nil {
		volumeIDPtr = &volumeID
	}

	// Insert history record
	query := `
		INSERT INTO catalog.novel_chapter_histories (
			chapter_id, volume_id, novel_id, version_number, action,
			changed_fields, change_summary, content_changed,
			changed_by, request_id, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	changedFieldsJSON, _ := json.Marshal([]string{"status"})

	_, err = r.pool.Exec(ctx, query,
		chapterID,
		volumeIDPtr,
		novelID,
		nextVersion,
		"published",
		changedFieldsJSON,
		"Published chapter",
		false, // content not changed
		changedBy,
		requestID,
		ipAddress,
		userAgent,
	)

	return err
}

// Helper functions (duplicated from service for now - can be moved to a shared util package novel_chapter)
// Update: Helper functions removed as they are already defined in service.go within the same package.
