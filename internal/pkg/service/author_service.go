package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/slugutil"
)

// AuthorService cung cấp business logic cho authors
type AuthorService struct {
	authorRepo domain.AuthorRepository
}

// NewAuthorService tạo một instance mới của AuthorService
func NewAuthorService(authorRepo domain.AuthorRepository) *AuthorService {
	return &AuthorService{
		authorRepo: authorRepo,
	}
}

// CreateAuthor tạo author mới
func (s *AuthorService) CreateAuthor(ctx context.Context, name, biography string, avatarURL *string, socialLinksJSON *string, createdBy uuid.UUID) (*domain.Author, error) {
	// Validate input
	if name == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Generate unique slug from name with random suffix
	authorSlug, err := slugutil.GenerateUniqueSlug(name)
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
			return nil, pkgerrors.ErrInvalidInput
		}
		socialLinks = json.RawMessage(*socialLinksJSON)
	} else {
		socialLinks = json.RawMessage("{}")
	}

	// Create author
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	author := &domain.Author{
		ID:          id,
		Name:        name,
		Slug:        authorSlug,
		Biography:   biographyJSON,
		AvatarURL:   avatarURL,
		SocialLinks: socialLinks,
		IsVerified:  false,
		CreatedBy:   createdBy,
	}

	if err := s.authorRepo.Create(ctx, author); err != nil {
		// Check for duplicate key error (slug constraint)
		if strings.Contains(err.Error(), "authors_slug_key") || 
		   strings.Contains(err.Error(), "duplicate key") {
			return nil, pkgerrors.ErrSlugAlreadyExists
		}
		return nil, err
	}

	// Retrieve the created author to get timestamps
	return s.authorRepo.GetByID(ctx, id)
}

// UpdateAuthor cập nhật thông tin author
func (s *AuthorService) UpdateAuthor(ctx context.Context, id uuid.UUID, name, biography string, avatarURL *string, socialLinksJSON *string) (*domain.Author, error) {
	// Validate input
	if name == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Get existing author
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.ErrAuthorNotFound
	}

	// Generate new slug if name changed
	newSlug := slug.Make(name)
	if newSlug != author.Slug {
		// Check if new slug already exists
		existing, err := s.authorRepo.GetBySlug(ctx, newSlug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.ErrSlugAlreadyExists
		}
		author.Slug = newSlug
	}

	// Update fields
	author.Name = name
	author.AvatarURL = avatarURL

	// Update biography JSON
	if biography != "" {
		author.Biography = json.RawMessage(fmt.Sprintf(`{"text": "%s"}`, biography))
	} else {
		author.Biography = json.RawMessage("{}")
	}

	// Update social links JSON
	if socialLinksJSON != nil && *socialLinksJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(*socialLinksJSON)) {
			return nil, pkgerrors.ErrInvalidInput
		}
		author.SocialLinks = json.RawMessage(*socialLinksJSON)
	} else {
		author.SocialLinks = json.RawMessage("{}")
	}

	if err := s.authorRepo.Update(ctx, author); err != nil {
		return nil, err
	}

	// Retrieve updated author
	return s.authorRepo.GetByID(ctx, id)
}

// DeleteAuthor xóa author (soft delete)
func (s *AuthorService) DeleteAuthor(ctx context.Context, id uuid.UUID) error {
	// Check if author exists
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.ErrAuthorNotFound
	}

	// Check if author has novels
	if author.NovelCount > 0 {
		return pkgerrors.ErrAuthorInUse
	}

	return s.authorRepo.Delete(ctx, id)
}

// GetAuthorByID lấy thông tin author theo ID
func (s *AuthorService) GetAuthorByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	return s.authorRepo.GetByID(ctx, id)
}

// GetAuthorBySlug lấy thông tin author theo slug
func (s *AuthorService) GetAuthorBySlug(ctx context.Context, slug string) (*domain.Author, error) {
	return s.authorRepo.GetBySlug(ctx, slug)
}

// ListAuthors lấy danh sách authors với pagination, search và sort
func (s *AuthorService) ListAuthors(ctx context.Context, page, limit int, search, sortBy, sortOrder string, isVerified *bool) ([]*domain.Author, int, error) {
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
		"views":   "total_views",
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

	filter := domain.AuthorFilter{
		SearchQuery: searchQuery,
		IsVerified:  isVerified,
		SortBy:      sortField,
		SortOrder:   sortOrder,
		Limit:       limit,
		Offset:      offset,
	}

	authors, total, err := s.authorRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return authors, int(total), nil
}
