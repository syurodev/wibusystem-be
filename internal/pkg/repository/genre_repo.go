package repository

import (
	"context"
	"fmt"
	"strings"
	"system/internal/domain"
	"system/internal/pkg/db"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	created_by, updated_by, created_at, updated_at
`

// GetByID lấy genre từ database theo ID
func (r *genreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE id = $1 AND deleted_at IS NULL
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query, id)
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
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE slug = $1 AND deleted_at IS NULL
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query, slug)
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
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query, parentID)
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
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY name ASC
	`, genreColumns)

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

// Create tạo genre mới
func (r *genreRepository) Create(ctx context.Context, genre *domain.Genre) error {
	query := `
		INSERT INTO catalog.genres (
			id, name, slug, description, parent_id,
			is_active, created_by, created_at, updated_at,
			anime_count, manga_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), 0, 0)
	`

	_, err := r.pool.Exec(ctx, query,
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
	query := `
		UPDATE catalog.genres
		SET name = $2,
		    slug = $3,
		    description = $4,
		    parent_id = $5,
		    is_active = $6,
		    updated_by = $7,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
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
	query := `
		UPDATE catalog.genres
		SET deleted_at = NOW(),
		    deleted_by = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id, deletedBy)
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
	query := `
		SELECT g.id, g.name, g.slug, g.description, g.parent_id,
		       g.is_active, g.novel_count, g.anime_count, g.manga_count, g.active_readers, g.total_views,
		       g.created_by, g.updated_by, g.created_at, g.updated_at
		FROM catalog.genres g
		INNER JOIN catalog.novel_genres ng ON g.id = ng.genre_id
		WHERE ng.novel_id = $1 AND g.deleted_at IS NULL
		ORDER BY ng.display_order ASC, g.name ASC
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

// AddNovelGenre thêm genre cho novel
func (r *genreRepository) AddNovelGenre(ctx context.Context, novelID, genreID, createdBy uuid.UUID) error {
	query := `
		INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at)
		VALUES (gen_random_uuid(), $1, $2, 0, $3, NOW())
		ON CONFLICT (novel_id, genre_id) DO UPDATE
		SET display_order = EXCLUDED.display_order
	`

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, query, novelID, genreID, createdBy)
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
	query := `
		DELETE FROM catalog.novel_genres
		WHERE novel_id = $1 AND genre_id = $2
	`

	_, err := r.pool.Exec(ctx, query, novelID, genreID)
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
	_, err = tx.Exec(ctx, "DELETE FROM catalog.novel_genres WHERE novel_id = $1", novelID)
	if err != nil {
		return err
	}

	// Insert new genres
	if len(genreIDs) > 0 {
		for i, genreID := range genreIDs {
			_, err = tx.Exec(ctx,
				"INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at) VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())",
				novelID, genreID, i, createdBy,
			)
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
		query := `
			UPDATE catalog.genres
			SET novel_count = novel_count + $2,
				updated_at = NOW()
			WHERE id = $1
		`
		batch.Queue(query, genreID, count)
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
		query := `
			UPDATE catalog.genres
			SET total_views = total_views + $2
			WHERE id = $1
		`
		batch.Queue(query, genreID, count)
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

	query := `
		SELECT novel_id, genre_id
		FROM catalog.novel_genres
		WHERE novel_id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, novelIDs)
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
		query := `
			UPDATE catalog.genres
			SET active_readers = $2
			WHERE id = $1
		`
		batch.Queue(query, genreID, count)
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
	// Only move if the novel doesn't already have the target genre to avoid unique constraint violation
	queryMove := `
		UPDATE catalog.novel_genres
		SET genre_id = $1
		WHERE genre_id = ANY($2::uuid[])
		AND novel_id NOT IN (
			SELECT novel_id FROM catalog.novel_genres WHERE genre_id = $1
		)
	`
	_, err = tx.Exec(ctx, queryMove, targetID, sourceIDStrings)
	if err != nil {
		return err
	}

	// 2. Remove all source genre assignments
	// This removes the "moved" records (which are no longer matching the WHERE genre_id = ANY($sourceIDs) IF the update worked?)
	// WAIT: UPDATE changes the genre_id to targetID. So those records NO LONGER have genre_id in sourceIDs.
	// The records that REMAIN with genre_id in sourceIDs are the ones skipped by the NOT IN clause (duplicates).
	// So we can safely delete them.
	queryRemove := `
		DELETE FROM catalog.novel_genres
		WHERE genre_id = ANY($1::uuid[])
	`
	_, err = tx.Exec(ctx, queryRemove, sourceIDStrings)
	if err != nil {
		return err
	}

	// 3. Update target genre stats (sum views, active_readers)
	// We sum up the stats from source genres and add to target
	queryUpdateStats := `
		UPDATE catalog.genres
		SET total_views = total_views + (
				SELECT COALESCE(SUM(total_views), 0) FROM catalog.genres WHERE id = ANY($2::uuid[])
			),
			active_readers = active_readers + (
				SELECT COALESCE(SUM(active_readers), 0) FROM catalog.genres WHERE id = ANY($2::uuid[])
			),
			updated_by = $3,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, queryUpdateStats, targetID, sourceIDStrings, mergedBy)
	if err != nil {
		return err
	}

	// 4. Recalculate novel_count for target genre
	// Because merge might have caused overlaps involving the same novel, direct sum might double count.
	// So we perform a fresh count.
	queryRecount := `
		UPDATE catalog.genres
		SET novel_count = (SELECT COUNT(*) FROM catalog.novel_genres WHERE genre_id = $1)
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, queryRecount, targetID)
	if err != nil {
		return err
	}

	// 5. Soft delete source genres
	queryDelete := `
		UPDATE catalog.genres
		SET deleted_at = NOW(),
			deleted_by = $2,
			is_active = false
		WHERE id = ANY($1::uuid[])
	`
	_, err = tx.Exec(ctx, queryDelete, sourceIDStrings, mergedBy)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

	// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
func (r *genreRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	if len(sourceIDs) == 0 {
		return []*domain.Novel{}, nil
	}

	query := `
		SELECT DISTINCT n.id, n.title, n.slug, n.cover_image_url
		FROM catalog.novels n
		JOIN catalog.novel_genres ng ON n.id = ng.novel_id
		WHERE ng.genre_id = ANY($1::uuid[])
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
