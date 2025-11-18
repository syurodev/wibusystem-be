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

const novelColumns = `
	id, title, slug, synopsis, cover_image_url, thumbnail_url,
	status, original_language, original_title,
	total_volumes, total_chapters, total_words, view_count,
	favorite_count, rating_average, rating_count, metadata,
	first_published_at, last_chapter_at, completed_at,
	created_at, updated_at, deleted_at
`

// GetByID lấy novel từ database theo ID
func (r *novelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.novels
		WHERE id = $1 AND deleted_at IS NULL
	`, novelColumns)

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
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.novels
		WHERE slug = $1 AND deleted_at IS NULL
	`, novelColumns)

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

// GetByAuthorID lấy danh sách novel theo author ID (via junction table)
func (r *novelRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Novel, error) {
	// Prefix each column with n.
	cols := strings.Split(novelColumns, ", ")
	prefixedCols := make([]string, len(cols))
	for i, col := range cols {
		prefixedCols[i] = "n." + strings.TrimSpace(col)
	}
	novelColumnsWithPrefix := strings.Join(prefixedCols, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.novels n
		INNER JOIN catalog.novel_authors na ON n.id = na.novel_id
		WHERE na.author_id = $1 AND n.deleted_at IS NULL
		ORDER BY n.created_at DESC
		LIMIT $2 OFFSET $3
	`, novelColumnsWithPrefix)

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
			id, title, slug, synopsis, cover_image_url, thumbnail_url,
			status, original_language, original_title, metadata,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
	`

	// Đảm bảo metadata không null
	if novel.Metadata == nil {
		novel.Metadata = json.RawMessage("{}")
	}

	// Đảm bảo synopsis không null
	if novel.Synopsis == nil {
		novel.Synopsis = json.RawMessage("{}")
	}

	_, err := r.pool.Exec(ctx, query,
		novel.ID,
		novel.Title,
		novel.Slug,
		novel.Synopsis,
		novel.CoverImageURL,
		novel.ThumbnailURL,
		novel.Status,
		novel.OriginalLanguage,
		novel.OriginalTitle,
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
		    original_language = $8,
		    original_title = $9,
		    metadata = $10,
		    first_published_at = $11,
		    completed_at = $12,
		    updated_at = NOW()
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
		novel.OriginalLanguage,
		novel.OriginalTitle,
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

	whereClauses = append(whereClauses, "n.deleted_at IS NULL")

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("n.status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	if filter.OriginalLanguage != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("n.original_language = $%d", argIdx))
		args = append(args, *filter.OriginalLanguage)
		argIdx++
	}

	// Full-text search trong title và synopsis
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(n.title ILIKE $%d OR n.synopsis::text ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIdx++
	}

	// Filter by author ID (via junction table)
	var joins []string
	if filter.AuthorID != nil {
		joins = append(joins, "INNER JOIN catalog.novel_authors na ON n.id = na.novel_id")
		whereClauses = append(whereClauses, fmt.Sprintf("na.author_id = $%d", argIdx))
		args = append(args, *filter.AuthorID)
		argIdx++
	}

	// Filter by translator ID (via junction table)
	if filter.TranslatorID != nil {
		joins = append(joins, "INNER JOIN catalog.novel_translators nt ON n.id = nt.novel_id")
		whereClauses = append(whereClauses, fmt.Sprintf("nt.translator_id = $%d", argIdx))
		args = append(args, *filter.TranslatorID)
		argIdx++
	}

	// Filter by genre IDs (via junction table)
	if len(filter.GenreIDs) > 0 {
		joins = append(joins, "INNER JOIN catalog.novel_genres ng ON n.id = ng.novel_id")
		whereClauses = append(whereClauses, fmt.Sprintf("ng.genre_id = ANY($%d)", argIdx))
		args = append(args, filter.GenreIDs)
		argIdx++
	}

	joinClause := ""
	if len(joins) > 0 {
		joinClause = strings.Join(joins, " ")
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Build ORDER BY clause
	orderBy := "n.created_at DESC"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "rating":
			orderBy = "n.rating_average"
		case "views":
			orderBy = "n.view_count"
		case "last_chapter":
			orderBy = "n.last_chapter_at"
		default:
			orderBy = "n." + filter.SortBy
		}

		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT n.id) FROM catalog.novels n %s WHERE %s", joinClause, whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Prefix each column with n.
	cols := strings.Split(novelColumns, ", ")
	prefixedCols := make([]string, len(cols))
	for i, col := range cols {
		prefixedCols[i] = "n." + strings.TrimSpace(col)
	}
	novelColumnsWithPrefix := strings.Join(prefixedCols, ", ")

	// Main query
	query := fmt.Sprintf(`
		SELECT DISTINCT %s
		FROM catalog.novels n
		%s
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, novelColumnsWithPrefix, joinClause, whereClause, orderBy, argIdx, argIdx+1)

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
		SET %s, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// IncrementViewCount tăng view count
func (r *novelRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novels
		SET view_count = view_count + 1,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}
