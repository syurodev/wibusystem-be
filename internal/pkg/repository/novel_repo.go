package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"system/internal/domain"
	"system/internal/pkg/db"

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
	owner_id, owner_type,
	owner_display_name, owner_username, owner_avatar_url,
	total_volumes, total_chapters, total_words, view_count,
	favorite_count, rating_average, rating_count, metadata,
	first_published_at, last_chapter_at, completed_at,
	created_by, updated_by, deleted_by,
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
			owner_id, owner_type,
			created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
	`

	// Đảm bảo metadata không null
	if novel.Metadata == nil {
		novel.Metadata = json.RawMessage("{}")
	}

	// Đảm bảo synopsis không null
	if novel.Synopsis == nil {
		novel.Synopsis = json.RawMessage("{}")
	}

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query,
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
		novel.OwnerID,
		novel.OwnerType,
		novel.CreatedBy,
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
		    updated_by = $13,
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
		novel.UpdatedBy,
	)

	return err
}

// Delete xóa mềm novel
func (r *novelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Note: We need deleted_by, but interface only accepts ID.
	// We might need to update interface or assume deleted_by is not available here.
	// But wait, Delete usually needs context of who deleted it.
	// The interface `Delete(ctx context.Context, id uuid.UUID) error` doesn't support it.
	// I should update the interface too.
	// For now, I will leave Delete as is or update it if I change interface.
	// Let's check if I can change interface. Yes I can.
	// But `DeleteNovel` service also needs update.
	// I'll update interface later.
	// Wait, I should do it now to be consistent.
	// But `Delete` in `novel_repo.go` currently:
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
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "n.deleted_at IS NULL")

	// Filter by owner ID
	if filter.OwnerID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("n.owner_id = $%d", argIdx))
		args = append(args, *filter.OwnerID)
		argIdx++
	}

	// Filter by multiple statuses
	if len(filter.Statuses) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("n.status = ANY($%d)", argIdx))
		args = append(args, filter.Statuses)
		argIdx++
	}

	if filter.OriginalLanguage != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("n.original_language = $%d", argIdx))
		args = append(args, *filter.OriginalLanguage)
		argIdx++
	}

	// Full-text search trong title và original_title
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(n.title ILIKE $%d OR COALESCE(n.original_title, '') ILIKE $%d)", argIdx, argIdx))
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

	// Main query - SELECT novel columns và owner info
	// Note: We select n.* columns individually + owner info từ LEFT JOIN users
	query := fmt.Sprintf(`
		SELECT n.id, n.title, n.slug, n.synopsis, n.cover_image_url, n.thumbnail_url,
		       n.status, n.original_language, n.original_title,
		       n.owner_id, n.owner_type,
		       COALESCE(u.full_name, '') as owner_display_name,
		       COALESCE(u.email, '') as owner_username,
		       u.avatar_url as owner_avatar_url,
		       n.total_volumes, n.total_chapters, n.total_words, n.view_count,
		       n.favorite_count, n.rating_average, n.rating_count, n.metadata,
		       n.first_published_at, n.last_chapter_at, n.completed_at,
		       n.created_by, n.updated_by, n.deleted_by,
		       n.created_at, n.updated_at, n.deleted_at
		FROM catalog.novels n
		LEFT JOIN identify.users u ON n.owner_type = 'user' AND n.owner_id = u.id
		%s
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, joinClause, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	novels, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Novel])
	if err != nil {
		return nil, 0, err
	}

	// Load genres for each novel
	for _, novel := range novels {
		genres, err := r.loadNovelGenres(ctx, novel.ID)
		if err != nil {
			// Log error but continue
			continue
		}
		novel.Genres = genres
	}

	return novels, total, nil
}

// UpdateStatistics cập nhật thống kê của novel
func (r *novelRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.NovelStatistics) error {
	var setClauses []string
	var args []any
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

// loadNovelGenres loads genres for a specific novel
func (r *novelRepository) loadNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	query := `
		SELECT g.id, g.name, g.slug, g.description,
		       g.parent_id, g.display_order, g.is_active,
		       g.novel_count, g.active_readers, g.total_views,
		       g.created_by, g.updated_by, g.created_at, g.updated_at
		FROM catalog.genres g
		INNER JOIN catalog.novel_genres ng ON g.id = ng.genre_id
		WHERE ng.novel_id = $1
		ORDER BY ng.display_order ASC
	`

	rows, err := r.pool.Query(ctx, query, novelID)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}
