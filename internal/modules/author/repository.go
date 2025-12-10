package author

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"system/internal/domain"
	db "system/internal/platform/database"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// authorRepository triển khai AuthorRepository sử dụng pgx
type authorRepository struct {
	pool *pgxpool.Pool
}

// NewAuthorRepository tạo một instance mới của authorRepository
func NewAuthorRepository(pool *pgxpool.Pool) domain.AuthorRepository {
	return &authorRepository{pool: pool}
}

// GetByID lấy author từ database theo ID
func (r *authorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	query := `
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       novel_count, total_chapters, total_views, follower_count,
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
		FROM catalog.authors
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	author, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Author])
	if err != nil {
		return nil, err
	}

	return &author, nil
}

// GetBySlug lấy author từ database theo slug
func (r *authorRepository) GetBySlug(ctx context.Context, slug string) (*domain.Author, error) {
	query := `
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       novel_count, total_chapters, total_views, follower_count,
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
		FROM catalog.authors
		WHERE slug = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, slug)
	if err != nil {
		return nil, err
	}

	author, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Author])
	if err != nil {
		return nil, err
	}

	return &author, nil
}

// GetByUserID lấy author từ database theo user ID
func (r *authorRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Author, error) {
	query := `
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       novel_count, total_chapters, total_views, follower_count,
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
		FROM catalog.authors
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	author, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Author])
	if err != nil {
		return nil, err
	}

	return &author, nil
}

// List lấy danh sách authors với filter
func (r *authorRepository) List(ctx context.Context, filter domain.AuthorFilter) ([]*domain.Author, int64, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIdx++
	}

	if filter.IsVerified != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_verified = $%d", argIdx))
		args = append(args, *filter.IsVerified)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.authors WHERE %s", whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	// Main query
	query := fmt.Sprintf(`
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       novel_count, total_chapters, total_views, follower_count,
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
		FROM catalog.authors
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	authors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Author])
	if err != nil {
		return nil, 0, err
	}

	return authors, total, nil
}

// ListSelection lấy danh sách authors rút gọn (chỉ ID và Name)
func (r *authorRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Author, int64, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")
	// whereClauses = append(whereClauses, "is_verified = true") // Optional: only verified authors?

	// Search by name
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereClause := " WHERE " + fmt.Sprintf("%s", whereClauses[0])
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	// Query to get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.authors%s", whereClause)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Query to get authors with pagination (only ID and Name)
	query := fmt.Sprintf(`
		SELECT id, name
		FROM catalog.authors
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var authors []*domain.Author
	for rows.Next() {
		var author domain.Author
		if err := rows.Scan(&author.ID, &author.Name); err != nil {
			return nil, 0, err
		}
		authors = append(authors, &author)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return authors, totalCount, nil
}

// Create tạo author mới
func (r *authorRepository) Create(ctx context.Context, author *domain.Author) error {
	query := `
		INSERT INTO catalog.authors (
			id, user_id, name, slug, biography, avatar_url, social_links, is_verified, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Ensure social_links is not null
	if author.SocialLinks == nil {
		author.SocialLinks = json.RawMessage("{}")
	}

	_, err := r.pool.Exec(ctx, query,
		author.ID,
		author.UserID,
		author.Name,
		author.Slug,
		author.Biography,
		author.AvatarURL,
		author.SocialLinks,
		author.IsVerified,
		author.CreatedBy,
	)

	return err
}

// Update cập nhật author
func (r *authorRepository) Update(ctx context.Context, author *domain.Author) error {
	query := `
		UPDATE catalog.authors
		SET user_id = $2,
		    name = $3,
		    slug = $4,
		    biography = $5,
		    avatar_url = $6,
		    social_links = $7,
		    is_verified = $8,
		    updated_by = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		author.ID,
		author.UserID,
		author.Name,
		author.Slug,
		author.Biography,
		author.AvatarURL,
		author.SocialLinks,
		author.IsVerified,
		author.UpdatedBy,
	)

	return err
}

// Delete xóa mềm author
func (r *authorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.authors
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetNovelAuthors lấy danh sách authors của một novel
func (r *authorRepository) GetNovelAuthors(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelAuthor, error) {
	query := `
		SELECT a.id, a.user_id, a.name, a.slug, a.biography, a.avatar_url, a.social_links,
		       a.novel_count, a.total_chapters, a.total_views, a.follower_count,
		       a.is_verified, a.created_by, a.updated_by, a.created_at, a.updated_at, a.deleted_at, a.deleted_by,
		       na.display_order
		FROM catalog.authors a
		INNER JOIN catalog.novel_authors na ON a.id = na.author_id
		WHERE na.novel_id = $1 AND a.deleted_at IS NULL
		ORDER BY na.display_order ASC
	`

	rows, err := r.pool.Query(ctx, query, novelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var novelAuthors []*domain.NovelAuthor
	for rows.Next() {
		var author domain.Author
		var displayOrder int

		err := rows.Scan(
			&author.ID, &author.UserID, &author.Name, &author.Slug,
			&author.Biography, &author.AvatarURL, &author.SocialLinks,
			&author.NovelCount, &author.TotalChapters, &author.TotalViews, &author.FollowerCount,
			&author.IsVerified, &author.CreatedBy, &author.UpdatedBy, &author.CreatedAt, &author.UpdatedAt, &author.DeletedAt, &author.DeletedBy,
			&displayOrder,
		)
		if err != nil {
			return nil, err
		}

		novelAuthors = append(novelAuthors, &domain.NovelAuthor{
			Author:       &author,
			Role:         "",
			DisplayOrder: displayOrder,
		})
	}

	return novelAuthors, rows.Err()
}

// AddNovelAuthor thêm author cho novel
func (r *authorRepository) AddNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID, displayOrder int) error {
	query := `
		INSERT INTO catalog.novel_authors (novel_id, author_id, display_order)
		VALUES ($1, $2, $3)
		ON CONFLICT (novel_id, author_id) DO UPDATE
		SET display_order = EXCLUDED.display_order
	`

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, novelID, authorID, displayOrder)
	return err
}

// AddNovelAuthors thêm nhiều authors cho novel (Batch Insert)
func (r *authorRepository) AddNovelAuthors(ctx context.Context, novelID uuid.UUID, authorIDs []uuid.UUID, role string) error {
	if len(authorIDs) == 0 {
		return nil
	}

	query := "INSERT INTO catalog.novel_authors (novel_id, author_id, role, display_order) VALUES "
	var args []any
	var values []string

	for i, authorID := range authorIDs {
		// $1, $2, $3, $4
		base := i * 4
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4))
		args = append(args, novelID, authorID, role, i)
	}

	query += strings.Join(values, ",")
	query += " ON CONFLICT (novel_id, author_id) DO UPDATE SET role = EXCLUDED.role, display_order = EXCLUDED.display_order"

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, args...)
	return err
}

// RemoveNovelAuthor xóa author khỏi novel
func (r *authorRepository) RemoveNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID) error {
	query := `
		DELETE FROM catalog.novel_authors
		WHERE novel_id = $1 AND author_id = $2
	`

	_, err := r.pool.Exec(ctx, query, novelID, authorID)
	return err
}

// UpdateStatistics cập nhật thống kê
func (r *authorRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.AuthorStatistics) error {
	var setClauses []string
	var args []any
	argIdx := 2

	args = append(args, id)

	if stats.NovelCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("novel_count = $%d", argIdx))
		args = append(args, *stats.NovelCount)
		argIdx++
	}

	if stats.TotalChapters != nil {
		setClauses = append(setClauses, fmt.Sprintf("total_chapters = $%d", argIdx))
		args = append(args, *stats.TotalChapters)
		argIdx++
	}

	if stats.TotalViews != nil {
		setClauses = append(setClauses, fmt.Sprintf("total_views = $%d", argIdx))
		args = append(args, *stats.TotalViews)
		argIdx++
	}

	if stats.FollowerCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("follower_count = $%d", argIdx))
		args = append(args, *stats.FollowerCount)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE catalog.authors
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// Merge gộp nhiều authors (sources) thành một author (target)
func (r *authorRepository) Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	if len(sourceIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Convert UUIDs to strings for pgx array compatibility
	sourceIDStrings := make([]string, len(sourceIDs))
	for i, id := range sourceIDs {
		sourceIDStrings[i] = id.String()
	}

	// 1. Move novels from source authors to target author
	queryMove := `
		UPDATE catalog.novel_authors
		SET author_id = $1
		WHERE author_id = ANY($2::uuid[])
		AND novel_id NOT IN (
			SELECT novel_id FROM catalog.novel_authors WHERE author_id = $1
		)
	`
	_, err = tx.Exec(ctx, queryMove, targetID, sourceIDStrings)
	if err != nil {
		return err
	}

	// 2. Remove all source author assignments (duplicates that weren't moved)
	queryRemove := `
		DELETE FROM catalog.novel_authors
		WHERE author_id = ANY($1::uuid[])
	`
	_, err = tx.Exec(ctx, queryRemove, sourceIDStrings)
	if err != nil {
		return err
	}

	// 3. Update target author stats
	queryUpdateStats := `
		UPDATE catalog.authors
		SET total_views = total_views + (
				SELECT COALESCE(SUM(total_views), 0) FROM catalog.authors WHERE id = ANY($2::uuid[])
			),
			novel_count = (SELECT COUNT(*) FROM catalog.novel_authors WHERE author_id = $1),
			updated_by = $3,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, queryUpdateStats, targetID, sourceIDStrings, mergedBy)
	if err != nil {
		return err
	}

	// 4. Soft delete source authors
	queryDelete := `
		UPDATE catalog.authors
		SET deleted_at = NOW(),
			deleted_by = $2
		WHERE id = ANY($1::uuid[])
	`
	_, err = tx.Exec(ctx, queryDelete, sourceIDStrings, mergedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
func (r *authorRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	if len(sourceIDs) == 0 {
		return []*domain.Novel{}, nil
	}

	query := `
		SELECT DISTINCT n.id, n.title, n.slug, n.cover_image_url
		FROM catalog.novels n
		JOIN catalog.novel_authors na ON n.id = na.novel_id
		WHERE na.author_id = ANY($1::uuid[])
		AND n.deleted_at IS NULL
		ORDER BY n.title ASC
	`

	// Convert UUIDs to strings for pgx array compatibility
	sourceIDStrings := make([]string, len(sourceIDs))
	for i, id := range sourceIDs {
		sourceIDStrings[i] = id.String()
	}

	rows, err := r.pool.Query(ctx, query, sourceIDStrings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var novels []*domain.Novel
	for rows.Next() {
		novel := &domain.Novel{}
		var id uuid.UUID
		var title, slug string
		var coverImageURL *string

		if err := rows.Scan(&id, &title, &slug, &coverImageURL); err != nil {
			return nil, err
		}

		novel.ID = id
		novel.Title = title
		novel.Slug = slug
		novel.CoverImageURL = coverImageURL

		novels = append(novels, novel)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return novels, nil
}
