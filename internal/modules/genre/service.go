package genre

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v5"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

// Error message templates
const (
	errMsgGetGenre      = "failed to get genre: %w"
	errMsgCheckGenre    = "failed to check existing genre: %w"
	errMsgValidateParent = "failed to validate parent genre: %w"
)

// GenreService xử lý business logic cho genres
type genreServiceImpl struct {
	genreRepo domain.GenreRepository
}

// NewGenreService tạo instance mới của GenreService
func NewService(genreRepo domain.GenreRepository) *genreServiceImpl {
	return &genreServiceImpl{
		genreRepo: genreRepo,
	}
}



// validateSlugUnique checks if slug is unique, excluding given ID
func (s *genreServiceImpl) validateSlugUnique(ctx context.Context, newSlug string, excludeID uuid.UUID) error {
	existing, err := s.genreRepo.GetBySlug(ctx, newSlug)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf(errMsgCheckGenre, err)
	}
	if existing != nil && existing.ID != excludeID {
		return pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
	}
	return nil
}

// validateParentID validates that parent exists and is not circular
func (s *genreServiceImpl) validateParentID(ctx context.Context, parentID *uuid.UUID, selfID uuid.UUID) error {
	if parentID == nil {
		return nil
	}
	if *parentID == selfID {
		return pkgerrors.BadRequest(I18nCircularReference, "circular parent reference not allowed")
	}
	_, err := s.genreRepo.GetByID(ctx, *parentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pkgerrors.NotFound(I18nParentNotFound, "parent genre not found")
		}
		return fmt.Errorf(errMsgValidateParent, err)
	}
	return nil
}

// GenreWithTrend là genre kèm theo trend
type GenreWithTrend struct {
	*domain.Genre
	Trend domain.Trend
}

// CreateGenre tạo genre mới
func (s *genreServiceImpl) CreateGenre(ctx context.Context, name, description string, parentID *uuid.UUID, userID uuid.UUID) (*domain.Genre, error) {
	// Validate name
	if strings.TrimSpace(name) == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "name is required")
	}

	// Generate slug from name
	genreSlug := slug.Make(name)

	// Check if slug already exists
	existing, err := s.genreRepo.GetBySlug(ctx, genreSlug)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing genre: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
	}

	// If parentID is provided, validate parent exists
	if err := s.validateParentID(ctx, parentID, uuid.Nil); err != nil {
		return nil, err
	}

	// Create genre
	genre := &domain.Genre{
		ID:            uuid.Must(uuid.NewV7()),
		Name:          name,
		Slug:          genreSlug,
		Description:   &description,
		ParentID:      parentID,
		IsActive:      true,
		NovelCount:    0,
		ActiveReaders: 0,
		TotalViews:    0,
		CreatedBy:     &userID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.genreRepo.Create(ctx, genre); err != nil {
		return nil, fmt.Errorf("failed to create genre: %w", err)
	}

	return genre, nil
}

// UpdateGenre cập nhật genre
func (s *genreServiceImpl) UpdateGenre(ctx context.Context, id uuid.UUID, name, description string, parentID *uuid.UUID, isActive bool, userID uuid.UUID) (*domain.Genre, error) {
	// Get existing genre
	genre, err := s.GetGenreByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate name
	if strings.TrimSpace(name) == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "name is required")
	}

	// Generate new slug if name changed
	newSlug := slug.Make(name)
	if newSlug != genre.Slug {
		if err := s.validateSlugUnique(ctx, newSlug, id); err != nil {
			return nil, err
		}
		genre.Slug = newSlug
	}

	// Validate parent
	if err := s.validateParentID(ctx, parentID, id); err != nil {
		return nil, err
	}

	// Update genre fields
	genre.Name = name
	genre.Description = &description
	genre.ParentID = parentID
	genre.IsActive = isActive
	genre.UpdatedBy = &userID
	genre.UpdatedAt = time.Now()

	if err := s.genreRepo.Update(ctx, genre); err != nil {
		return nil, fmt.Errorf("failed to update genre: %w", err)
	}

	return genre, nil
}

// DeleteGenre xóa genre (soft delete)
func (s *genreServiceImpl) DeleteGenre(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get existing genre
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pkgerrors.NotFound(I18nNotFound, "genre not found")
		}
		return fmt.Errorf(errMsgGetGenre, err)
	}

	// Check if genre has novels (cannot delete if in use)
	if genre.NovelCount > 0 {
		return pkgerrors.Conflict(I18nInUse, "genre is in use by novels")
	}

	// Check if genre has children
	children, err := s.genreRepo.GetByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if len(children) > 0 {
		return pkgerrors.Conflict(I18nHasChildren, "genre has children")
	}

	if err := s.genreRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	return nil
}

// GetGenreByID lấy genre theo ID
func (s *genreServiceImpl) GetGenreByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.NotFound(I18nNotFound, "genre not found")
		}
		return nil, fmt.Errorf(errMsgGetGenre, err)
	}
	return genre, nil
}

// GetGenreBySlug lấy genre theo slug
func (s *genreServiceImpl) GetGenreBySlug(ctx context.Context, slug string) (*domain.Genre, error) {
	genre, err := s.genreRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.NotFound(I18nNotFound, "genre not found")
		}
		return nil, fmt.Errorf(errMsgGetGenre, err)
	}

	return genre, nil
}

// ListGenres lấy danh sách genres với pagination, search và sort
func (s *genreServiceImpl) ListGenres(ctx context.Context, page, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*GenreWithTrend, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validate sortBy
	validSortFields := map[string]bool{
		"name":    true,
		"views":   true,
		"series":  true,
		"created": true,
		"updated": true,
		"readers": true, // Added readers
	}
	if sortBy != "" && !validSortFields[sortBy] {
		sortBy = "" // Reset to default if invalid
	}

	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "" // Reset to default if invalid
	}

	offset := (page - 1) * limit

	genres, totalCount, err := s.genreRepo.List(ctx, offset, limit, search, sortBy, sortOrder, activeOnly)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list genres: %w", err)
	}

	// Calculate trend for each genre
	genresWithTrend := make([]*GenreWithTrend, len(genres))
	for i, genre := range genres {
		trend := s.calculateTrend(genre)
		genresWithTrend[i] = &GenreWithTrend{
			Genre: genre,
			Trend: trend,
		}
	}

	return genresWithTrend, totalCount, nil
}

// ListSelection lấy danh sách genres rút gọn
func (s *genreServiceImpl) ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Genre, int, error) {
	offset := (page - 1) * limit
	return s.genreRepo.ListSelection(ctx, offset, limit, search)
}

// GetAllGenres lấy tất cả genres (không phân trang)
func (s *genreServiceImpl) GetAllGenres(ctx context.Context, activeOnly bool) ([]*domain.Genre, error) {
	genres, err := s.genreRepo.GetAll(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get all genres: %w", err)
	}

	return genres, nil
}

// GetRootGenres lấy các genre gốc
func (s *genreServiceImpl) GetRootGenres(ctx context.Context) ([]*domain.Genre, error) {
	genres, err := s.genreRepo.GetRootGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get root genres: %w", err)
	}

	return genres, nil
}

// GetGenreChildren lấy các genre con
func (s *genreServiceImpl) GetGenreChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Genre, error) {
	genres, err := s.genreRepo.GetByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre children: %w", err)
	}

	return genres, nil
}

// calculateTrend tính toán trend của genre dựa trên metrics
// Đây là một implementation đơn giản, có thể nâng cấp sau
// calculateTrend tính toán trend của genre dựa trên metrics
// Đây là một implementation đơn giản, có thể nâng cấp sau
func (s *genreServiceImpl) calculateTrend(genre *domain.Genre) domain.Trend {
	// Simple heuristic based on activity metrics
	// In a real implementation, this would compare current metrics with historical data

	// If genre has high activity (active readers > 1000 or high views)
	if genre.ActiveReaders > 1000 || genre.TotalViews > 100000 {
		return domain.TrendRising
	}

	// If genre has moderate activity
	if genre.ActiveReaders > 100 || genre.TotalViews > 10000 {
		return domain.TrendStable
	}

	// If genre has low activity
	return domain.TrendFalling
}

// MergeGenres gộp nhiều genres thành một
func (s *genreServiceImpl) MergeGenres(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	// Validate request
	if targetID == uuid.Nil {
		return pkgerrors.BadRequest(I18nInvalidInput, "target ID is required")
	}
	if len(sourceIDs) == 0 {
		return nil // Nothing to do
	}

	// Check if duplicate ID in sources same as target
	for _, id := range sourceIDs {
		if id == targetID {
			return pkgerrors.BadRequest(I18nInvalidInput, "cannot merge genre with itself")
		}
	}

	// Check target exists
	target, err := s.genreRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pkgerrors.NotFound(I18nNotFound, "target genre not found")
		}
		return fmt.Errorf("failed to get target genre: %w", err)
	}

	// Call repo
	if err := s.genreRepo.Merge(ctx, target.ID, sourceIDs, mergedBy); err != nil {
		return fmt.Errorf("failed to merge genres: %w", err)
	}

	return nil
}

// PreviewMergeGenres trả về thông tin xem trước khi gộp
func (s *genreServiceImpl) PreviewMergeGenres(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.AffectedNovel, error) {
	// Validate target exists
	target, err := s.genreRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target genre: %w", err)
	}
	if target == nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "target genre not found")
	}

	// Validate source genres exist
	for _, sourceID := range sourceIDs {
		source, err := s.genreRepo.GetByID(ctx, sourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get source genre: %w", err)
		}
		if source == nil {
			return nil, pkgerrors.NotFound(I18nNotFound, "source genre not found")
		}
	}

	return s.genreRepo.GetMergePreview(ctx, target.ID, sourceIDs)
}

