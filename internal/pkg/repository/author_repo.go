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
		       is_verified, created_at, updated_at, deleted_at
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
		       is_verified, created_at, updated_at, deleted_at
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
		       is_verified, created_at, updated_at, deleted_at
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
	var args []interface{}
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
		       is_verified, created_at, updated_at, deleted_at
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

// Create tạo author mới
func (r *authorRepository) Create(ctx context.Context, author *domain.Author) error {
	query := `
		INSERT INTO catalog.authors (
			id, user_id, name, slug, biography, avatar_url, social_links, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
		    is_verified = $8
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
		       a.is_verified, a.created_at, a.updated_at, a.deleted_at,
		       na.role, na.display_order
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
		var role string
		var displayOrder int

		err := rows.Scan(
			&author.ID, &author.UserID, &author.Name, &author.Slug,
			&author.Biography, &author.AvatarURL, &author.SocialLinks,
			&author.NovelCount, &author.TotalChapters, &author.TotalViews, &author.FollowerCount,
			&author.IsVerified, &author.CreatedAt, &author.UpdatedAt, &author.DeletedAt,
			&role, &displayOrder,
		)
		if err != nil {
			return nil, err
		}

		novelAuthors = append(novelAuthors, &domain.NovelAuthor{
			Author:       &author,
			Role:         role,
			DisplayOrder: displayOrder,
		})
	}

	return novelAuthors, rows.Err()
}

// AddNovelAuthor thêm author cho novel
func (r *authorRepository) AddNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID, role string, displayOrder int) error {
	query := `
		INSERT INTO catalog.novel_authors (novel_id, author_id, role, display_order)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (novel_id, author_id) DO UPDATE
		SET role = EXCLUDED.role, display_order = EXCLUDED.display_order
	`

	_, err := r.pool.Exec(ctx, query, novelID, authorID, role, displayOrder)
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
	var args []interface{}
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
