// ============================================================================
// Chapter History Repository (Ent Implementation)
// ============================================================================

package novel_chapter

import (
	"context"
	"encoding/json"
	"system/internal/platform/database"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/novelchapterhistory"
)

// chapterHistoryEntRepository implements ChapterHistoryRepository using Ent
type chapterHistoryEntRepository struct {
	client *ent.Client
}

// NewChapterHistoryEntRepository creates a new instance
func NewChapterHistoryEntRepository(client *ent.Client) ChapterHistoryRepository {
	return &chapterHistoryEntRepository{client: client}
}

// GetLatestVersion retrieves the latest version number for a chapter
func (r *chapterHistoryEntRepository) GetLatestVersion(ctx context.Context, chapterID uuid.UUID) (int, error) {
	// Get max version
	history, err := database.GetClientFromContext(ctx, r.client).NovelChapterHistory.Query().
		Where(novelchapterhistory.ChapterIDEQ(chapterID)).
		Order(ent.Desc(novelchapterhistory.FieldVersionNumber)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return history.VersionNumber, nil
}

// LogUpdate logs a chapter update to history
func (r *chapterHistoryEntRepository) LogUpdate(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, oldChapter, newChapter *domain.NovelChapter, changedBy uuid.UUID, requestContext map[string]any) error {
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

	// Build create
	builder := database.GetClientFromContext(ctx, r.client).NovelChapterHistory.Create().
		SetChapterID(chapterID).
		SetNovelID(novelID).
		SetVersionNumber(nextVersion).
		SetAction("updated").
		SetTitle(newChapter.Title).
		SetSlug(newChapter.Slug).
		SetChapterNumber(newChapter.ChapterNumber).
		SetStatus(string(newChapter.Status)).
		SetWordCount(newChapter.WordCount).
		SetCharacterCount(newChapter.CharacterCount).
		SetChangedFields(json.RawMessage(changedFieldsJSON)).
		SetChangeSummary(changeSummary).
		SetContentChanged(contentChanged).
		SetChangedBy(changedBy)

	if volumeID != uuid.Nil {
		builder.SetVolumeID(volumeID)
	}
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

// LogPublish logs a chapter publish action to history
func (r *chapterHistoryEntRepository) LogPublish(ctx context.Context, chapterID, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error {
	// Get next version number
	latestVersion, err := r.GetLatestVersion(ctx, chapterID)
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

	changedFieldsJSON, _ := json.Marshal([]string{"status"})

	builder := database.GetClientFromContext(ctx, r.client).NovelChapterHistory.Create().
		SetChapterID(chapterID).
		SetNovelID(novelID).
		SetVersionNumber(nextVersion).
		SetAction("published").
		SetChangedFields(json.RawMessage(changedFieldsJSON)).
		SetChangeSummary("Published chapter").
		SetContentChanged(false).
		SetChangedBy(changedBy)

	if volumeID != uuid.Nil {
		builder.SetVolumeID(volumeID)
	}
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
