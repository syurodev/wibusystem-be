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

// chapterTranslationRepository triển khai ChapterTranslationRepository sử dụng pgx
type chapterTranslationRepository struct {
	pool *pgxpool.Pool
}

// NewChapterTranslationRepository tạo một instance mới của chapterTranslationRepository
func NewChapterTranslationRepository(pool *pgxpool.Pool) domain.ChapterTranslationRepository {
	return &chapterTranslationRepository{pool: pool}
}

// GetByID lấy translation từ database theo ID
func (r *chapterTranslationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ChapterTranslation, error) {
	query := `
		SELECT id, chapter_id, language, title, content, translator_notes,
		       translator_id, version, status, word_count, character_count,
		       view_count, like_count, rating_average, rating_count,
		       published_at, created_at, updated_at, deleted_at
		FROM catalog.chapter_translations
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	translation, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.ChapterTranslation])
	if err != nil {
		return nil, err
	}

	return &translation, nil
}

// GetByChapterAndLanguage lấy translation theo chapter và language
func (r *chapterTranslationRepository) GetByChapterAndLanguage(ctx context.Context, chapterID uuid.UUID, language string) (*domain.ChapterTranslation, error) {
	query := `
		SELECT id, chapter_id, language, title, content, translator_notes,
		       translator_id, version, status, word_count, character_count,
		       view_count, like_count, rating_average, rating_count,
		       published_at, created_at, updated_at, deleted_at
		FROM catalog.chapter_translations
		WHERE chapter_id = $1 AND language = $2 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, chapterID, language)
	if err != nil {
		return nil, err
	}

	translation, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.ChapterTranslation])
	if err != nil {
		return nil, err
	}

	return &translation, nil
}

// GetByChapterID lấy tất cả translations của một chapter
func (r *chapterTranslationRepository) GetByChapterID(ctx context.Context, chapterID uuid.UUID) ([]*domain.ChapterTranslation, error) {
	query := `
		SELECT id, chapter_id, language, title, content, translator_notes,
		       translator_id, version, status, word_count, character_count,
		       view_count, like_count, rating_average, rating_count,
		       published_at, created_at, updated_at, deleted_at
		FROM catalog.chapter_translations
		WHERE chapter_id = $1 AND deleted_at IS NULL
		ORDER BY language ASC
	`

	rows, err := r.pool.Query(ctx, query, chapterID)
	if err != nil {
		return nil, err
	}

	translations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.ChapterTranslation])
	if err != nil {
		return nil, err
	}

	return translations, nil
}

// GetByTranslatorID lấy danh sách translations của translator
func (r *chapterTranslationRepository) GetByTranslatorID(ctx context.Context, translatorID uuid.UUID, filter domain.TranslationFilter) ([]*domain.ChapterTranslation, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 2

	whereClauses = append(whereClauses, "translator_id = $1")
	whereClauses = append(whereClauses, "deleted_at IS NULL")
	args = append(args, translatorID)

	if filter.Language != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, *filter.Language)
		argIdx++
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

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

	query := fmt.Sprintf(`
		SELECT id, chapter_id, language, title, content, translator_notes,
		       translator_id, version, status, word_count, character_count,
		       view_count, like_count, rating_average, rating_count,
		       published_at, created_at, updated_at, deleted_at
		FROM catalog.chapter_translations
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

	translations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.ChapterTranslation])
	if err != nil {
		return nil, err
	}

	return translations, nil
}

// Create tạo translation mới
func (r *chapterTranslationRepository) Create(ctx context.Context, translation *domain.ChapterTranslation) error {
	query := `
		INSERT INTO catalog.chapter_translations (
			id, chapter_id, language, title, content, translator_notes,
			translator_id, version, status, word_count, character_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(ctx, query,
		translation.ID,
		translation.ChapterID,
		translation.Language,
		translation.Title,
		translation.Content,
		translation.TranslatorNotes,
		translation.TranslatorID,
		translation.Version,
		translation.Status,
		translation.WordCount,
		translation.CharacterCount,
	)

	return err
}

// Update cập nhật translation
func (r *chapterTranslationRepository) Update(ctx context.Context, translation *domain.ChapterTranslation) error {
	query := `
		UPDATE catalog.chapter_translations
		SET language = $2,
		    title = $3,
		    content = $4,
		    translator_notes = $5,
		    translator_id = $6,
		    status = $7,
		    word_count = $8,
		    character_count = $9
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		translation.ID,
		translation.Language,
		translation.Title,
		translation.Content,
		translation.TranslatorNotes,
		translation.TranslatorID,
		translation.Status,
		translation.WordCount,
		translation.CharacterCount,
	)

	return err
}

// Delete xóa mềm translation
func (r *chapterTranslationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.chapter_translations
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Publish xuất bản translation
func (r *chapterTranslationRepository) Publish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.chapter_translations
		SET status = 'published',
		    published_at = COALESCE(published_at, NOW())
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// IncrementViewCount tăng view count
func (r *chapterTranslationRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.chapter_translations
		SET view_count = view_count + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpdateStatistics cập nhật thống kê
func (r *chapterTranslationRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.TranslationStatisticsUpdate) error {
	var setClauses []string
	var args []interface{}
	argIdx := 2

	args = append(args, id)

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
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE catalog.chapter_translations
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// translationContributionRepository triển khai TranslationContributionRepository
type translationContributionRepository struct {
	pool *pgxpool.Pool
}

// NewTranslationContributionRepository tạo một instance mới
func NewTranslationContributionRepository(pool *pgxpool.Pool) domain.TranslationContributionRepository {
	return &translationContributionRepository{pool: pool}
}

// GetByID lấy contribution từ database theo ID
func (r *translationContributionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TranslationContribution, error) {
	query := `
		SELECT id, chapter_id, contributor_id, language, contribution_type,
		       title, content, contributor_notes, status,
		       reviewed_by, reviewed_at, review_notes,
		       official_translation_id, credit_points, is_credited,
		       word_count, character_count, upvote_count, downvote_count,
		       created_at, updated_at, deleted_at
		FROM catalog.translation_contributions
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	contribution, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.TranslationContribution])
	if err != nil {
		return nil, err
	}

	return &contribution, nil
}

// GetByChapterID lấy danh sách contributions của một chapter
func (r *translationContributionRepository) GetByChapterID(ctx context.Context, chapterID uuid.UUID, filter domain.ContributionFilter) ([]*domain.TranslationContribution, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 2

	whereClauses = append(whereClauses, "chapter_id = $1")
	whereClauses = append(whereClauses, "deleted_at IS NULL")
	args = append(args, chapterID)

	if filter.Language != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, *filter.Language)
		argIdx++
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	if filter.ContributionType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("contribution_type = $%d", argIdx))
		args = append(args, *filter.ContributionType)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT id, chapter_id, contributor_id, language, contribution_type,
		       title, content, contributor_notes, status,
		       reviewed_by, reviewed_at, review_notes,
		       official_translation_id, credit_points, is_credited,
		       word_count, character_count, upvote_count, downvote_count,
		       created_at, updated_at, deleted_at
		FROM catalog.translation_contributions
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	contributions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.TranslationContribution])
	if err != nil {
		return nil, err
	}

	return contributions, nil
}

// GetByContributorID lấy danh sách contributions của contributor
func (r *translationContributionRepository) GetByContributorID(ctx context.Context, contributorID uuid.UUID, filter domain.ContributionFilter) ([]*domain.TranslationContribution, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 2

	whereClauses = append(whereClauses, "contributor_id = $1")
	whereClauses = append(whereClauses, "deleted_at IS NULL")
	args = append(args, contributorID)

	if filter.Language != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("language = $%d", argIdx))
		args = append(args, *filter.Language)
		argIdx++
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT id, chapter_id, contributor_id, language, contribution_type,
		       title, content, contributor_notes, status,
		       reviewed_by, reviewed_at, review_notes,
		       official_translation_id, credit_points, is_credited,
		       word_count, character_count, upvote_count, downvote_count,
		       created_at, updated_at, deleted_at
		FROM catalog.translation_contributions
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	contributions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.TranslationContribution])
	if err != nil {
		return nil, err
	}

	return contributions, nil
}

// GetPendingReview lấy danh sách contributions đang chờ review
func (r *translationContributionRepository) GetPendingReview(ctx context.Context, language *string, limit, offset int) ([]*domain.TranslationContribution, error) {
	query := `
		SELECT id, chapter_id, contributor_id, language, contribution_type,
		       title, content, contributor_notes, status,
		       reviewed_by, reviewed_at, review_notes,
		       official_translation_id, credit_points, is_credited,
		       word_count, character_count, upvote_count, downvote_count,
		       created_at, updated_at, deleted_at
		FROM catalog.translation_contributions
		WHERE status = 'pending_review' AND deleted_at IS NULL
	`

	var args []interface{}
	argIdx := 1

	if language != nil {
		query += fmt.Sprintf(" AND language = $%d", argIdx)
		args = append(args, *language)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	contributions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.TranslationContribution])
	if err != nil {
		return nil, err
	}

	return contributions, nil
}

// Create tạo contribution mới
func (r *translationContributionRepository) Create(ctx context.Context, contribution *domain.TranslationContribution) error {
	query := `
		INSERT INTO catalog.translation_contributions (
			id, chapter_id, contributor_id, language, contribution_type,
			title, content, contributor_notes, status,
			word_count, character_count, credit_points
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		contribution.ID,
		contribution.ChapterID,
		contribution.ContributorID,
		contribution.Language,
		contribution.ContributionType,
		contribution.Title,
		contribution.Content,
		contribution.ContributorNotes,
		contribution.Status,
		contribution.WordCount,
		contribution.CharacterCount,
		contribution.CreditPoints,
	)

	return err
}

// Update cập nhật contribution
func (r *translationContributionRepository) Update(ctx context.Context, contribution *domain.TranslationContribution) error {
	query := `
		UPDATE catalog.translation_contributions
		SET title = $2,
		    content = $3,
		    contributor_notes = $4,
		    status = $5,
		    word_count = $6,
		    character_count = $7
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		contribution.ID,
		contribution.Title,
		contribution.Content,
		contribution.ContributorNotes,
		contribution.Status,
		contribution.WordCount,
		contribution.CharacterCount,
	)

	return err
}

// Delete xóa mềm contribution
func (r *translationContributionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.translation_contributions
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Approve phê duyệt contribution
func (r *translationContributionRepository) Approve(ctx context.Context, id, reviewerID uuid.UUID, reviewNotes string, creditPoints int) error {
	now := time.Now()
	query := `
		UPDATE catalog.translation_contributions
		SET status = 'approved',
		    reviewed_by = $2,
		    reviewed_at = $3,
		    review_notes = $4,
		    credit_points = $5,
		    is_credited = true
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id, reviewerID, now, reviewNotes, creditPoints)
	return err
}

// Reject từ chối contribution
func (r *translationContributionRepository) Reject(ctx context.Context, id, reviewerID uuid.UUID, reviewNotes string) error {
	now := time.Now()
	query := `
		UPDATE catalog.translation_contributions
		SET status = 'rejected',
		    reviewed_by = $2,
		    reviewed_at = $3,
		    review_notes = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id, reviewerID, now, reviewNotes)
	return err
}

// Vote vote cho contribution
func (r *translationContributionRepository) Vote(ctx context.Context, contributionID, userID uuid.UUID, isUpvote bool) error {
	// This is a simplified implementation
	// In production, you'd want a separate votes table to track individual votes

	column := "upvote_count"
	if !isUpvote {
		column = "downvote_count"
	}

	query := fmt.Sprintf(`
		UPDATE catalog.translation_contributions
		SET %s = %s + 1
		WHERE id = $1 AND deleted_at IS NULL
	`, column, column)

	_, err := r.pool.Exec(ctx, query, contributionID)
	return err
}

// UpdateStatistics cập nhật thống kê
func (r *translationContributionRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.ContributionStatisticsUpdate) error {
	var setClauses []string
	var args []interface{}
	argIdx := 2

	args = append(args, id)

	if stats.UpvoteCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("upvote_count = $%d", argIdx))
		args = append(args, *stats.UpvoteCount)
		argIdx++
	}

	if stats.DownvoteCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("downvote_count = $%d", argIdx))
		args = append(args, *stats.DownvoteCount)
		argIdx++
	}

	if stats.CreditPoints != nil {
		setClauses = append(setClauses, fmt.Sprintf("credit_points = $%d", argIdx))
		args = append(args, *stats.CreditPoints)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE catalog.translation_contributions
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
