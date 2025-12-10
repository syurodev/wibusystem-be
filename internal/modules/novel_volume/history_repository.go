package novel_volume

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"system/internal/domain"
)

// volumeHistoryRepository implements VolumeHistoryRepository
type volumeHistoryRepository struct {
	pool *pgxpool.Pool
}

// NewVolumeHistoryRepository creates a new instance of volumeHistoryRepository
func NewVolumeHistoryRepository(pool *pgxpool.Pool) VolumeHistoryRepository {
	return &volumeHistoryRepository{pool: pool}
}

// GetLatestVersion retrieves the latest version number for a volume
func (r *volumeHistoryRepository) GetLatestVersion(ctx context.Context, volumeID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(version_number), 0) as version
		FROM catalog.novel_volume_histories
		WHERE volume_id = $1
	`

	var version int
	err := r.pool.QueryRow(ctx, query, volumeID).Scan(&version)
	if err != nil {
		return 0, err
	}

	return version, nil
}

// LogUpdate logs a volume update to history
func (r *volumeHistoryRepository) LogUpdate(ctx context.Context, volumeID, novelID uuid.UUID, oldVolume, newVolume *domain.Volume, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, volumeID)
	if err != nil {
		return err
	}
	nextVersion := latestVersion + 1

	// Detect changed fields
	changedFields := detectVolumeChangedFields(oldVolume, newVolume)
	if len(changedFields) == 0 {
		return nil // No changes to log
	}

	// Generate change summary
	changeSummary := generateVolumeChangeSummary(changedFields)

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

	// Insert history record
	query := `
		INSERT INTO catalog.novel_volume_histories (
			volume_id, novel_id, version_number, action,
			title, slug, volume_number, is_published,
			chapter_count, word_count,
			changed_fields, change_summary,
			changed_by, request_id, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = r.pool.Exec(ctx, query,
		volumeID,
		novelID,
		nextVersion,
		"updated",
		newVolume.Title,
		newVolume.Slug,
		newVolume.VolumeNumber,
		newVolume.IsPublished,
		newVolume.ChapterCount,
		newVolume.WordCount,
		changedFieldsJSON,
		changeSummary,
		changedBy,
		requestID,
		ipAddress,
		userAgent,
	)

	return err
}

// LogPublish logs a volume publish action to history
func (r *volumeHistoryRepository) LogPublish(ctx context.Context, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, volumeID)
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

	// Insert history record
	query := `
		INSERT INTO catalog.novel_volume_histories (
			volume_id, novel_id, version_number, action,
			changed_fields, change_summary,
			changed_by, request_id, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	changedFieldsJSON, _ := json.Marshal([]string{"is_published"})

	_, err = r.pool.Exec(ctx, query,
		volumeID,
		novelID,
		nextVersion,
		"published",
		changedFieldsJSON,
		"Published volume",
		changedBy,
		requestID,
		ipAddress,
		userAgent,
	)

	return err
}

// LogUnpublish logs a volume unpublish action to history
func (r *volumeHistoryRepository) LogUnpublish(ctx context.Context, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, volumeID)
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

	// Insert history record
	query := `
		INSERT INTO catalog.novel_volume_histories (
			volume_id, novel_id, version_number, action,
			changed_fields, change_summary,
			changed_by, request_id, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	changedFieldsJSON, _ := json.Marshal([]string{"is_published"})

	_, err = r.pool.Exec(ctx, query,
		volumeID,
		novelID,
		nextVersion,
		"unpublished",
		changedFieldsJSON,
		"Unpublished volume",
		changedBy,
		requestID,
		ipAddress,
		userAgent,
	)

	return err
}

// Helper functions (duplicated from service for now - can be moved to a shared util package novel_volume)
// Update: Helper functions removed as they are already defined in service.go within the same package.
