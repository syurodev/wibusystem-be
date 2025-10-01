package repositories

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// NovelRepository interface defines methods for novel data access
type NovelRepository interface {
	CreateNovel(ctx context.Context, req d.CreateNovelRequest) (*m.Novel, error)
	GetNovelByID(ctx context.Context, id uuid.UUID) (*m.Novel, error)
	UpdateNovel(ctx context.Context, id uuid.UUID, req d.UpdateNovelRequest) (*m.Novel, error)
	DeleteNovel(ctx context.Context, id uuid.UUID, deletedByUserID uuid.UUID) error
	CheckNovelPurchases(ctx context.Context, novelID uuid.UUID) (bool, error)
	ListNovels(ctx context.Context, req d.ListNovelsRequest) (*d.PaginatedNovelsResponse, error)
	// Optional data loaders for translations and stats
	GetNovelTranslations(ctx context.Context, novelID uuid.UUID, language string) ([]interface{}, error)
	GetNovelStats(ctx context.Context, novelID uuid.UUID) (map[string]interface{}, error)
}

// novelRepository implements NovelRepository interface
type novelRepository struct {
	pool *pgxpool.Pool
}

// NewNovelRepository creates a new novel repository instance
func NewNovelRepository(pool *pgxpool.Pool) NovelRepository {
	return &novelRepository{pool: pool}
}

// CreateNovel creates a new novel and associates it with genres
func (r *novelRepository) CreateNovel(ctx context.Context, req d.CreateNovelRequest) (*m.Novel, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert novel record with all new fields
	novelID := uuid.New()
	var novel m.Novel

	// Generate slug if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlugFromName(req.Title)
	}

	query := `
		INSERT INTO novel (
			id, status, cover_image, summary,
			tenant_id, published_at, original_language, source_url, isbn,
			age_rating, content_warnings, mature_content,
			is_public, is_featured, is_completed,
			slug, tags, keywords, meta_description,
			price_coins, rental_price_coins, rental_duration_days, is_premium,
			created_at, updated_at
		)
		VALUES (
			$1, 'DRAFT', $2, $3,
			$5, $6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22, $23,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING
			id, name, status, cover_image, summary,
			tenant_id, published_at, original_language, source_url, isbn,
			age_rating, content_warnings, mature_content,
			is_public, is_featured, is_completed,
			slug, tags, keywords, meta_description,
			price_coins, rental_price_coins, rental_duration_days, is_premium,
			view_count, like_count, bookmark_count, comment_count,
			rating_average, rating_count, total_chapters, total_volumes,
			estimated_reading_time, word_count,
			created_at, updated_at
	`

	// Prepare values
	var summaryBytes, contentWarningsBytes, tagsBytes []byte
	if req.Summary != nil {
		summaryBytes = *req.Summary
	}
	if req.ContentWarnings != nil {
		contentWarningsBytes = *req.ContentWarnings
	}
	if req.Tags != nil {
		tagsBytes = *req.Tags
	}

	// Prepare nullable string pointers
	var ageRating, sourceURL, isbn, keywords, metaDescription *string
	if req.AgeRating != "" {
		ageRating = &req.AgeRating
	}
	if req.SourceURL != "" {
		sourceURL = &req.SourceURL
	}
	if req.ISBN != "" {
		isbn = &req.ISBN
	}
	if req.Keywords != "" {
		keywords = &req.Keywords
	}
	if req.MetaDescription != "" {
		metaDescription = &req.MetaDescription
	}

	err = tx.QueryRow(ctx, query,
		novelID, req.Title, req.CoverImage, summaryBytes,
		req.TenantID, req.PublishedAt, req.OriginalLanguage, sourceURL, isbn,
		ageRating, contentWarningsBytes, req.MatureContent,
		req.IsPublic, req.IsFeatured, req.IsCompleted,
		slug, tagsBytes, keywords, metaDescription,
		req.PriceCoins, req.RentalPriceCoins, req.RentalDurationDays, req.IsPremium,
	).Scan(
		&novel.ID, &novel.Name, &novel.Status, &novel.CoverImage, &novel.Summary,
		&novel.TenantID, &novel.PublishedAt, &novel.OriginalLanguage, &novel.SourceURL, &novel.ISBN,
		&novel.AgeRating, &novel.ContentWarnings, &novel.MatureContent,
		&novel.IsPublic, &novel.IsFeatured, &novel.IsCompleted,
		&novel.Slug, &novel.Tags, &novel.Keywords, &novel.MetaDescription,
		&novel.PriceCoins, &novel.RentalPriceCoins, &novel.RentalDurationDays, &novel.IsPremium,
		&novel.ViewCount, &novel.LikeCount, &novel.BookmarkCount, &novel.CommentCount,
		&novel.RatingAverage, &novel.RatingCount, &novel.TotalChapters, &novel.TotalVolumes,
		&novel.EstimatedReadingTime, &novel.WordCount,
		&novel.CreatedAt, &novel.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create novel: %w", err)
	}

	// Validate genres exist before creating associations
	if len(req.Genres) > 0 {
		genreIDs := make([]uuid.UUID, len(req.Genres))
		for i, genreStr := range req.Genres {
			genreID, err := uuid.Parse(genreStr)
			if err != nil {
				return nil, fmt.Errorf("invalid genre ID %s: %w", genreStr, err)
			}
			genreIDs[i] = genreID
		}

		// Check if all genres exist
		existingGenres := make([]uuid.UUID, 0)
		rows, err := tx.Query(ctx, "SELECT id FROM genre WHERE id = ANY($1)", genreIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to validate genres: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var genreID uuid.UUID
			if err := rows.Scan(&genreID); err != nil {
				return nil, fmt.Errorf("failed to scan genre ID: %w", err)
			}
			existingGenres = append(existingGenres, genreID)
		}

		if len(existingGenres) != len(genreIDs) {
			return nil, fmt.Errorf("some genres do not exist")
		}

		// Create novel-genre associations
		for _, genreID := range genreIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO novel_genre (id, novel_id, genre_id, created_at, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`, uuid.New(), novel.ID, genreID)
			if err != nil {
				return nil, fmt.Errorf("failed to associate novel with genre %s: %w", genreID, err)
			}
		}
	}

	// Create novel-creator associations
	if len(req.Creators) > 0 {
		for _, creator := range req.Creators {
			creatorID, err := uuid.Parse(creator.CreatorID)
			if err != nil {
				return nil, fmt.Errorf("invalid creator ID %s: %w", creator.CreatorID, err)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO novel_creator (id, novel_id, creator_id, role, created_at, updated_at)
				VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`, uuid.New(), novel.ID, creatorID, creator.Role)
			if err != nil {
				return nil, fmt.Errorf("failed to associate novel with creator %s: %w", creatorID, err)
			}
		}
	}

	// Create novel-character associations
	if len(req.Characters) > 0 {
		characterIDs := make([]uuid.UUID, len(req.Characters))
		for i, characterStr := range req.Characters {
			characterID, err := uuid.Parse(characterStr)
			if err != nil {
				return nil, fmt.Errorf("invalid character ID %s: %w", characterStr, err)
			}
			characterIDs[i] = characterID
		}

		// Check if all characters exist
		existingCharacters := make([]uuid.UUID, 0)
		rows, err := tx.Query(ctx, "SELECT id FROM character WHERE id = ANY($1)", characterIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to validate characters: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var characterID uuid.UUID
			if err := rows.Scan(&characterID); err != nil {
				return nil, fmt.Errorf("failed to scan character ID: %w", err)
			}
			existingCharacters = append(existingCharacters, characterID)
		}

		if len(existingCharacters) != len(characterIDs) {
			return nil, fmt.Errorf("some characters do not exist")
		}

		// Create novel-character associations
		for _, characterID := range characterIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO novel_character (id, novel_id, character_id, created_at, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`, uuid.New(), novel.ID, characterID)
			if err != nil {
				return nil, fmt.Errorf("failed to associate novel with character %s: %w", characterID, err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &novel, nil
}

// generateSlugFromName creates a URL-friendly slug from novel title
func generateSlugFromName(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length to 255 characters
	if len(slug) > 255 {
		slug = slug[:255]
		// Ensure we don't cut in the middle of a word
		if lastHyphen := strings.LastIndex(slug, "-"); lastHyphen > 200 {
			slug = slug[:lastHyphen]
		}
	}

	return slug
}

// GetNovelByID retrieves a novel by its ID
func (r *novelRepository) GetNovelByID(ctx context.Context, id uuid.UUID) (*m.Novel, error) {
	var novel m.Novel

	query := `
		SELECT
			id, name, status, cover_image, summary,
			created_by_user_id, updated_by_user_id, tenant_id,
			published_at, original_language, source_url, isbn,
			age_rating, content_warnings, mature_content,
			is_public, is_featured, is_completed, is_deleted,
			deleted_at, deleted_by_user_id,
			slug, tags, keywords, meta_description,
			view_count, like_count, bookmark_count, comment_count,
			rating_average, rating_count,
			price_coins, rental_price_coins, rental_duration_days, is_premium,
			total_chapters, total_volumes, estimated_reading_time, word_count,
			created_at, updated_at
		FROM novel
		WHERE id = $1 AND is_deleted = FALSE
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&novel.ID, &novel.Name, &novel.Status, &novel.CoverImage, &novel.Summary,
		&novel.CreatedByUserID, &novel.UpdatedByUserID, &novel.TenantID,
		&novel.PublishedAt, &novel.OriginalLanguage, &novel.SourceURL, &novel.ISBN,
		&novel.AgeRating, &novel.ContentWarnings, &novel.MatureContent,
		&novel.IsPublic, &novel.IsFeatured, &novel.IsCompleted, &novel.IsDeleted,
		&novel.DeletedAt, &novel.DeletedByUserID,
		&novel.Slug, &novel.Tags, &novel.Keywords, &novel.MetaDescription,
		&novel.ViewCount, &novel.LikeCount, &novel.BookmarkCount, &novel.CommentCount,
		&novel.RatingAverage, &novel.RatingCount,
		&novel.PriceCoins, &novel.RentalPriceCoins, &novel.RentalDurationDays, &novel.IsPremium,
		&novel.TotalChapters, &novel.TotalVolumes, &novel.EstimatedReadingTime, &novel.WordCount,
		&novel.CreatedAt, &novel.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("novel not found")
		}
		return nil, fmt.Errorf("failed to get novel: %w", err)
	}

	return &novel, nil
}

// ListNovels retrieves a paginated list of novels with filtering and sorting
func (r *novelRepository) ListNovels(ctx context.Context, req d.ListNovelsRequest) (*d.PaginatedNovelsResponse, error) {
	// Set pagination defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Set default sorting
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build base query with JOINs to get basic novel info and latest chapter update
	baseQuery := `
	SELECT DISTINCT
		n.id, n.name, n.cover_image, n.view_count, n.created_at,
		n.created_by_user_id, n.tenant_id,
		latest_chapter.updated_at as latest_chapter_updated_at
	FROM novel n
	LEFT JOIN novel_genre ng ON n.id = ng.novel_id
	LEFT JOIN (
		SELECT
			nc.novel_id,
			MAX(nc.updated_at) as updated_at
		FROM novel_chapter nc
		WHERE nc.is_deleted = false
		GROUP BY nc.novel_id
	) latest_chapter ON n.id = latest_chapter.novel_id
	WHERE n.is_deleted = false`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add filtering conditions
	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("n.status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.AgeRating != "" {
		conditions = append(conditions, fmt.Sprintf("n.age_rating = $%d", argIndex))
		args = append(args, req.AgeRating)
		argIndex++
	}

	if req.MatureContent != nil {
		conditions = append(conditions, fmt.Sprintf("n.mature_content = $%d", argIndex))
		args = append(args, *req.MatureContent)
		argIndex++
	}

	if req.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("n.is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if req.IsFeatured != nil {
		conditions = append(conditions, fmt.Sprintf("n.is_featured = $%d", argIndex))
		args = append(args, *req.IsFeatured)
		argIndex++
	}

	if req.IsCompleted != nil {
		conditions = append(conditions, fmt.Sprintf("n.is_completed = $%d", argIndex))
		args = append(args, *req.IsCompleted)
		argIndex++
	}

	if req.IsPremium != nil {
		conditions = append(conditions, fmt.Sprintf("n.is_premium = $%d", argIndex))
		args = append(args, *req.IsPremium)
		argIndex++
	}

	if req.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("n.created_by_user_id = $%d", argIndex))
		args = append(args, *req.UserID)
		argIndex++
	}

	if req.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("n.tenant_id = $%d", argIndex))
		args = append(args, *req.TenantID)
		argIndex++
	}

	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(n.name ILIKE $%d OR n.summary::text ILIKE $%d)", argIndex, argIndex))
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	if len(req.GenreIDs) > 0 {
		genreUUIDs := make([]uuid.UUID, len(req.GenreIDs))
		for i, genreID := range req.GenreIDs {
			if genreUUID, err := uuid.Parse(genreID); err == nil {
				genreUUIDs[i] = genreUUID
			}
		}
		if len(genreUUIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("ng.genre_id = ANY($%d)", argIndex))
			args = append(args, genreUUIDs)
			argIndex++
		}
	}

	// Date filtering
	if req.CreatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("n.created_at >= $%d", argIndex))
		args = append(args, *req.CreatedAfter)
		argIndex++
	}

	if req.CreatedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("n.created_at <= $%d", argIndex))
		args = append(args, *req.CreatedBefore)
		argIndex++
	}

	if req.PublishedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("n.published_at >= $%d", argIndex))
		args = append(args, *req.PublishedAfter)
		argIndex++
	}

	if req.PublishedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("n.published_at <= $%d", argIndex))
		args = append(args, *req.PublishedBefore)
		argIndex++
	}

	if req.LatestChapterUpdatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("latest_chapter.updated_at >= $%d", argIndex))
		args = append(args, *req.LatestChapterUpdatedAfter)
		argIndex++
	}

	if req.LatestChapterUpdatedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("latest_chapter.updated_at <= $%d", argIndex))
		args = append(args, *req.LatestChapterUpdatedBefore)
		argIndex++
	}

	if req.MinViewCount != nil {
		conditions = append(conditions, fmt.Sprintf("n.view_count >= $%d", argIndex))
		args = append(args, *req.MinViewCount)
		argIndex++
	}

	if req.MaxViewCount != nil {
		conditions = append(conditions, fmt.Sprintf("n.view_count <= $%d", argIndex))
		args = append(args, *req.MaxViewCount)
		argIndex++
	}

	// Build complete query
	completeQuery := baseQuery
	if len(conditions) > 0 {
		completeQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY
	completeQuery += fmt.Sprintf(" ORDER BY n.%s %s", req.SortBy, strings.ToUpper(req.SortOrder))

	// Add pagination
	offset := (req.Page - 1) * req.PageSize
	completeQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, req.PageSize, offset)

	// Execute query
	rows, err := r.pool.Query(ctx, completeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute novels query: %w", err)
	}
	defer rows.Close()

	var novels []d.NovelSummaryResponse
	for rows.Next() {
		var novel d.NovelSummaryResponse
		var userID, tenantID *uuid.UUID
		var latestChapterUpdatedAt *time.Time

		err := rows.Scan(
			&novel.ID, &novel.Name, &novel.CoverImage, &novel.ViewCount, &novel.CreatedAt,
			&userID, &tenantID, &latestChapterUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan novel row: %w", err)
		}

		// Note: User and Tenant info will be populated via gRPC calls in the service layer
		// For now, we'll set them to nil, they will be filled by the service
		if userID != nil {
			novel.User = &d.UserSummary{ID: userID.String()}
		}
		if tenantID != nil {
			novel.Tenant = &d.TenantSummary{ID: tenantID.String()}
		}
		novel.LatestChapterUpdatedAt = latestChapterUpdatedAt

		novels = append(novels, novel)
	}

	// Get total count for pagination
	countQuery := `
	SELECT COUNT(DISTINCT n.id)
	FROM novel n
	LEFT JOIN novel_genre ng ON n.id = ng.novel_id
	LEFT JOIN (
		SELECT
			nc.novel_id,
			MAX(nc.updated_at) as updated_at
		FROM novel_chapter nc
		WHERE nc.is_deleted = false
		GROUP BY nc.novel_id
	) latest_chapter ON n.id = latest_chapter.novel_id
	WHERE n.is_deleted = false`

	if len(conditions) > 0 {
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	var total int64
	// Remove the LIMIT and OFFSET args for count query
	countArgs := args[:len(args)-2]
	err = r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	hasNext := req.Page < totalPages
	hasPrevious := req.Page > 1

	return &d.PaginatedNovelsResponse{
		Novels: novels,
		Pagination: d.PaginationMeta{
			Page:        req.Page,
			PageSize:    req.PageSize,
			Total:       total,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
	}, nil
}

// UpdateNovel updates an existing novel
func (r *novelRepository) UpdateNovel(ctx context.Context, id uuid.UUID, req d.UpdateNovelRequest) (*m.Novel, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Build dynamic update query
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	// Always update updated_at
	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Update fields if provided
	if req.Title != nil {
		updateFields = append(updateFields, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.CoverImage != nil {
		updateFields = append(updateFields, fmt.Sprintf("cover_image = $%d", argIndex))
		args = append(args, *req.CoverImage)
		argIndex++
	}

	if req.Summary != nil {
		updateFields = append(updateFields, fmt.Sprintf("summary = $%d", argIndex))
		args = append(args, *req.Summary)
		argIndex++
	}

	if req.PublishedAt != nil {
		updateFields = append(updateFields, fmt.Sprintf("published_at = $%d", argIndex))
		args = append(args, *req.PublishedAt)
		argIndex++
	}

	if req.OriginalLanguage != nil {
		updateFields = append(updateFields, fmt.Sprintf("original_language = $%d", argIndex))
		args = append(args, *req.OriginalLanguage)
		argIndex++
	}

	if req.SourceURL != nil {
		updateFields = append(updateFields, fmt.Sprintf("source_url = $%d", argIndex))
		args = append(args, *req.SourceURL)
		argIndex++
	}

	if req.ISBN != nil {
		updateFields = append(updateFields, fmt.Sprintf("isbn = $%d", argIndex))
		args = append(args, *req.ISBN)
		argIndex++
	}

	if req.AgeRating != nil {
		updateFields = append(updateFields, fmt.Sprintf("age_rating = $%d", argIndex))
		args = append(args, *req.AgeRating)
		argIndex++
	}

	if req.ContentWarnings != nil {
		updateFields = append(updateFields, fmt.Sprintf("content_warnings = $%d", argIndex))
		args = append(args, *req.ContentWarnings)
		argIndex++
	}

	if req.MatureContent != nil {
		updateFields = append(updateFields, fmt.Sprintf("mature_content = $%d", argIndex))
		args = append(args, *req.MatureContent)
		argIndex++
	}

	if req.IsPublic != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if req.IsFeatured != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_featured = $%d", argIndex))
		args = append(args, *req.IsFeatured)
		argIndex++
	}

	if req.IsCompleted != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_completed = $%d", argIndex))
		args = append(args, *req.IsCompleted)
		argIndex++
	}

	if req.Keywords != nil {
		updateFields = append(updateFields, fmt.Sprintf("keywords = $%d", argIndex))
		args = append(args, *req.Keywords)
		argIndex++
	}

	if req.MetaDescription != nil {
		updateFields = append(updateFields, fmt.Sprintf("meta_description = $%d", argIndex))
		args = append(args, *req.MetaDescription)
		argIndex++
	}

	if req.PriceCoins != nil {
		updateFields = append(updateFields, fmt.Sprintf("price_coins = $%d", argIndex))
		args = append(args, *req.PriceCoins)
		argIndex++
	}

	if req.RentalPriceCoins != nil {
		updateFields = append(updateFields, fmt.Sprintf("rental_price_coins = $%d", argIndex))
		args = append(args, *req.RentalPriceCoins)
		argIndex++
	}

	if req.RentalDurationDays != nil {
		updateFields = append(updateFields, fmt.Sprintf("rental_duration_days = $%d", argIndex))
		args = append(args, *req.RentalDurationDays)
		argIndex++
	}

	if req.IsPremium != nil {
		updateFields = append(updateFields, fmt.Sprintf("is_premium = $%d", argIndex))
		args = append(args, *req.IsPremium)
		argIndex++
	}

	if len(updateFields) == 1 { // Only updated_at field
		return nil, fmt.Errorf("no fields to update")
	}

	// Build and execute update query
	query := fmt.Sprintf(`
		UPDATE novel
		SET %s
		WHERE id = $%d AND is_deleted = FALSE
		RETURNING
			id, name, status, cover_image, summary,
			tenant_id, published_at, original_language, source_url, isbn,
			age_rating, content_warnings, mature_content,
			is_public, is_featured, is_completed,
			slug, tags, keywords, meta_description,
			price_coins, rental_price_coins, rental_duration_days, is_premium,
			view_count, like_count, bookmark_count, comment_count,
			rating_average, rating_count, total_chapters, total_volumes,
			estimated_reading_time, word_count,
			created_at, updated_at
	`, strings.Join(updateFields, ", "), argIndex)

	args = append(args, id)

	var novel m.Novel
	err = tx.QueryRow(ctx, query, args...).Scan(
		&novel.ID, &novel.Name, &novel.Status, &novel.CoverImage, &novel.Summary,
		&novel.TenantID, &novel.PublishedAt, &novel.OriginalLanguage, &novel.SourceURL, &novel.ISBN,
		&novel.AgeRating, &novel.ContentWarnings, &novel.MatureContent,
		&novel.IsPublic, &novel.IsFeatured, &novel.IsCompleted,
		&novel.Slug, &novel.Tags, &novel.Keywords, &novel.MetaDescription,
		&novel.PriceCoins, &novel.RentalPriceCoins, &novel.RentalDurationDays, &novel.IsPremium,
		&novel.ViewCount, &novel.LikeCount, &novel.BookmarkCount, &novel.CommentCount,
		&novel.RatingAverage, &novel.RatingCount, &novel.TotalChapters, &novel.TotalVolumes,
		&novel.EstimatedReadingTime, &novel.WordCount,
		&novel.CreatedAt, &novel.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("novel not found or already deleted")
		}
		return nil, fmt.Errorf("failed to update novel: %w", err)
	}

	// Update genre associations if provided
	if len(req.Genres) > 0 {
		// First, remove existing associations
		_, err = tx.Exec(ctx, "DELETE FROM novel_genre WHERE novel_id = $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing genre associations: %w", err)
		}

		// Validate and add new genres
		genreIDs := make([]uuid.UUID, len(req.Genres))
		for i, genreStr := range req.Genres {
			genreID, err := uuid.Parse(genreStr)
			if err != nil {
				return nil, fmt.Errorf("invalid genre ID %s: %w", genreStr, err)
			}
			genreIDs[i] = genreID
		}

		// Check if all genres exist
		existingGenres := make([]uuid.UUID, 0)
		rows, err := tx.Query(ctx, "SELECT id FROM genre WHERE id = ANY($1)", genreIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to validate genres: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var genreID uuid.UUID
			if err := rows.Scan(&genreID); err != nil {
				return nil, fmt.Errorf("failed to scan genre ID: %w", err)
			}
			existingGenres = append(existingGenres, genreID)
		}

		if len(existingGenres) != len(genreIDs) {
			return nil, fmt.Errorf("some genres do not exist")
		}

		// Create new associations
		for _, genreID := range genreIDs {
			_, err = tx.Exec(ctx, `
				INSERT INTO novel_genre (id, novel_id, genre_id, created_at, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`, uuid.New(), novel.ID, genreID)
			if err != nil {
				return nil, fmt.Errorf("failed to associate novel with genre %s: %w", genreID, err)
			}
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &novel, nil
}

// CheckNovelPurchases checks if any users have purchased content from this novel
func (r *novelRepository) CheckNovelPurchases(ctx context.Context, novelID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_content_purchases ucp
			WHERE (
				-- Check novel series purchases
				(ucp.item_type = 'NOVEL_SERIES' AND ucp.item_id = $1)
				OR
				-- Check novel volume purchases
				(ucp.item_type = 'NOVEL_VOLUME' AND ucp.item_id IN (
					SELECT id FROM novel_volume WHERE novel_id = $1 AND is_deleted = FALSE
				))
				OR
				-- Check novel chapter purchases
				(ucp.item_type = 'NOVEL_CHAPTER' AND ucp.item_id IN (
					SELECT nc.id FROM novel_chapter nc
					JOIN novel_volume nv ON nc.volume_id = nv.id
					WHERE nv.novel_id = $1 AND nc.is_deleted = FALSE AND nv.is_deleted = FALSE
				))
			)
			UNION
			-- Check rental purchases
			SELECT 1 FROM user_content_rentals ucr
			WHERE (
				-- Check novel series rentals
				(ucr.item_type = 'NOVEL_SERIES' AND ucr.item_id = $1)
				OR
				-- Check novel volume rentals
				(ucr.item_type = 'NOVEL_VOLUME' AND ucr.item_id IN (
					SELECT id FROM novel_volume WHERE novel_id = $1 AND is_deleted = FALSE
				))
			)
		)
	`

	var hasPurchases bool
	err := r.pool.QueryRow(ctx, query, novelID).Scan(&hasPurchases)
	if err != nil {
		return false, fmt.Errorf("failed to check novel purchases: %w", err)
	}

	return hasPurchases, nil
}

// DeleteNovel performs soft delete on a novel after checking for purchases
func (r *novelRepository) DeleteNovel(ctx context.Context, id uuid.UUID, deletedByUserID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if novel exists and is not already deleted
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM novel WHERE id = $1 AND is_deleted = FALSE)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check novel existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("novel not found or already deleted")
	}

	// Check if any users have purchased content from this novel
	hasPurchases, err := r.CheckNovelPurchases(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check purchases: %w", err)
	}
	if hasPurchases {
		return fmt.Errorf("cannot delete novel: users have purchased content from this novel")
	}

	// Perform soft delete on the novel
	now := time.Now()
	_, err = tx.Exec(ctx, `
		UPDATE novel
		SET is_deleted = TRUE, deleted_at = $2, deleted_by_user_id = $3, updated_at = $2
		WHERE id = $1
	`, id, now, deletedByUserID)
	if err != nil {
		return fmt.Errorf("failed to delete novel: %w", err)
	}

	// Also soft delete related volumes and chapters
	_, err = tx.Exec(ctx, `
		UPDATE novel_volume
		SET is_deleted = TRUE, deleted_at = $2
		WHERE novel_id = $1 AND is_deleted = FALSE
	`, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete novel volumes: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE novel_chapter
		SET is_deleted = TRUE, deleted_at = $2
		WHERE volume_id IN (
			SELECT id FROM novel_volume WHERE novel_id = $1
		) AND is_deleted = FALSE
	`, id, now)
	if err != nil {
		return fmt.Errorf("failed to delete novel chapters: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetNovelTranslations retrieves translations for a novel
func (r *novelRepository) GetNovelTranslations(ctx context.Context, novelID uuid.UUID, language string) ([]interface{}, error) {
	query := `
		SELECT nt.id, nt.language_code, nt.title, nt.description, nt.is_primary
		FROM novel_translation nt
		WHERE nt.novel_id = $1
	`

	args := []interface{}{novelID}
	if language != "" {
		query += " AND nt.language_code = $2"
		args = append(args, language)
	}
	query += " ORDER BY nt.is_primary DESC, nt.language_code"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query novel translations: %w", err)
	}
	defer rows.Close()

	var translations []interface{}
	for rows.Next() {
		var translation struct {
			ID           string  `json:"id"`
			LanguageCode string  `json:"language_code"`
			Title        string  `json:"title"`
			Description  *string `json:"description"`
			IsPrimary    bool    `json:"is_primary"`
		}

		err := rows.Scan(
			&translation.ID,
			&translation.LanguageCode,
			&translation.Title,
			&translation.Description,
			&translation.IsPrimary,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan translation: %w", err)
		}

		translations = append(translations, translation)
	}

	return translations, nil
}

// GetNovelStats retrieves statistics for a novel
func (r *novelRepository) GetNovelStats(ctx context.Context, novelID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get basic content counts
	contentStatsQuery := `
		SELECT
			COUNT(DISTINCT nv.id) as volume_count,
			COUNT(DISTINCT nc.id) as chapter_count,
			COALESCE(SUM(nc.word_count), 0) as total_word_count,
			COALESCE(AVG(nc.word_count), 0) as avg_chapter_word_count
		FROM novel n
		LEFT JOIN novel_volume nv ON n.id = nv.novel_id AND nv.is_deleted = FALSE
		LEFT JOIN novel_chapter nc ON nv.id = nc.volume_id AND nc.is_deleted = FALSE
		WHERE n.id = $1
	`

	var volumeCount, chapterCount, totalWordCount, avgChapterWordCount int
	err := r.pool.QueryRow(ctx, contentStatsQuery, novelID).Scan(
		&volumeCount, &chapterCount, &totalWordCount, &avgChapterWordCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get content stats: %w", err)
	}

	stats["content"] = map[string]interface{}{
		"volume_count":               volumeCount,
		"chapter_count":              chapterCount,
		"total_word_count":           totalWordCount,
		"average_chapter_word_count": avgChapterWordCount,
	}

	// Get engagement stats
	engagementStatsQuery := `
		SELECT
			n.view_count,
			n.like_count,
			n.bookmark_count,
			n.comment_count,
			n.rating_average,
			n.rating_count
		FROM novel n
		WHERE n.id = $1
	`

	var viewCount, likeCount, bookmarkCount, commentCount, ratingCount int64
	var ratingAverage *float64
	err = r.pool.QueryRow(ctx, engagementStatsQuery, novelID).Scan(
		&viewCount, &likeCount, &bookmarkCount, &commentCount, &ratingAverage, &ratingCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get engagement stats: %w", err)
	}

	stats["engagement"] = map[string]interface{}{
		"view_count":     viewCount,
		"like_count":     likeCount,
		"bookmark_count": bookmarkCount,
		"comment_count":  commentCount,
		"rating_average": ratingAverage,
		"rating_count":   ratingCount,
	}

	// Get purchase stats
	purchaseStatsQuery := `
		SELECT
			COUNT(DISTINCT CASE WHEN ucp.item_type = 'NOVEL_SERIES' THEN ucp.user_id END) as series_buyers,
			COUNT(DISTINCT CASE WHEN ucp.item_type = 'NOVEL_VOLUME' THEN ucp.user_id END) as volume_buyers,
			COUNT(DISTINCT CASE WHEN ucp.item_type = 'NOVEL_CHAPTER' THEN ucp.user_id END) as chapter_buyers,
			COUNT(DISTINCT CASE WHEN ucr.item_type = 'NOVEL_SERIES' THEN ucr.user_id END) as series_renters,
			COUNT(DISTINCT CASE WHEN ucr.item_type = 'NOVEL_VOLUME' THEN ucr.user_id END) as volume_renters
		FROM novel n
		LEFT JOIN novel_volume nv ON n.id = nv.novel_id
		LEFT JOIN novel_chapter nc ON nv.id = nc.volume_id
		LEFT JOIN user_content_purchases ucp ON (
			(ucp.item_type = 'NOVEL_SERIES' AND ucp.item_id = n.id)
			OR (ucp.item_type = 'NOVEL_VOLUME' AND ucp.item_id = nv.id)
			OR (ucp.item_type = 'NOVEL_CHAPTER' AND ucp.item_id = nc.id)
		)
		LEFT JOIN user_content_rentals ucr ON (
			(ucr.item_type = 'NOVEL_SERIES' AND ucr.item_id = n.id)
			OR (ucr.item_type = 'NOVEL_VOLUME' AND ucr.item_id = nv.id)
		)
		WHERE n.id = $1
	`

	var seriesBuyers, volumeBuyers, chapterBuyers, seriesRenters, volumeRenters int
	err = r.pool.QueryRow(ctx, purchaseStatsQuery, novelID).Scan(
		&seriesBuyers, &volumeBuyers, &chapterBuyers, &seriesRenters, &volumeRenters,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase stats: %w", err)
	}

	stats["purchases"] = map[string]interface{}{
		"series_buyers":  seriesBuyers,
		"volume_buyers":  volumeBuyers,
		"chapter_buyers": chapterBuyers,
		"series_renters": seriesRenters,
		"volume_renters": volumeRenters,
		"total_buyers":   seriesBuyers + volumeBuyers + chapterBuyers,
		"total_renters":  seriesRenters + volumeRenters,
	}

	return stats, nil
}