package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

// ArtistService cung cấp business logic cho artists
type ArtistService struct {
	artistRepo domain.ArtistRepository
}

// NewArtistService tạo một instance mới của ArtistService
func NewArtistService(artistRepo domain.ArtistRepository) *ArtistService {
	return &ArtistService{
		artistRepo: artistRepo,
	}
}

// CreateArtist tạo artist mới
func (s *ArtistService) CreateArtist(ctx context.Context, name, biography string, avatarURL *string, socialLinksJSON *string, specialization *string) (*domain.Artist, error) {
	// Validate input
	if name == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Generate slug from name
	artistSlug := slug.Make(name)

	// Check if slug already exists
	existing, err := s.artistRepo.GetBySlug(ctx, artistSlug)
	if err == nil && existing != nil {
		return nil, pkgerrors.ErrSlugAlreadyExists
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
			return nil, pkgerrors.ErrInvalidInput
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
	}

	if err := s.artistRepo.Create(ctx, artist); err != nil {
		return nil, err
	}

	// Retrieve the created artist to get timestamps
	return s.artistRepo.GetByID(ctx, id)
}

// UpdateArtist cập nhật thông tin artist
func (s *ArtistService) UpdateArtist(ctx context.Context, id uuid.UUID, name, biography string, avatarURL *string, socialLinksJSON *string, specialization *string) (*domain.Artist, error) {
	// Validate input
	if name == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Get existing artist
	artist, err := s.artistRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.ErrArtistNotFound
	}

	// Generate new slug if name changed
	newSlug := slug.Make(name)
	if newSlug != artist.Slug {
		// Check if new slug already exists
		existing, err := s.artistRepo.GetBySlug(ctx, newSlug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.ErrSlugAlreadyExists
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
			return nil, pkgerrors.ErrInvalidInput
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
func (s *ArtistService) DeleteArtist(ctx context.Context, id uuid.UUID) error {
	// Check if artist exists
	artist, err := s.artistRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.ErrArtistNotFound
	}

	// Check if artist has novels
	if artist.NovelCount > 0 {
		return pkgerrors.ErrArtistInUse
	}

	return s.artistRepo.Delete(ctx, id)
}

// GetArtistByID lấy thông tin artist theo ID
func (s *ArtistService) GetArtistByID(ctx context.Context, id uuid.UUID) (*domain.Artist, error) {
	return s.artistRepo.GetByID(ctx, id)
}

// GetArtistBySlug lấy thông tin artist theo slug
func (s *ArtistService) GetArtistBySlug(ctx context.Context, slug string) (*domain.Artist, error) {
	return s.artistRepo.GetBySlug(ctx, slug)
}

// ListArtists lấy danh sách artists với pagination, search và sort
func (s *ArtistService) ListArtists(ctx context.Context, page, limit int, search, sortBy, sortOrder string, specialization *string, isVerified *bool) ([]*domain.Artist, int, error) {
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
