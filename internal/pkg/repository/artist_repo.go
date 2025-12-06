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
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
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
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
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
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
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
	var args []any
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
		       is_verified, created_by, updated_by, created_at, updated_at, deleted_at, deleted_by
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

// ListSelection lấy danh sách artists rút gọn (chỉ ID và Name)
func (r *artistRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Artist, int64, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")
	// whereClauses = append(whereClauses, "is_verified = true") // Optional

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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.artists%s", whereClause)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Query to get artists with pagination (only ID and Name)
	query := fmt.Sprintf(`
		SELECT id, name
		FROM catalog.artists
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

	var artists []*domain.Artist
	for rows.Next() {
		var artist domain.Artist
		if err := rows.Scan(&artist.ID, &artist.Name); err != nil {
			return nil, 0, err
		}
		artists = append(artists, &artist)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return artists, totalCount, nil
}

// Create tạo artist mới
func (r *artistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	query := `
		INSERT INTO catalog.artists (
			id, user_id, name, slug, biography, avatar_url, social_links,
			specialization, is_verified, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		artist.CreatedBy,
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
		    is_verified = $9,
		    updated_by = $10
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
		artist.UpdatedBy,
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
		       a.is_verified, a.created_by, a.updated_by, a.created_at, a.updated_at, a.deleted_at, a.deleted_by,
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
			&artist.IsVerified, &artist.CreatedBy, &artist.UpdatedBy, &artist.CreatedAt, &artist.UpdatedAt, &artist.DeletedAt, &artist.DeletedBy,
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
func (r *artistRepository) AddNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, displayOrder int) error {
	query := `
		INSERT INTO catalog.novel_artists (novel_id, artist_id, display_order)
		VALUES ($1, $2, $3)
		ON CONFLICT (novel_id, artist_id) DO UPDATE
		SET display_order = EXCLUDED.display_order
	`

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, novelID, artistID, displayOrder)
	return err
}

// AddNovelArtists thêm nhiều artists cho novel (Batch Insert)
func (r *artistRepository) AddNovelArtists(ctx context.Context, novelID uuid.UUID, artistIDs []uuid.UUID, role string) error {
	if len(artistIDs) == 0 {
		return nil
	}

	query := "INSERT INTO catalog.novel_artists (novel_id, artist_id, role, display_order) VALUES "
	var args []any
	var values []string

	for i, artistID := range artistIDs {
		// $1, $2, $3, $4
		base := i * 4
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4))
		args = append(args, novelID, artistID, role, i)
	}

	query += strings.Join(values, ",")
	query += " ON CONFLICT (novel_id, artist_id, role) DO UPDATE SET display_order = EXCLUDED.display_order"

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, args...)
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
	var args []any
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

// Merge gộp nhiều artists (sources) thành một artist (target)
func (r *artistRepository) Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
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

	// 1. Move novels from source artists to target artist
	queryMove := `
		UPDATE catalog.novel_artists
		SET artist_id = $1
		WHERE artist_id = ANY($2::uuid[])
		AND novel_id NOT IN (
			SELECT novel_id FROM catalog.novel_artists WHERE artist_id = $1
		)
	`
	_, err = tx.Exec(ctx, queryMove, targetID, sourceIDStrings)
	if err != nil {
		return err
	}

	// 2. Remove all source artist assignments (duplicates that weren't moved)
	queryRemove := `
		DELETE FROM catalog.novel_artists
		WHERE artist_id = ANY($1::uuid[])
	`
	_, err = tx.Exec(ctx, queryRemove, sourceIDStrings)
	if err != nil {
		return err
	}

	// 3. Update target artist stats
	// Note: We might want to sum novel_count or just recount. Recounting is safer.
	// For artwork_count and follower_count, summing might be inappropriate if duplicates exist, 
	// but without detailed tracking, simple sum or ignoring is the choice. 
	// Let's assume we just want to update novel_count correctly based on the table.
	// For other stats like follower_count, it's complex to merge without user_following table. 
	// Let's stick to simple logic: update novel_count from relation table.
	queryUpdateStats := `
		UPDATE catalog.artists
		SET novel_count = (SELECT COUNT(*) FROM catalog.novel_artists WHERE artist_id = $1),
			updated_by = $2,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, queryUpdateStats, targetID, mergedBy)
	if err != nil {
		return err
	}

	// 4. Soft delete source artists
	queryDelete := `
		UPDATE catalog.artists
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
func (r *artistRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	if len(sourceIDs) == 0 {
		return []*domain.Novel{}, nil
	}

	query := `
		SELECT DISTINCT n.id, n.title, n.slug, n.cover_image_url
		FROM catalog.novels n
		JOIN catalog.novel_artists na ON n.id = na.novel_id
		WHERE na.artist_id = ANY($1::uuid[])
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
