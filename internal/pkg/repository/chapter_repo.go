package repository

import (
	"context"
	"fmt"
	"strings"
	"system/internal/domain"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// chapterRepository triển khai ChapterRepository sử dụng pgx
type chapterRepository struct {
	pool *pgxpool.Pool
}

// NewChapterRepository tạo một instance mới của chapterRepository
func NewChapterRepository(pool *pgxpool.Pool) domain.ChapterRepository {
	return &chapterRepository{pool: pool}
}

// GetByID lấy chapter từ database theo ID
func (r *chapterRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error) {
	query := `
		SELECT id, novel_id, volume_id, chapter_number, title, slug, content,
		       word_count, character_count, is_free, price, currency, status,
		       view_count, like_count, comment_count, display_order, author_notes,
		       published_at, scheduled_at, created_at, updated_at, deleted_at,
		       created_by, updated_by
		FROM catalog.novel_chapters
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	chapter, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Chapter])
	if err != nil {
		return nil, err
	}

	return &chapter, nil
}

// GetByNovelIDAndNumber lấy chapter theo novel ID và chapter number
func (r *chapterRepository) GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, chapterNumber int) (*domain.Chapter, error) {
	query := `
		SELECT id, novel_id, volume_id, chapter_number, title, slug, content,
		       word_count, character_count, is_free, price, currency, status,
		       view_count, like_count, comment_count, display_order, author_notes,
		       published_at, scheduled_at, created_at, updated_at, deleted_at,
		       created_by, updated_by
		FROM catalog.novel_chapters
		WHERE novel_id = $1 AND chapter_number = $2 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, novelID, chapterNumber)
	if err != nil {
		return nil, err
	}

	chapter, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Chapter])
	if err != nil {
		return nil, err
	}

	return &chapter, nil
}

// GetByNovelID lấy danh sách chapter theo novel ID
func (r *chapterRepository) GetByNovelID(ctx context.Context, novelID uuid.UUID, filter domain.ChapterFilter) ([]*domain.Chapter, error) {
	var whereClauses []string
	var args []any
	argIdx := 2

	whereClauses = append(whereClauses, "novel_id = $1")
	whereClauses = append(whereClauses, "deleted_at IS NULL")
	args = append(args, novelID)

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	if filter.VolumeID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("volume_id = $%d", argIdx))
		args = append(args, *filter.VolumeID)
		argIdx++
	}

	if filter.PublishedOnly {
		whereClauses = append(whereClauses, "status = 'published'")
	}

	if filter.IsFree != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_free = $%d", argIdx))
		args = append(args, *filter.IsFree)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Build ORDER BY clause
	orderBy := "display_order ASC"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "chapter_number":
			orderBy = "chapter_number"
		case "published_at":
			orderBy = "published_at"
		case "views":
			orderBy = "view_count"
		default:
			orderBy = filter.SortBy
		}

		if filter.SortOrder == "desc" {
			orderBy += " DESC"
		} else {
			orderBy += " ASC"
		}
	}

	query := fmt.Sprintf(`
		SELECT id, novel_id, volume_id, chapter_number, title, slug, content,
		       word_count, character_count, is_free, price, currency, status,
		       view_count, like_count, comment_count, display_order, author_notes,
		       published_at, scheduled_at, created_at, updated_at, deleted_at,
		       created_by, updated_by
		FROM catalog.novel_chapters
		WHERE %s
		ORDER BY %s
	`, whereClause, orderBy)

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	chapters, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Chapter])
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

// GetByVolumeID lấy danh sách chapter theo volume ID
func (r *chapterRepository) GetByVolumeID(ctx context.Context, volumeID uuid.UUID, publishedOnly bool) ([]*domain.Chapter, error) {
	query := `
		SELECT id, novel_id, volume_id, chapter_number, title, slug, content,
		       word_count, character_count, is_free, price, currency, status,
		       view_count, like_count, comment_count, display_order, author_notes,
		       published_at, scheduled_at, created_at, updated_at, deleted_at,
		       created_by, updated_by
		FROM catalog.novel_chapters
		WHERE volume_id = $1 AND deleted_at IS NULL
	`

	if publishedOnly {
		query += " AND status = 'published'"
	}

	query += " ORDER BY display_order ASC, chapter_number ASC"

	rows, err := r.pool.Query(ctx, query, volumeID)
	if err != nil {
		return nil, err
	}

	chapters, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Chapter])
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

// Create tạo chapter mới trong database
func (r *chapterRepository) Create(ctx context.Context, chapter *domain.Chapter) error {
	query := `
		INSERT INTO catalog.novel_chapters (
			id, novel_id, volume_id, chapter_number, title, slug, content,
			word_count, character_count, is_free, price, currency, status,
			display_order, author_notes, scheduled_at, created_by, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := r.pool.Exec(ctx, query,
		chapter.ID,
		chapter.NovelID,
		chapter.VolumeID,
		chapter.ChapterNumber,
		chapter.Title,
		chapter.Slug,
		chapter.Content,
		chapter.WordCount,
		chapter.CharacterCount,
		chapter.IsFree,
		chapter.Price,
		chapter.Currency,
		chapter.Status,
		chapter.DisplayOrder,
		chapter.AuthorNotes,
		chapter.ScheduledAt,
		chapter.CreatedBy,
		chapter.UpdatedBy,
	)

	return err
}

// Update cập nhật thông tin chapter
func (r *chapterRepository) Update(ctx context.Context, chapter *domain.Chapter) error {
	query := `
		UPDATE catalog.novel_chapters
		SET volume_id = $2,
		    chapter_number = $3,
		    title = $4,
		    slug = $5,
		    content = $6,
		    word_count = $7,
		    character_count = $8,
		    is_free = $9,
		    price = $10,
		    currency = $11,
		    status = $12,
		    display_order = $13,
		    author_notes = $14,
		    scheduled_at = $15,
		    updated_by = $16
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		chapter.ID,
		chapter.VolumeID,
		chapter.ChapterNumber,
		chapter.Title,
		chapter.Slug,
		chapter.Content,
		chapter.WordCount,
		chapter.CharacterCount,
		chapter.IsFree,
		chapter.Price,
		chapter.Currency,
		chapter.Status,
		chapter.DisplayOrder,
		chapter.AuthorNotes,
		chapter.ScheduledAt,
		chapter.UpdatedBy,
	)

	return err
}

// Delete xóa mềm chapter
func (r *chapterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novel_chapters
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Publish xuất bản chapter ngay lập tức
func (r *chapterRepository) Publish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novel_chapters
		SET status = 'published',
		    published_at = COALESCE(published_at, NOW()),
		    scheduled_at = NULL
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Schedule đặt lịch xuất bản chapter
func (r *chapterRepository) Schedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time) error {
	query := `
		UPDATE catalog.novel_chapters
		SET status = 'scheduled',
		    scheduled_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id, scheduledAt)
	return err
}

// GetScheduledChapters lấy danh sách chapter cần xuất bản
func (r *chapterRepository) GetScheduledChapters(ctx context.Context, before time.Time) ([]*domain.Chapter, error) {
	query := `
		SELECT id, novel_id, volume_id, chapter_number, title, slug, content,
		       word_count, character_count, is_free, price, currency, status,
		       view_count, like_count, comment_count, display_order, author_notes,
		       published_at, scheduled_at, created_at, updated_at, deleted_at,
		       created_by, updated_by
		FROM catalog.novel_chapters
		WHERE status = 'scheduled'
		  AND scheduled_at <= $1
		  AND deleted_at IS NULL
		ORDER BY scheduled_at ASC
	`

	rows, err := r.pool.Query(ctx, query, before)
	if err != nil {
		return nil, err
	}

	chapters, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Chapter])
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

// IncrementViewCount tăng view count
func (r *chapterRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novel_chapters
		SET view_count = view_count + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpdateStatistics cập nhật thống kê của chapter
func (r *chapterRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.ChapterStatistics) error {
	var setClauses []string
	var args []any
	argIdx := 2

	args = append(args, id) // $1 là id

	if stats.ViewCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("view_count = $%d", argIdx))
		args = append(args, *stats.ViewCount)
		argIdx++
	}

	if stats.LikeCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("like_count = $%d", argIdx))
		args = append(args, *stats.LikeCount)
		argIdx++
	}

	if stats.CommentCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("comment_count = $%d", argIdx))
		args = append(args, *stats.CommentCount)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil // Không có gì để update
	}

	query := fmt.Sprintf(`
		UPDATE catalog.novel_chapters
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// BatchIncrementViewCount tăng view count cho nhiều chapters cùng lúc.
// Sử dụng bulk UPDATE với VALUES pattern để tối ưu performance.
//
// SQL Pattern:
//
//	UPDATE catalog.novel_chapters AS c
//	SET view_count = c.view_count + v.increment
//	FROM (VALUES (uuid1, 15), (uuid2, 20)) AS v(id, increment)
//	WHERE c.id = v.id AND c.deleted_at IS NULL
//
// Parameters:
//   - ctx: Context
//   - increments: Map từ chapter ID -> increment amount
//
// Returns:
//   - error: Lỗi nếu có
func (r *chapterRepository) BatchIncrementViewCount(ctx context.Context, increments map[uuid.UUID]int64) error {
	if len(increments) == 0 {
		return nil
	}

	// Build VALUES clause cho bulk update
	var values []string
	var args []any
	argIdx := 1

	for id, count := range increments {
		values = append(values, fmt.Sprintf("($%d::uuid, $%d::bigint)", argIdx, argIdx+1))
		args = append(args, id, count)
		argIdx += 2
	}

	query := fmt.Sprintf(`
		UPDATE catalog.novel_chapters AS c
		SET view_count = c.view_count + v.increment
		FROM (VALUES %s) AS v(id, increment)
		WHERE c.id = v.id AND c.deleted_at IS NULL
	`, strings.Join(values, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
