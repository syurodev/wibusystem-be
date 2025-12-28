// ============================================================================
// Media Progress Repository (Ent Implementation)
// ============================================================================
//
// Repository này triển khai MediaProgressRepository interface sử dụng Ent ORM.
// Một số methods sử dụng raw SQL cho complex JOINs và PostgreSQL-specific features.
//
// ============================================================================

package media_progress

import (
	"context"
	"database/sql"
	"encoding/json"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/mediaprogress"
	"system/internal/ent/generated/unitprogress"
)

// entRepository implements domain.MediaProgressRepository using Ent
type entRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewEntRepository creates a new media progress repository instance using Ent.
func NewEntRepository(client *ent.Client, db *sql.DB) domain.MediaProgressRepository {
	return &entRepository{client: client, db: db}
}

// =============================================================================
// MEDIA PROGRESS OPERATIONS
// =============================================================================

// UpsertMediaProgress creates or updates media progress for a user.
func (r *entRepository) UpsertMediaProgress(ctx context.Context, progress *domain.MediaProgress) error {
	// Check existing
	existing, err := database.GetClientFromContext(ctx, r.client).MediaProgress.Query().
		Where(
			mediaprogress.UserIDEQ(progress.UserID),
			mediaprogress.MediaTypeEQ(progress.MediaType),
			mediaprogress.MediaIDEQ(progress.MediaID),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if progress.Position == nil {
		progress.Position = json.RawMessage("{}")
	}

	now := time.Now()

	if existing != nil {
		// Update existing
		updated, err := database.GetClientFromContext(ctx, r.client).MediaProgress.UpdateOne(existing).
			SetCurrentUnitID(progress.CurrentUnitID).
			SetPosition(progress.Position).
			SetTotalUnits(progress.TotalUnits).
			SetLastAccessedAt(now).
			Save(ctx)
		if err != nil {
			return err
		}
		// Update progress with returned values
		progress.ID = updated.ID
		progress.CompletedUnits = updated.CompletedUnits
		progress.ProgressPercentage = updated.ProgressPercentage
		progress.LastAccessedAt = updated.LastAccessedAt
		progress.CreatedAt = updated.CreatedAt
		progress.UpdatedAt = updated.UpdatedAt
	} else {
		// Create new
		created, err := database.GetClientFromContext(ctx, r.client).MediaProgress.Create().
			SetUserID(progress.UserID).
			SetMediaType(progress.MediaType).
			SetMediaID(progress.MediaID).
			SetCurrentUnitID(progress.CurrentUnitID).
			SetPosition(progress.Position).
			SetTotalUnits(progress.TotalUnits).
			SetCompletedUnits(0).
			SetProgressPercentage(0).
			SetLastAccessedAt(now).
			Save(ctx)
		if err != nil {
			return err
		}
		// Update progress with returned values
		progress.ID = created.ID
		progress.CompletedUnits = created.CompletedUnits
		progress.ProgressPercentage = created.ProgressPercentage
		progress.LastAccessedAt = created.LastAccessedAt
		progress.CreatedAt = created.CreatedAt
		progress.UpdatedAt = created.UpdatedAt
	}

	return nil
}

// GetRecentMediaProgress retrieves N most recent media progress entries for a user.
// Note: Uses raw SQL for complex JOINs with novels and chapters tables.
func (r *entRepository) GetRecentMediaProgress(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.MediaProgress, error) {
	query := `
		SELECT 
			mp.id, mp.user_id, mp.media_type, mp.media_id, mp.current_unit_id,
			mp.position, mp.total_units, mp.completed_units, mp.progress_percentage,
			mp.last_accessed_at, mp.created_at, mp.updated_at,
			n.id as novel_id, n.title, n.slug, n.cover_image_url, n.status,
			n.owner_id, u.display_name, u.username, u.avatar_url,
			c.id as chapter_id, c.chapter_number, c.title as chapter_title, c.slug as chapter_slug
		FROM catalog.media_progress mp
		LEFT JOIN catalog.novels n ON mp.media_id = n.id AND mp.media_type = 'novel'
		LEFT JOIN identify.users u ON n.owner_id = u.id
		LEFT JOIN catalog.novel_chapters c ON mp.current_unit_id = c.id
		WHERE mp.user_id = $1
		ORDER BY mp.last_accessed_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.MediaProgress
	for rows.Next() {
		var mp domain.MediaProgress
		var novelID uuid.UUID
		var novelTitle, novelSlug, novelStatus string
		var novelCoverURL *string
		var ownerID uuid.UUID
		var ownerDisplayName, ownerUsername, ownerAvatarURL *string
		var chapterID uuid.UUID
		var chapterNumber int
		var chapterTitle, chapterSlug string

		err := rows.Scan(
			&mp.ID, &mp.UserID, &mp.MediaType, &mp.MediaID, &mp.CurrentUnitID,
			&mp.Position, &mp.TotalUnits, &mp.CompletedUnits, &mp.ProgressPercentage,
			&mp.LastAccessedAt, &mp.CreatedAt, &mp.UpdatedAt,
			&novelID, &novelTitle, &novelSlug, &novelCoverURL, &novelStatus,
			&ownerID, &ownerDisplayName, &ownerUsername, &ownerAvatarURL,
			&chapterID, &chapterNumber, &chapterTitle, &chapterSlug,
		)
		if err != nil {
			return nil, err
		}

		mp.Media = &domain.ProgressMedia{
			ID:               novelID,
			Title:            novelTitle,
			Slug:             novelSlug,
			CoverURL:         novelCoverURL,
			Status:           novelStatus,
			OwnerID:          ownerID,
			OwnerDisplayName: ownerDisplayName,
			OwnerUsername:    ownerUsername,
			OwnerAvatarURL:   ownerAvatarURL,
		}

		mp.CurrentUnit = &domain.ProgressUnit{
			ID:     chapterID,
			Number: chapterNumber,
			Title:  chapterTitle,
			Slug:   chapterSlug,
		}

		results = append(results, &mp)
	}

	return results, rows.Err()
}

// GetMediaProgressByUserAndMedia retrieves progress for a specific user and media.
func (r *entRepository) GetMediaProgressByUserAndMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) (*domain.MediaProgress, error) {
	mp, err := database.GetClientFromContext(ctx, r.client).MediaProgress.Query().
		Where(
			mediaprogress.UserIDEQ(userID),
			mediaprogress.MediaTypeEQ(mediaType),
			mediaprogress.MediaIDEQ(mediaID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entMediaProgressToDomain(mp), nil
}

// GetMediaProgressByUserID retrieves paginated progress for a user.
func (r *entRepository) GetMediaProgressByUserID(ctx context.Context, userID uuid.UUID, filter domain.MediaProgressFilter) ([]*domain.MediaProgress, int64, error) {
	// Build query
	query := database.GetClientFromContext(ctx, r.client).MediaProgress.Query().
		Where(mediaprogress.UserIDEQ(userID)).
		Order(ent.Desc(mediaprogress.FieldLastAccessedAt))

	// Get total count
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	entries, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.MediaProgress, len(entries))
	for i, mp := range entries {
		results[i] = entMediaProgressToDomain(mp)
	}

	return results, int64(total), nil
}

// DeleteMediaProgress deletes a media progress entry and related unit progress.
func (r *entRepository) DeleteMediaProgress(ctx context.Context, id uuid.UUID) error {
	// Get the media progress first
	mp, err := database.GetClientFromContext(ctx, r.client).MediaProgress.Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete unit progress first
	_, err = database.GetClientFromContext(ctx, r.client).UnitProgress.Delete().
		Where(
			unitprogress.UserIDEQ(mp.UserID),
			unitprogress.MediaTypeEQ(mp.MediaType),
			unitprogress.MediaIDEQ(mp.MediaID),
		).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Delete media progress
	return database.GetClientFromContext(ctx, r.client).MediaProgress.DeleteOneID(id).Exec(ctx)
}

// DeleteAllMediaProgressByUser deletes all progress for a user.
func (r *entRepository) DeleteAllMediaProgressByUser(ctx context.Context, userID uuid.UUID) error {
	// Delete all unit progress first
	_, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Delete().
		Where(unitprogress.UserIDEQ(userID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Delete all media progress
	_, err = database.GetClientFromContext(ctx, r.client).MediaProgress.Delete().
		Where(mediaprogress.UserIDEQ(userID)).
		Exec(ctx)
	return err
}

// =============================================================================
// UNIT PROGRESS OPERATIONS
// =============================================================================

// UpsertUnitProgress creates or updates unit (chapter/episode) progress.
func (r *entRepository) UpsertUnitProgress(ctx context.Context, progress *domain.UnitProgress) error {
	// Check existing
	existing, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Query().
		Where(
			unitprogress.UserIDEQ(progress.UserID),
			unitprogress.UnitIDEQ(progress.UnitID),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if progress.Position == nil {
		progress.Position = json.RawMessage("{}")
	}

	now := time.Now()

	if existing != nil {
		// Update existing
		updated, err := database.GetClientFromContext(ctx, r.client).UnitProgress.UpdateOne(existing).
			SetStatus(unitprogress.Status(progress.Status)).
			SetPosition(progress.Position).
			SetLastAccessedAt(now).
			Save(ctx)
		if err != nil {
			return err
		}
		// Update progress with returned values
		progress.ID = updated.ID
		progress.StartedAt = updated.StartedAt
		progress.CompletedAt = updated.CompletedAt
		progress.LastAccessedAt = updated.LastAccessedAt
	} else {
		// Create new
		created, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Create().
			SetUserID(progress.UserID).
			SetMediaType(progress.MediaType).
			SetMediaID(progress.MediaID).
			SetUnitID(progress.UnitID).
			SetStatus(unitprogress.Status(progress.Status)).
			SetPosition(progress.Position).
			SetStartedAt(now).
			SetLastAccessedAt(now).
			Save(ctx)
		if err != nil {
			return err
		}
		// Update progress with returned values
		progress.ID = created.ID
		progress.StartedAt = created.StartedAt
		progress.CompletedAt = created.CompletedAt
		progress.LastAccessedAt = created.LastAccessedAt
	}

	return nil
}

// GetUnitProgressByMedia retrieves all unit progress for a specific media.
func (r *entRepository) GetUnitProgressByMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) ([]*domain.UnitProgress, error) {
	entries, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Query().
		Where(
			unitprogress.UserIDEQ(userID),
			unitprogress.MediaTypeEQ(mediaType),
			unitprogress.MediaIDEQ(mediaID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.UnitProgress, len(entries))
	for i, up := range entries {
		results[i] = entUnitProgressToDomain(up)
	}
	return results, nil
}

// GetUnitProgress retrieves progress for a specific unit.
func (r *entRepository) GetUnitProgress(ctx context.Context, userID uuid.UUID, unitID uuid.UUID) (*domain.UnitProgress, error) {
	up, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Query().
		Where(
			unitprogress.UserIDEQ(userID),
			unitprogress.UnitIDEQ(unitID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entUnitProgressToDomain(up), nil
}

// MarkUnitComplete marks a unit as completed.
func (r *entRepository) MarkUnitComplete(ctx context.Context, userID uuid.UUID, unitID uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Update().
		Where(
			unitprogress.UserIDEQ(userID),
			unitprogress.UnitIDEQ(unitID),
		).
		SetStatus(unitprogress.StatusCompleted).
		SetCompletedAt(now).
		Save(ctx)
	return err
}

// UpdateCompletedUnitsCount recalculates and updates completed_units in media_progress.
func (r *entRepository) UpdateCompletedUnitsCount(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) error {
	// Count completed units
	count, err := database.GetClientFromContext(ctx, r.client).UnitProgress.Query().
		Where(
			unitprogress.UserIDEQ(userID),
			unitprogress.MediaTypeEQ(mediaType),
			unitprogress.MediaIDEQ(mediaID),
			unitprogress.StatusEQ(unitprogress.StatusCompleted),
		).
		Count(ctx)
	if err != nil {
		return err
	}

	// Update media progress
	mp, err := database.GetClientFromContext(ctx, r.client).MediaProgress.Query().
		Where(
			mediaprogress.UserIDEQ(userID),
			mediaprogress.MediaTypeEQ(mediaType),
			mediaprogress.MediaIDEQ(mediaID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No media progress to update
		}
		return err
	}

	// Calculate percentage
	var percentage float64
	if mp.TotalUnits > 0 {
		percentage = float64(count) / float64(mp.TotalUnits) * 100
	}

	_, err = database.GetClientFromContext(ctx, r.client).MediaProgress.UpdateOne(mp).
		SetCompletedUnits(count).
		SetProgressPercentage(percentage).
		Save(ctx)
	return err
}

// =============================================================================
// Helper Functions
// =============================================================================

func entMediaProgressToDomain(mp *ent.MediaProgress) *domain.MediaProgress {
	return &domain.MediaProgress{
		ID:                 mp.ID,
		UserID:             mp.UserID,
		MediaType:          mp.MediaType,
		MediaID:            mp.MediaID,
		CurrentUnitID:      mp.CurrentUnitID,
		Position:           mp.Position,
		TotalUnits:         mp.TotalUnits,
		CompletedUnits:     mp.CompletedUnits,
		ProgressPercentage: mp.ProgressPercentage,
		LastAccessedAt:     mp.LastAccessedAt,
		CreatedAt:          mp.CreatedAt,
		UpdatedAt:          mp.UpdatedAt,
	}
}

func entUnitProgressToDomain(up *ent.UnitProgress) *domain.UnitProgress {
	return &domain.UnitProgress{
		ID:             up.ID,
		UserID:         up.UserID,
		MediaType:      up.MediaType,
		MediaID:        up.MediaID,
		UnitID:         up.UnitID,
		Status:         domain.UnitProgressStatus(up.Status),
		Position:       up.Position,
		StartedAt:      up.StartedAt,
		CompletedAt:    up.CompletedAt,
		LastAccessedAt: up.LastAccessedAt,
	}
}
