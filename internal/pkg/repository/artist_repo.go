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

// artistRepository triển khai ArtistRepository sử dụng pgx
type artistRepository struct {
	pool *pgxpool.Pool
}

// NewArtistRepository tạo một instance mới của artistRepository
func NewArtistRepository(pool *pgxpool.Pool) domain.ArtistRepository {
	return &artistRepository{pool: pool}
}

// GetByID lấy artist từ database theo ID
func (r *artistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Artist, error) {
	query := `
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       specialization, novel_count, artwork_count, follower_count,
		       is_verified, created_at, updated_at, deleted_at
		FROM catalog.artists
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	artist, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Artist])
	if err != nil {
		return nil, err
	}

	return &artist, nil
}

// GetBySlug lấy artist từ database theo slug
func (r *artistRepository) GetBySlug(ctx context.Context, slug string) (*domain.Artist, error) {
	query := `
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       specialization, novel_count, artwork_count, follower_count,
		       is_verified, created_at, updated_at, deleted_at
		FROM catalog.artists
		WHERE slug = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, slug)
	if err != nil {
		return nil, err
	}

	artist, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Artist])
	if err != nil {
		return nil, err
	}

	return &artist, nil
}

// GetByUserID lấy artist từ database theo user ID
func (r *artistRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Artist, error) {
	query := `
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       specialization, novel_count, artwork_count, follower_count,
		       is_verified, created_at, updated_at, deleted_at
		FROM catalog.artists
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	artist, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Artist])
	if err != nil {
		return nil, err
	}

	return &artist, nil
}

// List lấy danh sách artists với filter
func (r *artistRepository) List(ctx context.Context, filter domain.ArtistFilter) ([]*domain.Artist, int64, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIdx++
	}

	if filter.Specialization != nil && *filter.Specialization != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("specialization = $%d", argIdx))
		args = append(args, *filter.Specialization)
		argIdx++
	}

	if filter.IsVerified != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_verified = $%d", argIdx))
		args = append(args, *filter.IsVerified)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.artists WHERE %s", whereClause)
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
		       specialization, novel_count, artwork_count, follower_count,
		       is_verified, created_at, updated_at, deleted_at
		FROM catalog.artists
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	artists, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Artist])
	if err != nil {
		return nil, 0, err
	}

	return artists, total, nil
}

// Create tạo artist mới
func (r *artistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	query := `
		INSERT INTO catalog.artists (
			id, user_id, name, slug, biography, avatar_url, social_links,
			specialization, is_verified
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Ensure social_links is not null
	if artist.SocialLinks == nil {
		artist.SocialLinks = json.RawMessage("{}")
	}

	_, err := r.pool.Exec(ctx, query,
		artist.ID,
		artist.UserID,
		artist.Name,
		artist.Slug,
		artist.Biography,
		artist.AvatarURL,
		artist.SocialLinks,
		artist.Specialization,
		artist.IsVerified,
	)

	return err
}

// Update cập nhật artist
func (r *artistRepository) Update(ctx context.Context, artist *domain.Artist) error {
	query := `
		UPDATE catalog.artists
		SET user_id = $2,
		    name = $3,
		    slug = $4,
		    biography = $5,
		    avatar_url = $6,
		    social_links = $7,
		    specialization = $8,
		    is_verified = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		artist.ID,
		artist.UserID,
		artist.Name,
		artist.Slug,
		artist.Biography,
		artist.AvatarURL,
		artist.SocialLinks,
		artist.Specialization,
		artist.IsVerified,
	)

	return err
}

// Delete xóa mềm artist
func (r *artistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.artists
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetNovelArtists lấy danh sách artists của một novel
func (r *artistRepository) GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelArtist, error) {
	query := `
		SELECT a.id, a.user_id, a.name, a.slug, a.biography, a.avatar_url, a.social_links,
		       a.specialization, a.novel_count, a.artwork_count, a.follower_count,
		       a.is_verified, a.created_at, a.updated_at, a.deleted_at,
		       na.role, na.display_order
		FROM catalog.artists a
		INNER JOIN catalog.novel_artists na ON a.id = na.artist_id
		WHERE na.novel_id = $1 AND a.deleted_at IS NULL
		ORDER BY na.display_order ASC
	`

	rows, err := r.pool.Query(ctx, query, novelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var novelArtists []*domain.NovelArtist
	for rows.Next() {
		var artist domain.Artist
		var role string
		var displayOrder int

		err := rows.Scan(
			&artist.ID, &artist.UserID, &artist.Name, &artist.Slug,
			&artist.Biography, &artist.AvatarURL, &artist.SocialLinks,
			&artist.Specialization, &artist.NovelCount, &artist.ArtworkCount, &artist.FollowerCount,
			&artist.IsVerified, &artist.CreatedAt, &artist.UpdatedAt, &artist.DeletedAt,
			&role, &displayOrder,
		)
		if err != nil {
			return nil, err
		}

		novelArtists = append(novelArtists, &domain.NovelArtist{
			Artist:       &artist,
			Role:         role,
			DisplayOrder: displayOrder,
		})
	}

	return novelArtists, rows.Err()
}

// AddNovelArtist thêm artist cho novel
func (r *artistRepository) AddNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, role string, displayOrder int) error {
	query := `
		INSERT INTO catalog.novel_artists (novel_id, artist_id, role, display_order)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (novel_id, artist_id, role) DO UPDATE
		SET display_order = EXCLUDED.display_order
	`

	_, err := r.pool.Exec(ctx, query, novelID, artistID, role, displayOrder)
	return err
}

// RemoveNovelArtist xóa artist khỏi novel
func (r *artistRepository) RemoveNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, role string) error {
	query := `
		DELETE FROM catalog.novel_artists
		WHERE novel_id = $1 AND artist_id = $2 AND role = $3
	`

	_, err := r.pool.Exec(ctx, query, novelID, artistID, role)
	return err
}

// UpdateStatistics cập nhật thống kê
func (r *artistRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.ArtistStatistics) error {
	var setClauses []string
	var args []interface{}
	argIdx := 2

	args = append(args, id)

	if stats.NovelCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("novel_count = $%d", argIdx))
		args = append(args, *stats.NovelCount)
		argIdx++
	}

	if stats.ArtworkCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("artwork_count = $%d", argIdx))
		args = append(args, *stats.ArtworkCount)
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
		UPDATE catalog.artists
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
