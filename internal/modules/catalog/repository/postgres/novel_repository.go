package postgres

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NovelRepositoryPG is a PostgreSQL implementation of the NovelRepository interface.
type NovelRepositoryPG struct {
	DB *pgxpool.Pool
}

// NewNovelRepositoryPG creates a new instance of NovelRepositoryPG.
func NewNovelRepositoryPG(db *pgxpool.Pool) *NovelRepositoryPG {
	return &NovelRepositoryPG{DB: db}
}

// Create inserts a new novel record into the database.
func (r *NovelRepositoryPG) Create(ctx context.Context, novel *domain.Novel) error {
	query := `
		INSERT INTO novels (
			id, status, title, cover_image, summary, ownership_type, primary_owner_id,
			original_creator_id, access_level, last_modified_by_user_id, ownership_transferred_at,
			published_at, original_language, source_url, isbn, age_rating, content_warnings,
			mature_content, is_public, is_featured, is_completed, is_deleted, deleted_at,
			deleted_by_user_id, slug, tags, keywords, meta_description, view_count, like_count,
			bookmark_count, comment_count, rating_average, rating_count, price_coins,
			rental_price_coins, rental_duration_days, is_premium, total_chapters, total_volumes,
			estimated_reading_time, word_count, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34,
			$35, $36, $37, $38, $39, $40, $41, $42, $43, $44
		)`

	_, err := r.DB.Exec(ctx, query,
		novel.ID, novel.Status, novel.Title, novel.CoverImage, novel.Summary, novel.OwnershipType,
		novel.PrimaryOwnerID, novel.OriginalCreatorID, novel.AccessLevel, novel.LastModifiedByUserID,
		novel.OwnershipTransferredAt, novel.PublishedAt, novel.OriginalLanguage, novel.SourceURL,
		novel.ISBN, novel.AgeRating, novel.ContentWarnings, novel.MatureContent, novel.IsPublic,
		novel.IsFeatured, novel.IsCompleted, novel.IsDeleted, novel.DeletedAt, novel.DeletedByUserID,
		novel.Slug, novel.Tags, novel.Keywords, novel.MetaDescription, novel.ViewCount, novel.LikeCount,
		novel.BookmarkCount, novel.CommentCount, novel.RatingAverage, novel.RatingCount, novel.PriceCoins,
		novel.RentalPriceCoins, novel.RentalDurationDays, novel.IsPremium, novel.TotalChapters,
		novel.TotalVolumes, novel.EstimatedReadingTime, novel.WordCount, novel.CreatedAt, novel.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.NovelRepositoryPG.Create: %w", err)
	}

	return nil
}

// GetByID retrieves a novel by its ID.
func (r *NovelRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	query := `SELECT * FROM novels WHERE id = $1 AND is_deleted = false`
	novel := &domain.Novel{}
	row := r.DB.QueryRow(ctx, query, id)
	err := scanNovel(row, novel)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("novel with id %s not found", id)
		}
		return nil, fmt.Errorf("postgres.NovelRepositoryPG.GetByID: %w", err)
	}
	return novel, nil
}

// GetBySlug retrieves a novel by its slug.
func (r *NovelRepositoryPG) GetBySlug(ctx context.Context, slug string) (*domain.Novel, error) {
	query := `SELECT * FROM novels WHERE slug = $1 AND is_deleted = false`
	novel := &domain.Novel{}
	row := r.DB.QueryRow(ctx, query, slug)
	err := scanNovel(row, novel)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("novel with slug %s not found", slug)
		}
		return nil, fmt.Errorf("postgres.NovelRepositoryPG.GetBySlug: %w", err)
	}
	return novel, nil
}

// Update modifies an existing novel record in the database.
func (r *NovelRepositoryPG) Update(ctx context.Context, novel *domain.Novel) error {
	query := `
		UPDATE novels SET
			status = $2, title = $3, cover_image = $4, summary = $5, ownership_type = $6,
			primary_owner_id = $7, original_creator_id = $8, access_level = $9,
			last_modified_by_user_id = $10, ownership_transferred_at = $11, published_at = $12,
			original_language = $13, source_url = $14, isbn = $15, age_rating = $16,
			content_warnings = $17, mature_content = $18, is_public = $19, is_featured = $20,
			is_completed = $21, is_deleted = $22, deleted_at = $23, deleted_by_user_id = $24,
			slug = $25, tags = $26, keywords = $27, meta_description = $28, view_count = $29,
			like_count = $30, bookmark_count = $31, comment_count = $32, rating_average = $33,
			rating_count = $34, price_coins = $35, rental_price_coins = $36,
			rental_duration_days = $37, is_premium = $38, total_chapters = $39,
			total_volumes = $40, estimated_reading_time = $41, word_count = $42, updated_at = $43
		WHERE id = $1`

	_, err := r.DB.Exec(ctx, query,
		novel.ID, novel.Status, novel.Title, novel.CoverImage, novel.Summary, novel.OwnershipType,
		novel.PrimaryOwnerID, novel.OriginalCreatorID, novel.AccessLevel, novel.LastModifiedByUserID,
		novel.OwnershipTransferredAt, novel.PublishedAt, novel.OriginalLanguage, novel.SourceURL,
		novel.ISBN, novel.AgeRating, novel.ContentWarnings, novel.MatureContent, novel.IsPublic,
		novel.IsFeatured, novel.IsCompleted, novel.IsDeleted, novel.DeletedAt, novel.DeletedByUserID,
		novel.Slug, novel.Tags, novel.Keywords, novel.MetaDescription, novel.ViewCount, novel.LikeCount,
		novel.BookmarkCount, novel.CommentCount, novel.RatingAverage, novel.RatingCount, novel.PriceCoins,
		novel.RentalPriceCoins, novel.RentalDurationDays, novel.IsPremium, novel.TotalChapters,
		novel.TotalVolumes, novel.EstimatedReadingTime, novel.WordCount, novel.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.NovelRepositoryPG.Update: %w", err)
	}

	return nil
}

// Delete soft deletes a novel by its ID.
func (r *NovelRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	// This is a hard delete in this implementation. Soft delete logic should be in the service layer.
	query := `DELETE FROM novels WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.NovelRepositoryPG.Delete: %w", err)
	}
	return nil
}

// List retrieves a paginated list of novels.
func (r *NovelRepositoryPG) List(ctx context.Context, limit, offset int) ([]*domain.Novel, int64, error) {
	query := `SELECT * FROM novels WHERE is_deleted = false ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.DB.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.NovelRepositoryPG.List: %w", err)
	}
	defer rows.Close()

	novels := []*domain.Novel{}
	for rows.Next() {
		novel := &domain.Novel{}
		if err := scanNovel(rows, novel); err != nil {
			return nil, 0, fmt.Errorf("postgres.NovelRepositoryPG.List: failed to scan novel: %w", err)
		}
		novels = append(novels, novel)
	}

	countQuery := `SELECT COUNT(*) FROM novels WHERE is_deleted = false`
	var total int64
	err = r.DB.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.NovelRepositoryPG.List: failed to count novels: %w", err)
	}

	return novels, total, nil
}

// scanNovel is a helper function to scan a row into a Novel struct.
func scanNovel(row pgx.Row, novel *domain.Novel) error {
	return row.Scan(
		&novel.ID, &novel.Status, &novel.Title, &novel.CoverImage, &novel.Summary, &novel.OwnershipType,
		&novel.PrimaryOwnerID, &novel.OriginalCreatorID, &novel.AccessLevel, &novel.LastModifiedByUserID,
		&novel.OwnershipTransferredAt, &novel.PublishedAt, &novel.OriginalLanguage, &novel.SourceURL,
		&novel.ISBN, &novel.AgeRating, &novel.ContentWarnings, &novel.MatureContent, &novel.IsPublic,
		&novel.IsFeatured, &novel.IsCompleted, &novel.IsDeleted, &novel.DeletedAt, &novel.DeletedByUserID,
		&novel.Slug, &novel.Tags, &novel.Keywords, &novel.MetaDescription, &novel.ViewCount, &novel.LikeCount,
		&novel.BookmarkCount, &novel.CommentCount, &novel.RatingAverage, &novel.RatingCount, &novel.PriceCoins,
		&novel.RentalPriceCoins, &novel.RentalDurationDays, &novel.IsPremium, &novel.TotalChapters,
		&novel.TotalVolumes, &novel.EstimatedReadingTime, &novel.WordCount, &novel.CreatedAt, &novel.UpdatedAt,
	)
}
