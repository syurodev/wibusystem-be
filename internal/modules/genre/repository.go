package genre

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"system/internal/domain"
	db "system/internal/platform/database"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL queries embedded từ file
//
//go:embed queries/get_by_id.sql
var getByIDQuery string

//go:embed queries/get_by_slug.sql
var getBySlugQuery string

//go:embed queries/get_by_parent_id.sql
var getByParentIDQuery string

//go:embed queries/get_root_genres.sql
var getRootGenresQuery string

//go:embed queries/create.sql
var createQuery string

//go:embed queries/update.sql
var updateQuery string

//go:embed queries/delete.sql
var deleteQuery string

//go:embed queries/get_novel_genres.sql
var getNovelGenresQuery string

//go:embed queries/add_novel_genre.sql
var addNovelGenreQuery string

//go:embed queries/remove_novel_genre.sql
var removeNovelGenreQuery string

//go:embed queries/delete_novel_genres.sql
var deleteNovelGenresQuery string

//go:embed queries/insert_novel_genre.sql
var insertNovelGenreQuery string

//go:embed queries/increment_novel_count.sql
var incrementNovelCountQuery string

//go:embed queries/increment_total_views.sql
var incrementTotalViewsQuery string

//go:embed queries/update_active_readers.sql
var updateActiveReadersQuery string

//go:embed queries/get_genres_by_novel_ids.sql
var getGenresByNovelIDsQuery string

//go:embed queries/merge_move_novels.sql
var mergeMoveNovelsQuery string

//go:embed queries/merge_remove_source_assignments.sql
var mergeRemoveSourceAssignmentsQuery string

//go:embed queries/merge_update_target_stats.sql
var mergeUpdateTargetStatsQuery string

//go:embed queries/merge_recount_novels.sql
var mergeRecountNovelsQuery string

//go:embed queries/merge_delete_sources.sql
var mergeDeleteSourcesQuery string

//go:embed queries/get_merge_preview.sql
var getMergePreviewQuery string

// genreRepository triển khai GenreRepository sử dụng pgx
type genreRepository struct {
	pool *pgxpool.Pool
}

// NewGenreRepository tạo một instance mới của genreRepository
func NewGenreRepository(pool *pgxpool.Pool) domain.GenreRepository {
	return &genreRepository{pool: pool}
}

const genreColumns = `
	id, name, slug, description, parent_id,
	is_active, novel_count, anime_count, manga_count, active_readers, total_views,
	created_by, deleted_by, updated_by, created_at, updated_at, deleted_at, version
`

// GetByID lấy genre từ database theo ID
func (r *genreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	rows, err := r.pool.Query(ctx, getByIDQuery, id)
	if err != nil {
		return nil, err
	}

	genre, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return &genre, nil
}

// GetBySlug lấy genre từ database theo slug
func (r *genreRepository) GetBySlug(ctx context.Context, slug string) (*domain.Genre, error) {
	rows, err := r.pool.Query(ctx, getBySlugQuery, slug)
	if err != nil {
		return nil, err
	}

	genre, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return &genre, nil
}

// GetAll lấy tất cả genres
func (r *genreRepository) GetAll(ctx context.Context, activeOnly bool) ([]*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE deleted_at IS NULL
	`, genreColumns)

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY name ASC"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// List lấy danh sách genres với pagination, search và sort
func (r *genreRepository) List(ctx context.Context, offset, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*domain.Genre, int, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	if activeOnly {
		whereClauses = append(whereClauses, "is_active = true")
	}

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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.genres%s", whereClause)
	var totalCount int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderClause := "name ASC" // Default sort
	if sortBy != "" {
		orderField := ""
		switch sortBy {
		case "name":
			orderField = "name"
		case "views":
			orderField = "total_views"
		case "series":
			orderField = "(novel_count + anime_count + manga_count)"
		case "created":
			orderField = "created_at"
		case "updated":
			orderField = "updated_at"
		case "readers":
			orderField = "active_readers"
		default:
			orderField = "name"
		}

		if orderField != "" {
			direction := "ASC"
			if sortOrder == "desc" {
				direction = "DESC"
			}
			orderClause = fmt.Sprintf("%s %s", orderField, direction)
		}
	}

	// Query to get genres with pagination
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, genreColumns, whereClause, orderClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, 0, err
	}

	return genres, totalCount, nil
}

// ListSelection lấy danh sách genres rút gọn (chỉ ID và Name)
func (r *genreRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Genre, int, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")
	whereClauses = append(whereClauses, "is_active = true") // Only active genres for selection

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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.genres%s", whereClause)
	var totalCount int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Query to get genres with pagination (only ID and Name)
	query := fmt.Sprintf(`
		SELECT id, name
		FROM catalog.genres
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

	var genres []*domain.Genre
	for rows.Next() {
		var genre domain.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, 0, err
		}
		genres = append(genres, &genre)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return genres, totalCount, nil
}

// GetByParentID lấy các genre con theo parent ID
func (r *genreRepository) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*domain.Genre, error) {
	rows, err := r.pool.Query(ctx, getByParentIDQuery, parentID)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// GetRootGenres lấy các genre gốc (không có parent)
func (r *genreRepository) GetRootGenres(ctx context.Context) ([]*domain.Genre, error) {
	rows, err := r.pool.Query(ctx, getRootGenresQuery)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// Create tạo genre mới
func (r *genreRepository) Create(ctx context.Context, genre *domain.Genre) error {
	_, err := r.pool.Exec(ctx, createQuery,
		genre.ID,
		genre.Name,
		genre.Slug,
		genre.Description,
		genre.ParentID,
		genre.IsActive,
		genre.CreatedBy,
	)

	return err
}

// Update cập nhật genre
func (r *genreRepository) Update(ctx context.Context, genre *domain.Genre) error {
	result, err := r.pool.Exec(ctx, updateQuery,
		genre.ID,
		genre.Name,
		genre.Slug,
		genre.Description,
		genre.ParentID,
		genre.IsActive,
		genre.UpdatedBy,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete xóa genre (soft delete)
func (r *genreRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteQuery, id, deletedBy)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetNovelGenres lấy danh sách genres của một novel
func (r *genreRepository) GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	rows, err := r.pool.Query(ctx, getNovelGenresQuery, novelID)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// AddNovelGenre thêm genre cho novel
func (r *genreRepository) AddNovelGenre(ctx context.Context, novelID, genreID, createdBy uuid.UUID) error {
	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, addNovelGenreQuery, novelID, genreID, createdBy)
	return err
}

// AddNovelGenres thêm nhiều genres cho novel (Batch Insert)
func (r *genreRepository) AddNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error {
	if len(genreIDs) == 0 {
		return nil
	}

	// Build batch insert query
	query := "INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at) VALUES "
	var args []any
	var values []string

	args = append(args, novelID, createdBy) // $1, $2

	for i, genreID := range genreIDs {
		// Indices:
		// novelID: $1
		// createdBy: $2
		// genreID: $3 + i
		
		paramIdx := 3 + i
		values = append(values, fmt.Sprintf("(gen_random_uuid(), $1, $%d, 0, $2, NOW())", paramIdx))
		args = append(args, genreID)
	}

	query += strings.Join(values, ",")
	query += " ON CONFLICT (novel_id, genre_id) DO UPDATE SET display_order = 0"

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, args...)
	return err
}

// RemoveNovelGenre xóa genre khỏi novel
func (r *genreRepository) RemoveNovelGenre(ctx context.Context, novelID, genreID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, removeNovelGenreQuery, novelID, genreID)
	return err
}

// UpdateNovelGenres cập nhật toàn bộ genres của novel
func (r *genreRepository) UpdateNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing genres
	_, err = tx.Exec(ctx, deleteNovelGenresQuery, novelID)
	if err != nil {
		return err
	}

	// Insert new genres
	if len(genreIDs) > 0 {
		for i, genreID := range genreIDs {
			_, err = tx.Exec(ctx, insertNovelGenreQuery, novelID, genreID, i, createdBy)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

// BatchIncrementNovelCount tăng số lượng novel cho nhiều genres
func (r *genreRepository) BatchIncrementNovelCount(ctx context.Context, increments map[uuid.UUID]int) error {
	if len(increments) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for genreID, count := range increments {
		batch.Queue(incrementNovelCountQuery, genreID, count)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(increments); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch update for index %d: %w", i, err)
		}
	}

	return nil
}

// BatchIncrementTotalViews tăng total views cho nhiều genres
func (r *genreRepository) BatchIncrementTotalViews(ctx context.Context, increments map[uuid.UUID]int64) error {
	if len(increments) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for genreID, count := range increments {
		batch.Queue(incrementTotalViewsQuery, genreID, count)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(increments); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch update for index %d: %w", i, err)
		}
	}

	return nil
}

// GetGenresByNovelIDs lấy genre IDs cho danh sách novel IDs
func (r *genreRepository) GetGenresByNovelIDs(ctx context.Context, novelIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	if len(novelIDs) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, getGenresByNovelIDsQuery, novelIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]uuid.UUID)
	for rows.Next() {
		var novelID, genreID uuid.UUID
		if err := rows.Scan(&novelID, &genreID); err != nil {
			return nil, err
		}
		result[novelID] = append(result[novelID], genreID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// BatchUpdateActiveReaders cập nhật số lượng active readers cho nhiều genres
func (r *genreRepository) BatchUpdateActiveReaders(ctx context.Context, updates map[uuid.UUID]int64) error {
	if len(updates) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for genreID, count := range updates {
		batch.Queue(updateActiveReadersQuery, genreID, count)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(updates); i++ {
		_, err := br.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch update for index %d: %w", i, err)
		}
	}

	return nil
}

// Merge gộp nhiều genres (sources) thành một genre (target)
func (r *genreRepository) Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
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

	// 1. Move novels from source genres to target genre
	_, err = tx.Exec(ctx, mergeMoveNovelsQuery, targetID, sourceIDStrings)
	if err != nil {
		return err
	}

	// 2. Remove all source genre assignments
	_, err = tx.Exec(ctx, mergeRemoveSourceAssignmentsQuery, sourceIDStrings)
	if err != nil {
		return err
	}

	// 3. Update target genre stats (sum views, active_readers)
	_, err = tx.Exec(ctx, mergeUpdateTargetStatsQuery, targetID, sourceIDStrings, mergedBy)
	if err != nil {
		return err
	}

	// 4. Recalculate novel_count for target genre
	_, err = tx.Exec(ctx, mergeRecountNovelsQuery, targetID)
	if err != nil {
		return err
	}

	// 5. Soft delete source genres
	_, err = tx.Exec(ctx, mergeDeleteSourcesQuery, sourceIDStrings, mergedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
func (r *genreRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.AffectedNovel, error) {
	if len(sourceIDs) == 0 {
		return []*domain.AffectedNovel{}, nil
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

	var novels []*domain.AffectedNovel
	for rows.Next() {
		novel := &domain.AffectedNovel{}
		if err := rows.Scan(&novel.ID, &novel.Title, &novel.Slug, &novel.CoverImageURL); err != nil {
			return nil, err
		}
		novels = append(novels, novel)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return novels, nil
}
