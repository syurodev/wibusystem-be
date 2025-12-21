// ============================================================================
// Artist Repository
// ============================================================================
//
// Repository này thực hiện các thao tác database cho Artist:
//
// Tables:
//   - catalog.artists: Bảng chính chứa thông tin artist
//   - catalog.novel_artists: Bảng junction cho quan hệ N:N với novels
//
// Flow tổng quan:
//   Handler -> UseCase -> Service -> Repository -> Database
//
// Merge Flow (trong một transaction):
//   1. MergeMoveNovels: Move novels từ source artists sang target
//   2. MergeRemoveDuplicates: Xóa các assignments trùng lặp
//   3. MergeUpdateStats: Cập nhật novel_count của target
//   4. MergeSoftDelete: Soft delete source artists
//
// ============================================================================

package artist

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"system/internal/domain"
	db "system/internal/platform/database"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ============================================================================
// Embedded SQL Queries
// ============================================================================

//go:embed queries/get_by_id.sql
var getByIDQuery string

//go:embed queries/get_by_slug.sql
var getBySlugQuery string

//go:embed queries/get_by_user_id.sql
var getByUserIDQuery string

//go:embed queries/create.sql
var createQuery string

//go:embed queries/update.sql
var updateQuery string

//go:embed queries/delete.sql
var deleteQuery string

//go:embed queries/get_novel_artists.sql
var getNovelArtistsQuery string

//go:embed queries/add_novel_artist.sql
var addNovelArtistQuery string

//go:embed queries/remove_novel_artist.sql
var removeNovelArtistQuery string

//go:embed queries/merge_move_novels.sql
var mergeMoveNovelsQuery string

//go:embed queries/merge_remove_duplicates.sql
var mergeRemoveDuplicatesQuery string

//go:embed queries/merge_update_stats.sql
var mergeUpdateStatsQuery string

//go:embed queries/merge_soft_delete.sql
var mergeSoftDeleteQuery string

//go:embed queries/get_merge_preview.sql
var getMergePreviewQuery string

// ============================================================================
// Repository Implementation
// ============================================================================

// artistRepository triển khai ArtistRepository sử dụng pgx
type artistRepository struct {
	pool *pgxpool.Pool
}

// NewArtistRepository tạo một instance mới của artistRepository
func NewArtistRepository(pool *pgxpool.Pool) domain.ArtistRepository {
	return &artistRepository{pool: pool}
}

// ============================================================================
// Basic CRUD Operations
// ============================================================================

// GetByID lấy artist từ database theo ID
func (r *artistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Artist, error) {
	rows, err := r.pool.Query(ctx, getByIDQuery, id)
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
	rows, err := r.pool.Query(ctx, getBySlugQuery, slug)
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
	rows, err := r.pool.Query(ctx, getByUserIDQuery, userID)
	if err != nil {
		return nil, err
	}

	artist, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Artist])
	if err != nil {
		return nil, err
	}

	return &artist, nil
}

// ============================================================================
// List Operations (Dynamic Query Building)
// ============================================================================

// List lấy danh sách artists với filter
// Query được build động do có nhiều filter options
// Note: SortBy đã được validate ở DTO với binding:"oneof=name novels created"
// và được map sang DB column names ở Service layer
func (r *artistRepository) List(ctx context.Context, filter domain.ArtistFilter) ([]*domain.Artist, int64, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	// Filter by search query (name ILIKE)
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIdx++
	}

	// Filter by specialization
	if filter.Specialization != nil && *filter.Specialization != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("specialization = $%d", argIdx))
		args = append(args, *filter.Specialization)
		argIdx++
	}

	// Filter by verified status
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
	// SortBy đã được validate và map ở layers trước (DTO -> Service)
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	// Main query với embedded base columns
	query := fmt.Sprintf(`
		SELECT id, user_id, name, slug, biography, avatar_url, social_links,
		       specialization, portfolio_url, novel_count, artwork_count, follower_count,
		       is_verified, metadata, version, created_by, updated_by, 
		       created_at, updated_at, deleted_at, deleted_by
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
// Được sử dụng cho selection dropdowns
func (r *artistRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Artist, int64, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

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

// ============================================================================
// Create/Update/Delete Operations
// ============================================================================

// Create tạo artist mới
func (r *artistRepository) Create(ctx context.Context, artist *domain.Artist) error {
	// Ensure social_links is not null
	if artist.SocialLinks == nil {
		artist.SocialLinks = json.RawMessage("{}")
	}

	_, err := r.pool.Exec(ctx, createQuery,
		artist.ID,
		artist.UserID,
		artist.Name,
		artist.Slug,
		artist.Biography,
		artist.AvatarURL,
		artist.SocialLinks,
		artist.Specialization,
		artist.PortfolioURL,
		artist.IsVerified,
		artist.CreatedBy,
	)

	return err
}

// Update cập nhật artist
func (r *artistRepository) Update(ctx context.Context, artist *domain.Artist) error {
	_, err := r.pool.Exec(ctx, updateQuery,
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
	_, err := r.pool.Exec(ctx, deleteQuery, id)
	return err
}

// ============================================================================
// Novel-Artist Relationship Operations
// ============================================================================

// GetNovelArtists lấy danh sách artists của một novel
func (r *artistRepository) GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelArtist, error) {
	rows, err := r.pool.Query(ctx, getNovelArtistsQuery, novelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var novelArtists []*domain.NovelArtist
	for rows.Next() {
		var artist domain.Artist
		var displayOrder int

		err := rows.Scan(
			&artist.ID, &artist.UserID, &artist.Name, &artist.Slug,
			&artist.Biography, &artist.AvatarURL, &artist.SocialLinks,
			&artist.Specialization, &artist.NovelCount, &artist.ArtworkCount, &artist.FollowerCount,
			&artist.IsVerified, &artist.CreatedBy, &artist.UpdatedBy, &artist.CreatedAt, &artist.UpdatedAt, &artist.DeletedAt, &artist.DeletedBy,
			&displayOrder,
		)
		if err != nil {
			return nil, err
		}

		novelArtists = append(novelArtists, &domain.NovelArtist{
			Artist:       &artist,
			Role:         "",
			DisplayOrder: displayOrder,
		})
	}

	return novelArtists, rows.Err()
}

// AddNovelArtist thêm artist cho novel
func (r *artistRepository) AddNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, displayOrder int) error {
	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, addNovelArtistQuery, novelID, artistID, displayOrder)
	return err
}

// AddNovelArtists thêm nhiều artists cho novel (Batch Insert)
func (r *artistRepository) AddNovelArtists(ctx context.Context, novelID uuid.UUID, artistIDs []uuid.UUID, role string) error {
	if len(artistIDs) == 0 {
		return nil
	}

	// Build dynamic batch insert query
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
	_, err := r.pool.Exec(ctx, removeNovelArtistQuery, novelID, artistID, role)
	return err
}

// ============================================================================
// Statistics Operations
// ============================================================================

// UpdateStatistics cập nhật thống kê
// Query được build động do các field là optional
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

// ============================================================================
// Merge Operations
// ============================================================================

// Merge gộp nhiều artists (sources) thành một artist (target)
// Thực hiện trong một transaction với 4 bước:
//  1. Move novels từ source artists sang target (tránh duplicates)
//  2. Xóa các novel assignments còn lại của source artists
//  3. Cập nhật novel_count của target artist
//  4. Soft delete source artists
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

	// Step 1: Move novels from source artists to target artist
	_, err = tx.Exec(ctx, mergeMoveNovelsQuery, targetID, sourceIDStrings)
	if err != nil {
		return err
	}

	// Step 2: Remove all source artist assignments (duplicates that weren't moved)
	_, err = tx.Exec(ctx, mergeRemoveDuplicatesQuery, sourceIDStrings)
	if err != nil {
		return err
	}

	// Step 3: Update target artist stats
	_, err = tx.Exec(ctx, mergeUpdateStatsQuery, targetID, mergedBy)
	if err != nil {
		return err
	}

	// Step 4: Soft delete source artists
	_, err = tx.Exec(ctx, mergeSoftDeleteQuery, sourceIDStrings, mergedBy)
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

	// Convert UUIDs to strings for pgx array compatibility
	sourceIDStrings := make([]string, len(sourceIDs))
	for i, id := range sourceIDs {
		sourceIDStrings[i] = id.String()
	}

	rows, err := r.pool.Query(ctx, getMergePreviewQuery, sourceIDStrings)
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
