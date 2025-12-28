// ============================================================================
// Volume History Repository (Ent Implementation)
// ============================================================================

package novel_volume

import (
	"context"
	"encoding/json"
	"system/internal/platform/database"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/novelvolumehistory"
)

// volumeHistoryEntRepository implements VolumeHistoryRepository using Ent
type volumeHistoryEntRepository struct {
	client *ent.Client
}

// NewVolumeHistoryEntRepository creates a new instance
func NewVolumeHistoryEntRepository(client *ent.Client) VolumeHistoryRepository {
	return &volumeHistoryEntRepository{client: client}
}

// GetLatestVersion retrieves the latest version number for a volume
func (r *volumeHistoryEntRepository) GetLatestVersion(ctx context.Context, volumeID uuid.UUID) (int, error) {
	// Get max version
	history, err := database.GetClientFromContext(ctx, r.client).NovelVolumeHistory.Query().
		Where(novelvolumehistory.VolumeIDEQ(volumeID)).
		Order(ent.Desc(novelvolumehistory.FieldVersionNumber)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return history.VersionNumber, nil
}

// LogUpdate logs a volume update to history
func (r *volumeHistoryEntRepository) LogUpdate(ctx context.Context, volumeID, novelID uuid.UUID, oldVolume, newVolume *domain.NovelVolume, changedBy uuid.UUID, requestContext map[string]any) error {
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

	// Convert changed fields to JSON
	changedFieldsJSON, err := json.Marshal(changedFields)
	if err != nil {
		return err
	}

	// Extract request context
	var requestID, ipAddress, userAgent *string
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

	builder := database.GetClientFromContext(ctx, r.client).NovelVolumeHistory.Create().
		SetVolumeID(volumeID).
		SetNovelID(novelID).
		SetVersionNumber(nextVersion).
		SetAction("updated").
		SetTitle(newVolume.Title).
		SetSlug(newVolume.Slug).
		SetVolumeNumber(newVolume.VolumeNumber).
		SetIsPublished(newVolume.IsPublished).
		SetChapterCount(newVolume.ChapterCount).
		SetWordCount(int(newVolume.WordCount)).
		SetChangedFields(json.RawMessage(changedFieldsJSON)).
		SetChangeSummary(changeSummary).
		SetChangedBy(changedBy)

	if requestID != nil {
		builder.SetRequestID(*requestID)
	}
	if ipAddress != nil {
		builder.SetIPAddress(*ipAddress)
	}
	if userAgent != nil {
		builder.SetUserAgent(*userAgent)
	}

	_, err = builder.Save(ctx)
	return err
}

// LogPublish logs a volume publish action to history
func (r *volumeHistoryEntRepository) LogPublish(ctx context.Context, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	return r.logAction(ctx, volumeID, novelID, "published", "Published volume", changedBy, requestContext)
}

// LogUnpublish logs a volume unpublish action to history
func (r *volumeHistoryEntRepository) LogUnpublish(ctx context.Context, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	return r.logAction(ctx, volumeID, novelID, "unpublished", "Unpublished volume", changedBy, requestContext)
}

// logAction is a helper to log publish/unpublish actions
func (r *volumeHistoryEntRepository) logAction(ctx context.Context, volumeID, novelID uuid.UUID, action, summary string, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, volumeID)
	if err != nil {
		return err
	}
	nextVersion := latestVersion + 1

	// Extract request context
	var requestID, ipAddress, userAgent *string
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

	changedFieldsJSON, _ := json.Marshal([]string{"is_published"})

	builder := database.GetClientFromContext(ctx, r.client).NovelVolumeHistory.Create().
		SetVolumeID(volumeID).
		SetNovelID(novelID).
		SetVersionNumber(nextVersion).
		SetAction(action).
		SetChangedFields(json.RawMessage(changedFieldsJSON)).
		SetChangeSummary(summary).
		SetChangedBy(changedBy)

	if requestID != nil {
		builder.SetRequestID(*requestID)
	}
	if ipAddress != nil {
		builder.SetIPAddress(*ipAddress)
	}
	if userAgent != nil {
		builder.SetUserAgent(*userAgent)
	}

	_, err = builder.Save(ctx)
	return err
}
