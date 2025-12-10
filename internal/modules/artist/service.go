package artist

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/stringutil"
)

// ArtistService cung cấp business logic cho artists
type artistServiceImpl struct {
	artistRepo domain.ArtistRepository
}

// NewArtistService tạo một instance mới của ArtistService
func NewService(artistRepo domain.ArtistRepository) *artistServiceImpl {
	return &artistServiceImpl{
		artistRepo: artistRepo,
	}
}

// CreateArtist tạo artist mới
func (s *artistServiceImpl) CreateArtist(ctx context.Context, name, biography string, avatarURL *string, socialLinksJSON *string, specialization *string, createdBy uuid.UUID) (*domain.Artist, error) {
	// Validate input
	if name == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "name is required")
	}

	// Generate unique slug from name with random suffix
	artistSlug, err := stringutil.GenerateUniqueSlug(name)
	if err != nil {
		return nil, err
	}

	// Prepare biography JSON
	var biographyJSON json.RawMessage
	if biography != "" {
		biographyJSON = json.RawMessage(fmt.Sprintf(`{"text": "%s"}`, biography))
	} else {
		biographyJSON = json.RawMessage("{}")
	}

	// Prepare social links JSON
	var socialLinks json.RawMessage
	if socialLinksJSON != nil && *socialLinksJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(*socialLinksJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid social links JSON")
		}
		socialLinks = json.RawMessage(*socialLinksJSON)
	} else {
		socialLinks = json.RawMessage("{}")
	}

	// Create artist
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	artist := &domain.Artist{
		ID:             id,
		Name:           name,
		Slug:           artistSlug,
		Biography:      biographyJSON,
		AvatarURL:      avatarURL,
		SocialLinks:    socialLinks,
		Specialization: specialization,
		IsVerified:     false,
		CreatedBy:      createdBy,
	}

	if err := s.artistRepo.Create(ctx, artist); err != nil {
		// Check for duplicate key error (slug constraint)
		if strings.Contains(err.Error(), "artists_slug_key") ||
		   strings.Contains(err.Error(), "duplicate key") {
			return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
		}
		return nil, err
	}

	// Retrieve the created artist to get timestamps
	return s.artistRepo.GetByID(ctx, id)
}

// UpdateArtist cập nhật thông tin artist
func (s *artistServiceImpl) UpdateArtist(ctx context.Context, id uuid.UUID, name, biography string, avatarURL *string, socialLinksJSON *string, specialization *string) (*domain.Artist, error) {
	// Validate input
	if name == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "name is required")
	}

	// Get existing artist
	artist, err := s.artistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "artist not found")
	}

	// Generate new slug if name changed
	newSlug := slug.Make(name)
	if newSlug != artist.Slug {
		// Check if new slug already exists
		existing, err := s.artistRepo.GetBySlug(ctx, newSlug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
		}
		artist.Slug = newSlug
	}

	// Update fields
	artist.Name = name
	artist.AvatarURL = avatarURL
	artist.Specialization = specialization

	// Update biography JSON
	if biography != "" {
		artist.Biography = json.RawMessage(fmt.Sprintf(`{"text": "%s"}`, biography))
	} else {
		artist.Biography = json.RawMessage("{}")
	}

	// Update social links JSON
	if socialLinksJSON != nil && *socialLinksJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(*socialLinksJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid social links JSON")
		}
		artist.SocialLinks = json.RawMessage(*socialLinksJSON)
	} else {
		artist.SocialLinks = json.RawMessage("{}")
	}

	if err := s.artistRepo.Update(ctx, artist); err != nil {
		return nil, err
	}

	// Retrieve updated artist
	return s.artistRepo.GetByID(ctx, id)
}

// DeleteArtist xóa artist (soft delete)
func (s *artistServiceImpl) DeleteArtist(ctx context.Context, id uuid.UUID) error {
	// Check if artist exists
	artist, err := s.artistRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "artist not found")
	}

	// Check if artist has novels
	if artist.NovelCount > 0 {
		return pkgerrors.Conflict(I18nInUse, "artist is in use by novels")
	}

	return s.artistRepo.Delete(ctx, id)
}

// GetArtistByID lấy thông tin artist theo ID
func (s *artistServiceImpl) GetArtistByID(ctx context.Context, id uuid.UUID) (*domain.Artist, error) {
	return s.artistRepo.GetByID(ctx, id)
}

// GetArtistBySlug lấy thông tin artist theo slug
func (s *artistServiceImpl) GetArtistBySlug(ctx context.Context, slug string) (*domain.Artist, error) {
	return s.artistRepo.GetBySlug(ctx, slug)
}

// ListArtists lấy danh sách artists với pagination, search và sort
func (s *artistServiceImpl) ListArtists(ctx context.Context, page, limit int, search, sortBy, sortOrder string, specialization *string, isVerified *bool) ([]*domain.Artist, int, error) {
	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validate sortBy
	validSortFields := map[string]string{
		"name":    "name",
		"novels":  "novel_count",
		"created": "created_at",
	}

	sortField := "created_at" // default
	if sortBy != "" {
		if field, ok := validSortFields[sortBy]; ok {
			sortField = field
		}
	}

	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Build filter
	offset := (page - 1) * limit
	var searchQuery *string
	if search != "" {
		searchQuery = &search
	}

	filter := domain.ArtistFilter{
		SearchQuery:    searchQuery,
		Specialization: specialization,
		IsVerified:     isVerified,
		SortBy:         sortField,
		SortOrder:      sortOrder,
		Limit:          limit,
		Offset:         offset,
	}

	artists, total, err := s.artistRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return artists, int(total), nil
}

// ListSelection lấy danh sách artists rút gọn
func (s *artistServiceImpl) ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Artist, int64, error) {
	offset := (page - 1) * limit
	return s.artistRepo.ListSelection(ctx, offset, limit, search)
}

// MergeArtists gộp nhiều artists thành một
func (s *artistServiceImpl) MergeArtists(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	// Validate input
	if len(sourceIDs) == 0 {
		return pkgerrors.BadRequest(I18nInvalidInput, "source IDs are required")
	}

	// Check if target exists
	_, err := s.artistRepo.GetByID(ctx, targetID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "target artist not found")
	}

	// Validate: Ensure target is not in sources (prevent self-merge)
	for _, id := range sourceIDs {
		if id == targetID {
			return pkgerrors.BadRequest(I18nInvalidInput, "cannot merge artist with itself")
		}
	}

	return s.artistRepo.Merge(ctx, targetID, sourceIDs, mergedBy)
}

// PreviewMergeArtists xem trước kết quả gộp artists
func (s *artistServiceImpl) PreviewMergeArtists(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	// Check if target exists
	_, err := s.artistRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "target artist not found")
	}

	if len(sourceIDs) == 0 {
		return []*domain.Novel{}, nil
	}

	// Validate: Ensure target is not in sources
	for _, id := range sourceIDs {
		if id == targetID {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "cannot merge artist with itself")
		}
	}

	return s.artistRepo.GetMergePreview(ctx, targetID, sourceIDs)
}

