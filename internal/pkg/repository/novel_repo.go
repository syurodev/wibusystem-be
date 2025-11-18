package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// novelRepository triển khai NovelRepository sử dụng pgx
type novelRepository struct {
	pool *pgxpool.Pool
}

// NewNovelRepository tạo một instance mới của novelRepository
func NewNovelRepository(pool *pgxpool.Pool) domain.NovelRepository {
	return &novelRepository{pool: pool}
}

// GetByID lấy novel từ database theo ID
func (r *novelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	query := `
		SELECT id, title, slug, author_id, synopsis, cover_image_url, thumbnail_url,
		       status, total_volumes, total_chapters, total_words, view_count,
		       favorite_count, rating_average, rating_count, metadata,
		       first_published_at, last_chapter_at, completed_at,
		       created_at, updated_at, deleted_at
		FROM catalog.novels
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	novel, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Novel])
	if err != nil {
		return nil, err
	}

	return &novel, nil
}

// GetBySlug lấy novel từ database theo slug
func (r *novelRepository) GetBySlug(ctx context.Context, slug string) (*domain.Novel, error) {
	query := `
		SELECT id, title, slug, author_id, synopsis, cover_image_url, thumbnail_url,
		       status, total_volumes, total_chapters, total_words, view_count,
		       favorite_count, rating_average, rating_count, metadata,
		       first_published_at, last_chapter_at, completed_at,
		       created_at, updated_at, deleted_at
		FROM catalog.novels
		WHERE slug = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, slug)
	if err != nil {
		return nil, err
	}

	novel, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Novel])
	if err != nil {
		return nil, err
	}

	return &novel, nil
}

// GetByAuthorID lấy danh sách novel theo author ID
func (r *novelRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Novel, error) {
	query := `
		SELECT id, title, slug, author_id, synopsis, cover_image_url, thumbnail_url,
		       status, total_volumes, total_chapters, total_words, view_count,
		       favorite_count, rating_average, rating_count, metadata,
		       first_published_at, last_chapter_at, completed_at,
		       created_at, updated_at, deleted_at
		FROM catalog.novels
		WHERE author_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}

	novels, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Novel])
	if err != nil {
		return nil, err
	}

	return novels, nil
}

// Create tạo novel mới trong database
func (r *novelRepository) Create(ctx context.Context, novel *domain.Novel) error {
	query := `
		INSERT INTO catalog.novels (
			id, title, slug, author_id, synopsis, cover_image_url, thumbnail_url,
			status, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Đảm bảo metadata không null
	if novel.Metadata == nil {
		novel.Metadata = json.RawMessage("{}")
	}

	_, err := r.pool.Exec(ctx, query,
		novel.ID,
		novel.Title,
		novel.Slug,
		novel.AuthorID,
		novel.Synopsis,
		novel.CoverImageURL,
		novel.ThumbnailURL,
		novel.Status,
		novel.Metadata,
	)

	return err
}

// Update cập nhật thông tin novel
func (r *novelRepository) Update(ctx context.Context, novel *domain.Novel) error {
	query := `
		UPDATE catalog.novels
		SET title = $2,
		    slug = $3,
		    synopsis = $4,
		    cover_image_url = $5,
		    thumbnail_url = $6,
		    status = $7,
		    metadata = $8,
		    first_published_at = $9,
		    completed_at = $10
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		novel.ID,
		novel.Title,
		novel.Slug,
		novel.Synopsis,
		novel.CoverImageURL,
		novel.ThumbnailURL,
		novel.Status,
		novel.Metadata,
		novel.FirstPublishedAt,
		novel.CompletedAt,
	)

	return err
}

// Delete xóa mềm novel
func (r *novelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novels
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// List lấy danh sách novel với filter và pagination
func (r *novelRepository) List(ctx context.Context, filter domain.NovelFilter) ([]*domain.Novel, int64, error) {
	// Build WHERE clause động
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	if filter.AuthorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("author_id = $%d", argIdx))
		args = append(args, *filter.AuthorID)
		argIdx++
	}

	// Full-text search trong title và synopsis
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(title ILIKE $%d OR synopsis::text ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIdx++
	}

	// Filter by tags trong metadata
	if len(filter.Tags) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("metadata @> $%d::jsonb", argIdx))
		tagsJSON, _ := json.Marshal(map[string][]string{"tags": filter.Tags})
		args = append(args, tagsJSON)
		argIdx++
	}

	// Filter by categories trong metadata
	if len(filter.Categories) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("metadata @> $%d::jsonb", argIdx))
		categoriesJSON, _ := json.Marshal(map[string][]string{"categories": filter.Categories})
		args = append(args, categoriesJSON)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "rating":
			orderBy = "rating_average"
		case "views":
			orderBy = "view_count"
		case "last_chapter":
			orderBy = "last_chapter_at"
		default:
			orderBy = filter.SortBy
		}

		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.novels WHERE %s", whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT id, title, slug, author_id, synopsis, cover_image_url, thumbnail_url,
		       status, total_volumes, total_chapters, total_words, view_count,
		       favorite_count, rating_average, rating_count, metadata,
		       first_published_at, last_chapter_at, completed_at,
		       created_at, updated_at, deleted_at
		FROM catalog.novels
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	novels, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Novel])
	if err != nil {
		return nil, 0, err
	}

	return novels, total, nil
}

// UpdateStatistics cập nhật thống kê của novel
func (r *novelRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.NovelStatistics) error {
	var setClauses []string
	var args []interface{}
	argIdx := 2

	args = append(args, id) // $1 là id

	if stats.ViewCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("view_count = $%d", argIdx))
		args = append(args, *stats.ViewCount)
		argIdx++
	}

	if stats.FavoriteCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("favorite_count = $%d", argIdx))
		args = append(args, *stats.FavoriteCount)
		argIdx++
	}

	if stats.RatingAverage != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating_average = $%d", argIdx))
		args = append(args, *stats.RatingAverage)
		argIdx++
	}

	if stats.RatingCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating_count = $%d", argIdx))
		args = append(args, *stats.RatingCount)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // Không có gì để update
	}

	query := fmt.Sprintf(`
		UPDATE catalog.novels
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// IncrementViewCount tăng view count
func (r *novelRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novels
		SET view_count = view_count + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}
