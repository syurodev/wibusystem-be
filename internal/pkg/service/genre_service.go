package service

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

// GenreService xử lý business logic cho genres
type GenreService struct {
	genreRepo domain.GenreRepository
}

// NewGenreService tạo instance mới của GenreService
func NewGenreService(genreRepo domain.GenreRepository) *GenreService {
	return &GenreService{
		genreRepo: genreRepo,
	}
}

// Use domain.Trend instead of local type
// Removed local GenreTrend definition

// GenreWithTrend là genre kèm theo trend
type GenreWithTrend struct {
	*domain.Genre
	Trend domain.Trend
}

// CreateGenre tạo genre mới
func (s *GenreService) CreateGenre(ctx context.Context, name, description string, parentID *uuid.UUID, userID uuid.UUID) (*domain.Genre, error) {
	// Validate name
	if strings.TrimSpace(name) == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Generate slug from name
	genreSlug := slug.Make(name)

	// Check if slug already exists
	existing, err := s.genreRepo.GetBySlug(ctx, genreSlug)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing genre: %w", err)
	}
	if existing != nil {
		return nil, pkgerrors.ErrSlugAlreadyExists
	}

	// If parentID is provided, validate parent exists
	if parentID != nil {
		_, err := s.genreRepo.GetByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, pkgerrors.ErrParentGenreNotFound
			}
			return nil, fmt.Errorf("failed to validate parent genre: %w", err)
		}
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
func (s *GenreService) UpdateGenre(ctx context.Context, id uuid.UUID, name, description string, parentID *uuid.UUID, isActive bool, userID uuid.UUID) (*domain.Genre, error) {
	// Get existing genre
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrGenreNotFound
		}
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}

	// Validate name
	if strings.TrimSpace(name) == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Generate new slug if name changed
	newSlug := slug.Make(name)
	if newSlug != genre.Slug {
		// Check if new slug already exists
		existing, err := s.genreRepo.GetBySlug(ctx, newSlug)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to check existing genre: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, pkgerrors.ErrSlugAlreadyExists
		}
		genre.Slug = newSlug
	}

	// Validate parent (cannot be itself)
	if parentID != nil {
		if *parentID == id {
			return nil, pkgerrors.ErrCircularParentReference
		}

		// Validate parent exists
		_, err := s.genreRepo.GetByID(ctx, *parentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, pkgerrors.ErrParentGenreNotFound
			}
			return nil, fmt.Errorf("failed to validate parent genre: %w", err)
		}
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
func (s *GenreService) DeleteGenre(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get existing genre
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pkgerrors.ErrGenreNotFound
		}
		return fmt.Errorf("failed to get genre: %w", err)
	}

	// Check if genre has novels (cannot delete if in use)
	if genre.NovelCount > 0 {
		return pkgerrors.ErrGenreInUse
	}

	// Check if genre has children
	children, err := s.genreRepo.GetByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if len(children) > 0 {
		return pkgerrors.ErrGenreHasChildren
	}

	if err := s.genreRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	return nil
}

// GetGenreByID lấy genre theo ID
func (s *GenreService) GetGenreByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	genre, err := s.genreRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrGenreNotFound
		}
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}

	return genre, nil
}

// GetGenreBySlug lấy genre theo slug
func (s *GenreService) GetGenreBySlug(ctx context.Context, slug string) (*domain.Genre, error) {
	genre, err := s.genreRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrGenreNotFound
		}
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}

	return genre, nil
}

// ListGenres lấy danh sách genres với pagination, search và sort
func (s *GenreService) ListGenres(ctx context.Context, page, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*GenreWithTrend, int, error) {
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
func (s *GenreService) ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Genre, int, error) {
	offset := (page - 1) * limit
	return s.genreRepo.ListSelection(ctx, offset, limit, search)
}

// GetAllGenres lấy tất cả genres (không phân trang)
func (s *GenreService) GetAllGenres(ctx context.Context, activeOnly bool) ([]*domain.Genre, error) {
	genres, err := s.genreRepo.GetAll(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to get all genres: %w", err)
	}

	return genres, nil
}

// GetRootGenres lấy các genre gốc
func (s *GenreService) GetRootGenres(ctx context.Context) ([]*domain.Genre, error) {
	genres, err := s.genreRepo.GetRootGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get root genres: %w", err)
	}

	return genres, nil
}

// GetGenreChildren lấy các genre con
func (s *GenreService) GetGenreChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Genre, error) {
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
func (s *GenreService) calculateTrend(genre *domain.Genre) domain.Trend {
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
func (s *GenreService) MergeGenres(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	// Validate request
	if targetID == uuid.Nil {
		return pkgerrors.ErrInvalidInput
	}
	if len(sourceIDs) == 0 {
		return nil // Nothing to do
	}

	// Check if duplicate ID in sources same as target
	for _, id := range sourceIDs {
		if id == targetID {
			return pkgerrors.ErrInvalidInput // Cannot merge into itself
		}
	}

	// Check target exists
	target, err := s.genreRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pkgerrors.ErrGenreNotFound
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
func (s *GenreService) PreviewMergeGenres(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	// Validate target
	target, err := s.genreRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrGenreNotFound
		}
		return nil, fmt.Errorf("failed to get target genre: %w", err)
	}

	// Validate sources
	if len(sourceIDs) == 0 {
		return nil, nil
	}

	// Check self-merge
	for _, id := range sourceIDs {
		if id == targetID {
			return nil, pkgerrors.ErrInvalidInput
		}
	}

	return s.genreRepo.GetMergePreview(ctx, target.ID, sourceIDs)
}
