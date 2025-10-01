package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// VolumeRepository defines CRUD and listing operations for novel volumes
// This repository handles all database operations related to novel volumes,
// following the repository pattern established in the catalog service.
type VolumeRepository interface {
	// CreateVolume inserts a new volume for a specific novel
	// Returns the created volume with generated ID and timestamps
	CreateVolume(ctx context.Context, novelID uuid.UUID, req d.CreateVolumeRequest) (*m.NovelVolume, error)

	// GetVolumeByID retrieves a single volume by its ID
	// Returns error if volume not found or is deleted
	GetVolumeByID(ctx context.Context, id uuid.UUID) (*m.NovelVolume, error)

	// ListVolumesByNovelID retrieves all volumes for a specific novel with pagination
	// Returns paginated list of volumes ordered by volume_number
	ListVolumesByNovelID(ctx context.Context, novelID uuid.UUID, req d.ListVolumesRequest) (*d.PaginatedVolumesResponse, error)

	// UpdateVolume modifies an existing volume
	// Returns the updated volume with new timestamps
	UpdateVolume(ctx context.Context, id uuid.UUID, req d.UpdateVolumeRequest) (*m.NovelVolume, error)

	// DeleteVolume performs soft delete on a volume
	// Sets is_deleted=true and records deletion timestamp
	DeleteVolume(ctx context.Context, id uuid.UUID) error

	// CheckVolumePurchases checks if any users have purchased content from this volume
	// Used to prevent deletion of volumes that users have paid for
	CheckVolumePurchases(ctx context.Context, volumeID uuid.UUID) (bool, error)
}

// volumeRepository implements VolumeRepository interface
type volumeRepository struct {
	pool *pgxpool.Pool
}

// NewVolumeRepository creates a new volume repository instance
// Takes a postgres connection pool for database operations
func NewVolumeRepository(pool *pgxpool.Pool) VolumeRepository {
	return &volumeRepository{pool: pool}
}

// CreateVolume inserts a new volume for a specific novel
// This method validates the novel exists before creating the volume
func (r *volumeRepository) CreateVolume(ctx context.Context, novelID uuid.UUID, req d.CreateVolumeRequest) (*m.NovelVolume, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Verify novel exists and is not deleted
	var novelExists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM novel WHERE id = $1 AND is_deleted = FALSE)", novelID).Scan(&novelExists)
	if err != nil {
		return nil, fmt.Errorf("failed to verify novel existence: %w", err)
	}
	if !novelExists {
		return nil, fmt.Errorf("novel not found or already deleted")
	}

	// Check if volume number already exists for this novel
	var volumeExists bool
	err = tx.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM novel_volume WHERE novel_id = $1 AND volume_number = $2 AND is_deleted = FALSE)",
		novelID, req.VolumeNumber,
	).Scan(&volumeExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check volume number: %w", err)
	}
	if volumeExists {
		return nil, fmt.Errorf("volume number %d already exists for this novel", req.VolumeNumber)
	}

	// Insert volume record
	volumeID := uuid.New()
	var volume m.NovelVolume

	query := `
		INSERT INTO novel_volume (
			id, novel_id, volume_number, volume_title, description, cover_image,
			is_available, price_coins, chapter_count,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING
			id, novel_id, volume_number, volume_title, description, cover_image,
			published_at, is_deleted, deleted_at, is_available,
			price_coins, rental_price_coins, rental_duration_days,
			page_count, word_count, chapter_count, estimated_reading_time,
			created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		volumeID, novelID, req.VolumeNumber, req.Title, req.Description, req.CoverImage,
		req.IsPublic, req.PriceCoins,
	).Scan(
		&volume.ID, &volume.NovelID, &volume.VolumeNumber, &volume.VolumeTitle, &volume.Description, &volume.CoverImage,
		&volume.PublishedAt, &volume.IsDeleted, &volume.DeletedAt, &volume.IsAvailable,
		&volume.PriceCoins, &volume.RentalPriceCoins, &volume.RentalDurationDays,
		&volume.PageCount, &volume.WordCount, &volume.ChapterCount, &volume.EstimatedReadingTime,
		&volume.CreatedAt, &volume.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &volume, nil
}

// GetVolumeByID retrieves a single volume by its ID
// Returns error if volume is not found or has been soft-deleted
func (r *volumeRepository) GetVolumeByID(ctx context.Context, id uuid.UUID) (*m.NovelVolume, error) {
	var volume m.NovelVolume

	query := `
		SELECT
			id, novel_id, volume_number, volume_title, description, cover_image,
			created_by_user_id, updated_by_user_id,
			published_at, is_deleted, deleted_at, is_available,
			price_coins, rental_price_coins, rental_duration_days,
			page_count, word_count, chapter_count, estimated_reading_time,
			created_at, updated_at
		FROM novel_volume
		WHERE id = $1 AND is_deleted = FALSE
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&volume.ID, &volume.NovelID, &volume.VolumeNumber, &volume.VolumeTitle, &volume.Description, &volume.CoverImage,
		&volume.CreatedByUserID, &volume.UpdatedByUserID,
		&volume.PublishedAt, &volume.IsDeleted, &volume.DeletedAt, &volume.IsAvailable,
		&volume.PriceCoins, &volume.RentalPriceCoins, &volume.RentalDurationDays,
		&volume.PageCount, &volume.WordCount, &volume.ChapterCount, &volume.EstimatedReadingTime,
		&volume.CreatedAt, &volume.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("volume not found")
		}
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	return &volume, nil
}

// ListVolumesByNovelID retrieves all volumes for a specific novel with pagination
// Results are ordered by volume_number ascending
func (r *volumeRepository) ListVolumesByNovelID(ctx context.Context, novelID uuid.UUID, req d.ListVolumesRequest) (*d.PaginatedVolumesResponse, error) {
	// Set pagination defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Build query to fetch volumes
	query := `
		SELECT
			id, novel_id, volume_number, volume_title, description, cover_image,
			published_at, is_available, price_coins, chapter_count,
			created_at, updated_at
		FROM novel_volume
		WHERE novel_id = $1 AND is_deleted = FALSE
		ORDER BY volume_number ASC
		LIMIT $2 OFFSET $3
	`

	offset := (req.Page - 1) * req.Limit
	rows, err := r.pool.Query(ctx, query, novelID, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query volumes: %w", err)
	}
	defer rows.Close()

	var volumes []d.VolumeResponse
	for rows.Next() {
		var volume d.VolumeResponse
		var id, novelIDStr string
		var chapterCount int
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &novelIDStr, &volume.VolumeNumber, &volume.Title, &volume.Description, &volume.CoverImage,
			&volume.PublishedAt, &volume.IsPublic, &volume.PriceCoins, &chapterCount,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan volume: %w", err)
		}

		volume.ID = id
		volume.NovelID = novelIDStr
		volume.ChapterCount = chapterCount
		volume.CreatedAt = createdAt
		volume.UpdatedAt = updatedAt

		// Initialize Chapters as empty array if includeChapters is false
		if !req.IncludeChapters {
			volume.Chapters = nil
		} else {
			// TODO: Implement chapter loading when chapter repository is ready
			volume.Chapters = []interface{}{}
		}

		volumes = append(volumes, volume)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate volumes: %w", rows.Err())
	}

	// Get total count for pagination
	var total int64
	countQuery := `SELECT COUNT(*) FROM novel_volume WHERE novel_id = $1 AND is_deleted = FALSE`
	err = r.pool.QueryRow(ctx, countQuery, novelID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrevious := req.Page > 1

	return &d.PaginatedVolumesResponse{
		Volumes: volumes,
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

// UpdateVolume modifies an existing volume
// Only updates fields that are provided in the request (non-nil values)
func (r *volumeRepository) UpdateVolume(ctx context.Context, id uuid.UUID, req d.UpdateVolumeRequest) (*m.NovelVolume, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	// Always update updated_at
	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Update fields if provided
	if req.Title != nil {
		updateFields = append(updateFields, fmt.Sprintf("volume_title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Description != nil {
		updateFields = append(updateFields, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.CoverImage != nil {
		updateFields = append(updateFields, fmt.Sprintf("cover_image = $%d", argIndex))
		args = append(args, *req.CoverImage)
		argIndex++
	}

	if req.IsPublic != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_available = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if req.PriceCoins != nil {
		updateFields = append(updateFields, fmt.Sprintf("price_coins = $%d", argIndex))
		args = append(args, *req.PriceCoins)
		argIndex++
	}

	if len(updateFields) == 1 { // Only updated_at field
		return nil, fmt.Errorf("no fields to update")
	}

	// Build and execute update query
	query := fmt.Sprintf(`
		UPDATE novel_volume
		SET %s
		WHERE id = $%d AND is_deleted = FALSE
		RETURNING
			id, novel_id, volume_number, volume_title, description, cover_image,
			published_at, is_deleted, deleted_at, is_available,
			price_coins, rental_price_coins, rental_duration_days,
			page_count, word_count, chapter_count, estimated_reading_time,
			created_at, updated_at
	`, strings.Join(updateFields, ", "), argIndex)

	args = append(args, id)

	var volume m.NovelVolume
	err = tx.QueryRow(ctx, query, args...).Scan(
		&volume.ID, &volume.NovelID, &volume.VolumeNumber, &volume.VolumeTitle, &volume.Description, &volume.CoverImage,
		&volume.PublishedAt, &volume.IsDeleted, &volume.DeletedAt, &volume.IsAvailable,
		&volume.PriceCoins, &volume.RentalPriceCoins, &volume.RentalDurationDays,
		&volume.PageCount, &volume.WordCount, &volume.ChapterCount, &volume.EstimatedReadingTime,
		&volume.CreatedAt, &volume.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("volume not found or already deleted")
		}
		return nil, fmt.Errorf("failed to update volume: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &volume, nil
}

// CheckVolumePurchases checks if any users have purchased content from this volume
// Returns true if there are any purchase or rental records for this volume or its chapters
func (r *volumeRepository) CheckVolumePurchases(ctx context.Context, volumeID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_content_purchases ucp
			WHERE (
				-- Check volume purchases
				(ucp.item_type = 'NOVEL_VOLUME' AND ucp.item_id = $1)
				OR
				-- Check chapter purchases within this volume
				(ucp.item_type = 'NOVEL_CHAPTER' AND ucp.item_id IN (
					SELECT id FROM novel_chapter WHERE volume_id = $1 AND is_deleted = FALSE
				))
			)
			UNION
			-- Check rental purchases
			SELECT 1 FROM user_content_rentals ucr
			WHERE ucr.item_type = 'NOVEL_VOLUME' AND ucr.item_id = $1
		)
	`

	var hasPurchases bool
	err := r.pool.QueryRow(ctx, query, volumeID).Scan(&hasPurchases)
	if err != nil {
		return false, fmt.Errorf("failed to check volume purchases: %w", err)
	}

	return hasPurchases, nil
}

// DeleteVolume performs soft delete on a volume
// This operation also soft-deletes all chapters within the volume
// Fails if any users have purchased content from this volume
func (r *volumeRepository) DeleteVolume(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if volume exists and is not already deleted
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM novel_volume WHERE id = $1 AND is_deleted = FALSE)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check volume existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("volume not found or already deleted")
	}

	// Check if any users have purchased content from this volume
	hasPurchases, err := r.CheckVolumePurchases(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check purchases: %w", err)
	}
	if hasPurchases {
		return fmt.Errorf("cannot delete volume: users have purchased content from this volume")
	}

	// Perform soft delete on the volume
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE novel_volume
		SET is_deleted = TRUE, deleted_at = $2, updated_at = $2
		WHERE id = $1
	`, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	// Also soft delete related chapters
	_, err = tx.Exec(ctx, `
		UPDATE novel_chapter
		SET is_deleted = TRUE, deleted_at = $2
		WHERE volume_id = $1 AND is_deleted = FALSE
	`, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete volume chapters: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}