package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"system/internal/domain"
	"system/internal/pkg/service"
)

// volumeHistoryRepository implements VolumeHistoryRepository
type volumeHistoryRepository struct {
	pool *pgxpool.Pool
}

// NewVolumeHistoryRepository creates a new instance of volumeHistoryRepository
func NewVolumeHistoryRepository(pool *pgxpool.Pool) service.VolumeHistoryRepository {
	return &volumeHistoryRepository{pool: pool}
}

// GetLatestVersion retrieves the latest version number for a volume
func (r *volumeHistoryRepository) GetLatestVersion(ctx context.Context, volumeID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(version_number), 0) as version
		FROM catalog.volume_history
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
		INSERT INTO catalog.volume_history (
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
		INSERT INTO catalog.volume_history (
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
		INSERT INTO catalog.volume_history (
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

// Helper functions (duplicated from service for now - can be moved to a shared util package later)

func detectVolumeChangedFields(old, new *domain.Volume) []string {
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

func generateVolumeChangeSummary(changedFields []string) string {
	if len(changedFields) == 0 {
		return "No changes"
	}

	fieldDescriptions := map[string]string{
		"volume_number":  "volume number",
		"title":          "title",
		"slug":           "slug",
		"description":    "description",
		"cover_image_url": "cover image",
		"display_order":  "display order",
		"is_published":   "publication status",
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

	return fmt.Sprintf("Updated %s", fmt.Sprint(descriptions))
}

func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
