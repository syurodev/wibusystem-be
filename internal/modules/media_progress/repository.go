/*
Media Progress Repository - Data Access Layer
==============================================

ARCHITECTURE:
─────────────
Repository này triển khai MediaProgressRepository interface từ domain layer.
Sử dụng pgx để query PostgreSQL với embedded SQL queries.

FLOW OVERVIEW:
──────────────
1. Handler nhận HTTP request
2. Handler gọi Service method
3. Service gọi Repository method (file này)
4. Repository thực thi SQL query và map kết quả về domain entity

EMBEDDED QUERIES:
─────────────────
SQL queries được load từ thư mục queries/ sử dụng go:embed directive.
Điều này cho phép:
- Dễ dàng quản lý và review SQL queries
- Syntax highlighting trong IDE
- Tách biệt SQL khỏi Go code
*/

package media_progress

import (
	"context"
	_ "embed"
	"encoding/json"
	"system/internal/domain"
	db "system/internal/platform/database"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =============================================================================
// EMBEDDED SQL QUERIES
// =============================================================================

//go:embed queries/upsert_media_progress.sql
var upsertMediaProgressQuery string

//go:embed queries/upsert_unit_progress.sql
var upsertUnitProgressQuery string

//go:embed queries/get_recent.sql
var getRecentQuery string

//go:embed queries/get_units_by_media.sql
var getUnitsByMediaQuery string

//go:embed queries/mark_unit_complete.sql
var markUnitCompleteQuery string

//go:embed queries/update_completed_count.sql
var updateCompletedCountQuery string

// =============================================================================
// REPOSITORY IMPLEMENTATION
// =============================================================================

// repository implements domain.MediaProgressRepository
type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new media progress repository instance.
//
// Parameters:
//   - pool: PostgreSQL connection pool
//
// Returns:
//   - domain.MediaProgressRepository: Repository interface implementation
func NewRepository(pool *pgxpool.Pool) domain.MediaProgressRepository {
	return &repository{pool: pool}
}

// =============================================================================
// MEDIA PROGRESS OPERATIONS
// =============================================================================

// UpsertMediaProgress creates or updates media progress for a user.
//
// FLOW:
// 1. Nhận MediaProgress từ Service
// 2. Execute UPSERT query (INSERT ... ON CONFLICT DO UPDATE)
// 3. Return updated MediaProgress
//
// Use case: User mở chapter để đọc → update current position
func (r *repository) UpsertMediaProgress(ctx context.Context, progress *domain.MediaProgress) error {
	conn := db.GetDB(ctx, r.pool)

	// Ensure position is valid JSON
	if progress.Position == nil {
		progress.Position = json.RawMessage("{}")
	}

	row := conn.QueryRow(ctx, upsertMediaProgressQuery,
		progress.UserID,         // $1
		progress.MediaType,      // $2
		progress.MediaID,        // $3
		progress.CurrentUnitID,  // $4
		progress.Position,       // $5
		progress.TotalUnits,     // $6
	)

	// Scan returned row to update progress with DB-generated values
	err := row.Scan(
		&progress.ID,
		&progress.UserID,
		&progress.MediaType,
		&progress.MediaID,
		&progress.CurrentUnitID,
		&progress.Position,
		&progress.TotalUnits,
		&progress.CompletedUnits,
		&progress.ProgressPercentage,
		&progress.LastAccessedAt,
		&progress.CreatedAt,
		&progress.UpdatedAt,
	)

	return err
}

// GetRecentMediaProgress retrieves N most recent media progress entries for a user.
//
// FLOW:
// 1. Query media_progress với ORDER BY last_accessed_at DESC
// 2. JOIN với novels + chapters để lấy thông tin hiển thị
// 3. Map kết quả về []*MediaProgress với Media và CurrentUnit populated
//
// Use case: Hiển thị "Continue Reading" section trên homepage
func (r *repository) GetRecentMediaProgress(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.MediaProgress, error) {
	rows, err := r.pool.Query(ctx, getRecentQuery, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.MediaProgress

	for rows.Next() {
		var mp domain.MediaProgress

		// Media (Novel) info
		var novelID uuid.UUID
		var novelTitle, novelSlug, novelStatus string
		var novelCoverURL *string
		var ownerID uuid.UUID
		var ownerDisplayName, ownerUsername *string
		var ownerAvatarURL *string

		// Current Unit (Chapter) info
		var chapterID uuid.UUID
		var chapterNumber int
		var chapterTitle, chapterSlug string

		err := rows.Scan(
			// MediaProgress fields
			&mp.ID,
			&mp.UserID,
			&mp.MediaType,
			&mp.MediaID,
			&mp.CurrentUnitID,
			&mp.Position,
			&mp.TotalUnits,
			&mp.CompletedUnits,
			&mp.ProgressPercentage,
			&mp.LastAccessedAt,
			&mp.CreatedAt,
			&mp.UpdatedAt,

			// Novel (Media) fields
			&novelID,
			&novelTitle,
			&novelSlug,
			&novelCoverURL,
			&novelStatus,
			&ownerID,
			&ownerDisplayName,
			&ownerUsername,
			&ownerAvatarURL,

			// Chapter (CurrentUnit) fields
			&chapterID,
			&chapterNumber,
			&chapterTitle,
			&chapterSlug,
		)
		if err != nil {
			return nil, err
		}

		// Populate Media (Novel) info
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

		// Populate CurrentUnit (Chapter) info
		mp.CurrentUnit = &domain.ProgressUnit{
			ID:     chapterID,
			Number: chapterNumber,
			Title:  chapterTitle,
			Slug:   chapterSlug,
		}

		results = append(results, &mp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetMediaProgressByUserAndMedia retrieves progress for a specific user and media.
//
// Use case: User vào trang detail của novel để xem tiến độ hiện tại
func (r *repository) GetMediaProgressByUserAndMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) (*domain.MediaProgress, error) {
	query := `
		SELECT id, user_id, media_type, media_id, current_unit_id, position,
		       total_units, completed_units, progress_percentage,
		       last_accessed_at, created_at, updated_at
		FROM catalog.media_progress
		WHERE user_id = $1 AND media_type = $2 AND media_id = $3
	`

	rows, err := r.pool.Query(ctx, query, userID, mediaType, mediaID)
	if err != nil {
		return nil, err
	}

	mp, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.MediaProgress])
	if err != nil {
		return nil, err
	}

	return &mp, nil
}

// GetMediaProgressByUserID retrieves paginated progress for a user.
//
// Use case: Hiển thị trang History với pagination và filters
func (r *repository) GetMediaProgressByUserID(ctx context.Context, userID uuid.UUID, filter domain.MediaProgressFilter) ([]*domain.MediaProgress, int64, error) {
	// TODO: Implement with dynamic query building for filters
	// For now, return basic implementation
	query := `
		SELECT id, user_id, media_type, media_id, current_unit_id, position,
		       total_units, completed_units, progress_percentage,
		       last_accessed_at, created_at, updated_at
		FROM catalog.media_progress
		WHERE user_id = $1
		ORDER BY last_accessed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.MediaProgress])
	if err != nil {
		return nil, 0, err
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM catalog.media_progress WHERE user_id = $1`
	var total int64
	err = r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// DeleteMediaProgress deletes a media progress entry and related unit progress.
//
// FLOW:
// 1. Get media_type and media_id from the progress entry
// 2. Delete all unit_progress for that media
// 3. Delete the media_progress entry
//
// Use case: User xóa lịch sử của 1 novel
func (r *repository) DeleteMediaProgress(ctx context.Context, id uuid.UUID) error {
	// First get the media info
	var mediaType string
	var mediaID, userID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT user_id, media_type, media_id FROM catalog.media_progress WHERE id = $1
	`, id).Scan(&userID, &mediaType, &mediaID)
	if err != nil {
		return err
	}

	// Delete unit progress first
	_, err = r.pool.Exec(ctx, `
		DELETE FROM catalog.unit_progress 
		WHERE user_id = $1 AND media_type = $2 AND media_id = $3
	`, userID, mediaType, mediaID)
	if err != nil {
		return err
	}

	// Delete media progress
	_, err = r.pool.Exec(ctx, `
		DELETE FROM catalog.media_progress WHERE id = $1
	`, id)

	return err
}

// DeleteAllMediaProgressByUser deletes all progress for a user.
//
// Use case: User chọn "Clear All History"
func (r *repository) DeleteAllMediaProgressByUser(ctx context.Context, userID uuid.UUID) error {
	// Delete all unit progress first
	_, err := r.pool.Exec(ctx, `
		DELETE FROM catalog.unit_progress WHERE user_id = $1
	`, userID)
	if err != nil {
		return err
	}

	// Delete all media progress
	_, err = r.pool.Exec(ctx, `
		DELETE FROM catalog.media_progress WHERE user_id = $1
	`, userID)

	return err
}

// =============================================================================
// UNIT PROGRESS OPERATIONS
// =============================================================================

// UpsertUnitProgress creates or updates unit (chapter/episode) progress.
//
// FLOW:
// 1. Nhận UnitProgress từ Service
// 2. Execute UPSERT query
// 3. Return updated UnitProgress
//
// Use case: User mở chapter → track rằng user đã bắt đầu đọc chapter này
func (r *repository) UpsertUnitProgress(ctx context.Context, progress *domain.UnitProgress) error {
	conn := db.GetDB(ctx, r.pool)

	if progress.Position == nil {
		progress.Position = json.RawMessage("{}")
	}

	row := conn.QueryRow(ctx, upsertUnitProgressQuery,
		progress.UserID,    // $1
		progress.MediaType, // $2
		progress.MediaID,   // $3
		progress.UnitID,    // $4
		progress.Status,    // $5
		progress.Position,  // $6
	)

	err := row.Scan(
		&progress.ID,
		&progress.UserID,
		&progress.MediaType,
		&progress.MediaID,
		&progress.UnitID,
		&progress.Status,
		&progress.Position,
		&progress.StartedAt,
		&progress.CompletedAt,     // Column 9: completed_at (nullable)
		&progress.LastAccessedAt,  // Column 10: last_accessed_at
	)

	return err
}

// GetUnitProgressByMedia retrieves all unit progress for a specific media.
//
// FLOW:
// 1. Query unit_progress WHERE media_type = ? AND media_id = ?
// 2. Return list of UnitProgress
//
// Use case: Hiển thị chapter list với icon "đã đọc"
func (r *repository) GetUnitProgressByMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) ([]*domain.UnitProgress, error) {
	rows, err := r.pool.Query(ctx, getUnitsByMediaQuery, userID, mediaType, mediaID)
	if err != nil {
		return nil, err
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.UnitProgress])
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUnitProgress retrieves progress for a specific unit.
func (r *repository) GetUnitProgress(ctx context.Context, userID uuid.UUID, unitID uuid.UUID) (*domain.UnitProgress, error) {
	query := `
		SELECT id, user_id, media_type, media_id, unit_id, status, position,
		       started_at, completed_at, last_accessed_at
		FROM catalog.unit_progress
		WHERE user_id = $1 AND unit_id = $2
	`

	rows, err := r.pool.Query(ctx, query, userID, unitID)
	if err != nil {
		return nil, err
	}

	up, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.UnitProgress])
	if err != nil {
		return nil, err
	}

	return &up, nil
}

// MarkUnitComplete marks a unit as completed.
//
// FLOW:
// 1. UPDATE unit_progress SET status = 'completed', completed_at = NOW()
// 2. Sau đó Service sẽ gọi UpdateCompletedUnitsCount()
//
// Use case: User scroll tới cuối chapter → đánh dấu chapter là đã đọc xong
func (r *repository) MarkUnitComplete(ctx context.Context, userID uuid.UUID, unitID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, markUnitCompleteQuery, userID, unitID)
	return err
}

// UpdateCompletedUnitsCount recalculates and updates completed_units in media_progress.
//
// FLOW:
// 1. COUNT units có status = 'completed' từ unit_progress
// 2. UPDATE media_progress với count mới
// 3. Tính lại progress_percentage
//
// Use case: Sau khi mark unit complete, sync lại số liệu
func (r *repository) UpdateCompletedUnitsCount(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, updateCompletedCountQuery, userID, mediaType, mediaID)
	return err
}
