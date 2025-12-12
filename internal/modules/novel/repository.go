package novel

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

// SQL queries embedded từ file
//
//go:embed queries/get_by_id.sql
var getByIDQuery string

//go:embed queries/get_by_slug.sql
var getBySlugQuery string

//go:embed queries/get_by_author_id.sql
var getByAuthorIDQuery string

//go:embed queries/create.sql
var createQuery string

//go:embed queries/update.sql
var updateQuery string

//go:embed queries/delete.sql
var deleteQuery string

//go:embed queries/increment_view_count.sql
var incrementViewCountQuery string

//go:embed queries/get_authors.sql
var getAuthorsQuery string

//go:embed queries/get_genres.sql
var getGenresQuery string

//go:embed queries/get_artists.sql
var getArtistsQuery string

//go:embed queries/get_organization_assignments.sql
var getOrganizationAssignmentsQuery string

//go:embed queries/load_novel_genres.sql
var loadNovelGenresQuery string

//go:embed queries/get_novel_full.sql
var getNovelFullQuery string

// novelRepository triển khai NovelRepository sử dụng pgx
type novelRepository struct {
	pool *pgxpool.Pool
}

// NewNovelRepository tạo một instance mới của novelRepository
func NewNovelRepository(pool *pgxpool.Pool) domain.NovelRepository {
	return &novelRepository{pool: pool}
}

// GetByID lấy novel từ database theo ID
func (r *novelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	rows, err := r.pool.Query(ctx, getByIDQuery, id)
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
	rows, err := r.pool.Query(ctx, getBySlugQuery, slug)
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
	rows, err := r.pool.Query(ctx, getByAuthorIDQuery, authorID, limit, offset)
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
	// Đảm bảo metadata không null
	if novel.Metadata == nil {
		novel.Metadata = json.RawMessage("{}")
	}

	// Đảm bảo synopsis không null
	if novel.Synopsis == nil {
		novel.Synopsis = json.RawMessage("{}")
	}

	db := db.GetDB(ctx, r.pool)
	_, err := db.Exec(ctx, createQuery,
		novel.ID,
		novel.Title,
		novel.Slug,
		novel.Synopsis,
		novel.CoverImageURL,
		novel.ThumbnailURL,
		novel.Status,
		novel.IsOneshot,
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
	_, err := r.pool.Exec(ctx, updateQuery,
		novel.ID,
		novel.Title,
		novel.Slug,
		novel.Synopsis,
		novel.CoverImageURL,
		novel.ThumbnailURL,
		novel.Status,
		novel.IsOneshot,
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
	_, err := r.pool.Exec(ctx, deleteQuery, id)
	return err
}

// List lấy danh sách novel với filter và pagination
func (r *novelRepository) List(ctx context.Context, filter domain.NovelFilter) ([]*domain.Novel, int64, error) {
	// Build WHERE clause động
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "n.deleted_at IS NULL")

	// Filter by specific IDs
	if len(filter.IDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("n.id = ANY($%d::uuid[])", argIdx))
		// Convert []uuid.UUID to []string for pgx text array encoding
		idStrings := make([]string, len(filter.IDs))
		for i, id := range filter.IDs {
			idStrings[i] = id.String()
		}
		args = append(args, idStrings)
		argIdx++
	}

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
	if filter.AuthorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("EXISTS (SELECT 1 FROM catalog.novel_authors na WHERE na.novel_id = n.id AND na.author_id = $%d)", argIdx))
		args = append(args, *filter.AuthorID)
		argIdx++
	}

	// Filter by translator ID (via junction table)
	if filter.TranslatorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("EXISTS (SELECT 1 FROM catalog.novel_translators nt WHERE nt.novel_id = n.id AND nt.translator_id = $%d)", argIdx))
		args = append(args, *filter.TranslatorID)
		argIdx++
	}

	// Filter by genre IDs (via junction table)
	if len(filter.GenreIDs) > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("EXISTS (SELECT 1 FROM catalog.novel_genres ng WHERE ng.novel_id = n.id AND ng.genre_id = ANY($%d::uuid[]))", argIdx))
		// Convert []uuid.UUID to []string for pgx text array encoding
		genreIDStrings := make([]string, len(filter.GenreIDs))
		for i, id := range filter.GenreIDs {
			genreIDStrings[i] = id.String()
		}
		args = append(args, genreIDStrings)
		argIdx++
	}

	joinClause := ""
	// joinClause is now empty as we moved filter joins to EXISTS clauses
	// If there are other joins needed in future, we can add them here


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
		       n.status, n.is_oneshot, n.original_language, n.original_title,
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
	_, err := r.pool.Exec(ctx, incrementViewCountQuery, id)
	return err
}

// BatchIncrementViewCount tăng view count cho nhiều novels cùng lúc.
// Sử dụng bulk UPDATE với VALUES pattern để tối ưu performance.
//
// SQL Pattern:
//
//	UPDATE catalog.novels AS n
//	SET view_count = n.view_count + v.increment, updated_at = NOW()
//	FROM (VALUES (uuid1, 15), (uuid2, 20)) AS v(id, increment)
//	WHERE n.id = v.id AND n.deleted_at IS NULL
//
// Parameters:
//   - ctx: Context
//   - increments: Map từ novel ID -> increment amount
//
// Returns:
//   - error: Lỗi nếu có
func (r *novelRepository) BatchIncrementViewCount(ctx context.Context, increments map[uuid.UUID]int64) error {
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
		UPDATE catalog.novels AS n
		SET view_count = n.view_count + v.increment,
		    updated_at = NOW()
		FROM (VALUES %s) AS v(id, increment)
		WHERE n.id = v.id AND n.deleted_at IS NULL
	`, strings.Join(values, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// GetAuthors lấy danh sách author IDs của novel
func (r *novelRepository) GetAuthors(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, getAuthorsQuery, novelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
}

// GetGenres lấy danh sách genre IDs của novel
func (r *novelRepository) GetGenres(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, getGenresQuery, novelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
}

// GetArtists lấy danh sách artist IDs của novel
func (r *novelRepository) GetArtists(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, getArtistsQuery, novelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
}

// GetOrganizationAssignments lấy danh sách organization IDs được assign cho novel
func (r *novelRepository) GetOrganizationAssignments(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, getOrganizationAssignmentsQuery, novelID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
}

// loadNovelGenres loads genres for a specific novel
func (r *novelRepository) loadNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	rows, err := r.pool.Query(ctx, loadNovelGenresQuery, novelID)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// NovelFullData chứa toàn bộ dữ liệu cần thiết cho trang chi tiết novel
type NovelFullData struct {
	Novel              *domain.Novel
	Genres             []*domain.Genre
	Authors            []*domain.NovelAuthor
	Artists            []*domain.NovelArtist
	Volumes            []*domain.NovelVolume
	Chapters           []*domain.NovelChapter           // All published chapters
	ChaptersWithoutVol []*domain.NovelChapter           // Chapters không thuộc volume nào
	VolumesWithChapters []*domain.NovelVolumeWithChapters
}

// GetNovelFullBySlug lấy toàn bộ dữ liệu novel trong một transaction
// Sử dụng single query với JSON aggregation để giảm round-trips
// Query được load từ queries/get_novel_full.sql
func (r *novelRepository) GetNovelFullBySlug(ctx context.Context, slug string) (*NovelFullData, error) {
	var novel domain.Novel
	var genresJSON, authorsJSON, artistsJSON, volumesJSON, chaptersJSON json.RawMessage

	row := r.pool.QueryRow(ctx, getNovelFullQuery, slug)
	err := row.Scan(
		&novel.ID, &novel.Title, &novel.Slug, &novel.Synopsis, &novel.CoverImageURL, &novel.ThumbnailURL,
		&novel.Status, &novel.IsOneshot, &novel.OriginalLanguage, &novel.OriginalTitle,
		&novel.OwnerID, &novel.OwnerType,
		&novel.OwnerDisplayName, &novel.OwnerUsername, &novel.OwnerAvatarURL,
		&novel.TotalVolumes, &novel.TotalChapters, &novel.TotalWords, &novel.ViewCount,
		&novel.FavoriteCount, &novel.RatingAverage, &novel.RatingCount, &novel.Metadata,
		&novel.FirstPublishedAt, &novel.LastChapterAt, &novel.CompletedAt,
		&novel.CreatedBy, &novel.UpdatedBy, &novel.DeletedBy,
		&novel.CreatedAt, &novel.UpdatedAt, &novel.DeletedAt,
		&genresJSON, &authorsJSON, &artistsJSON, &volumesJSON, &chaptersJSON,
	)
	if err != nil {
		return nil, err
	}

	result := &NovelFullData{
		Novel:               &novel,
		Genres:              make([]*domain.Genre, 0),
		Authors:             make([]*domain.NovelAuthor, 0),
		Artists:             make([]*domain.NovelArtist, 0),
		Volumes:             make([]*domain.NovelVolume, 0),
		Chapters:            make([]*domain.NovelChapter, 0),
		ChaptersWithoutVol:  make([]*domain.NovelChapter, 0),
		VolumesWithChapters: make([]*domain.NovelVolumeWithChapters, 0),
	}

	// Parse genres
	type genreJSON struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
		Slug string    `json:"slug"`
	}
	var genres []genreJSON
	if err := json.Unmarshal(genresJSON, &genres); err == nil {
		for _, g := range genres {
			result.Genres = append(result.Genres, &domain.Genre{ID: g.ID, Name: g.Name, Slug: g.Slug})
		}
	}

	// Parse authors
	type authorJSON struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	var authors []authorJSON
	if err := json.Unmarshal(authorsJSON, &authors); err == nil {
		for _, a := range authors {
			result.Authors = append(result.Authors, &domain.NovelAuthor{
				Author: &domain.Author{ID: a.ID, Name: a.Name},
			})
		}
	}

	// Parse artists
	type artistJSON struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}
	var artists []artistJSON
	if err := json.Unmarshal(artistsJSON, &artists); err == nil {
		for _, a := range artists {
			result.Artists = append(result.Artists, &domain.NovelArtist{
				Artist: &domain.Artist{ID: a.ID, Name: a.Name},
			})
		}
	}

	// Parse volumes
	type volumeJSON struct {
		ID            uuid.UUID  `json:"id"`
		VolumeNumber  int        `json:"volume_number"`
		Title         string     `json:"title"`
		Slug          string     `json:"slug"`
		CoverImageURL *string    `json:"cover_image_url"`
		DisplayOrder  int        `json:"display_order"`
		IsPublished   bool       `json:"is_published"`
		PublishedAt   *string    `json:"published_at"`
	}
	var volumes []volumeJSON
	if err := json.Unmarshal(volumesJSON, &volumes); err == nil {
		for _, v := range volumes {
			vol := &domain.NovelVolume{
				ID:            v.ID,
				VolumeNumber:  v.VolumeNumber,
				Title:         v.Title,
				Slug:          v.Slug,
				CoverImageURL: v.CoverImageURL,
				DisplayOrder:  v.DisplayOrder,
				IsPublished:   v.IsPublished,
			}
			result.Volumes = append(result.Volumes, vol)
		}
	}

	// Parse chapters
	type chapterJSON struct {
		ID            uuid.UUID  `json:"id"`
		VolumeID      *uuid.UUID `json:"volume_id"`
		ChapterNumber int        `json:"chapter_number"`
		Title         string     `json:"title"`
		Slug          string     `json:"slug"`
		DisplayOrder  int        `json:"display_order"`
		Status        string     `json:"status"`
		PublishedAt   *string    `json:"published_at"`
	}
	var chapters []chapterJSON
	if err := json.Unmarshal(chaptersJSON, &chapters); err == nil {
		for _, c := range chapters {
			ch := &domain.NovelChapter{
				ID:            c.ID,
				VolumeID:      c.VolumeID,
				ChapterNumber: c.ChapterNumber,
				Title:         c.Title,
				Slug:          c.Slug,
				DisplayOrder:  c.DisplayOrder,
				Status:        domain.NovelChapterStatus(c.Status),
			}
			result.Chapters = append(result.Chapters, ch)
			if c.VolumeID == nil {
				result.ChaptersWithoutVol = append(result.ChaptersWithoutVol, ch)
			}
		}
	}

	// Build VolumesWithChapters
	volumeChaptersMap := make(map[uuid.UUID][]*domain.NovelChapter)
	for _, ch := range result.Chapters {
		if ch.VolumeID != nil {
			volumeChaptersMap[*ch.VolumeID] = append(volumeChaptersMap[*ch.VolumeID], ch)
		}
	}
	for _, vol := range result.Volumes {
		result.VolumesWithChapters = append(result.VolumesWithChapters, &domain.NovelVolumeWithChapters{
			Volume:   vol,
			Chapters: volumeChaptersMap[vol.ID],
		})
	}

	return result, nil
}
