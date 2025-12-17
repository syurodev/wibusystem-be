package author

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

// AuthorService cung cấp business logic cho authors
type authorServiceImpl struct {
	authorRepo domain.AuthorRepository
}

// NewAuthorService tạo một instance mới của AuthorService
func NewService(authorRepo domain.AuthorRepository) *authorServiceImpl {
	return &authorServiceImpl{
		authorRepo: authorRepo,
	}
}

// CreateAuthor tạo author mới
func (s *authorServiceImpl) CreateAuthor(ctx context.Context, name, biography string, avatarURL *string, socialLinksJSON *string, createdBy uuid.UUID) (*domain.Author, error) {
	if name == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "name is required")
	}

	authorSlug, err := stringutil.GenerateUniqueSlug(name)
	if err != nil {
		return nil, err
	}

	var biographyJSON json.RawMessage
	if biography != "" {
		bioMap := map[string]string{"text": biography}
		bioBytes, err := json.Marshal(bioMap)
		if err != nil {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid biography")
		}
		biographyJSON = json.RawMessage(bioBytes)
	} else {
		biographyJSON = json.RawMessage("{}")
	}

	var socialLinks json.RawMessage
	if socialLinksJSON != nil && *socialLinksJSON != "" {
		if !json.Valid([]byte(*socialLinksJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid social links JSON")
		}
		socialLinks = json.RawMessage(*socialLinksJSON)
	} else {
		socialLinks = json.RawMessage("{}")
	}

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
		if strings.Contains(err.Error(), "authors_slug_key") || 
		   strings.Contains(err.Error(), "duplicate key") {
			return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
		}
		return nil, err
	}

	return s.authorRepo.GetByID(ctx, id)
}

// UpdateAuthor cập nhật thông tin author
func (s *authorServiceImpl) UpdateAuthor(ctx context.Context, id uuid.UUID, name, biography string, avatarURL *string, socialLinksJSON *string) (*domain.Author, error) {
	if name == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "name is required")
	}

	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "author not found")
	}

	newSlug := slug.Make(name)
	if newSlug != author.Slug {
		existing, err := s.authorRepo.GetBySlug(ctx, newSlug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
		}
		author.Slug = newSlug
	}

	author.Name = name
	author.AvatarURL = avatarURL

	if biography != "" {
		bioMap := map[string]string{"text": biography}
		bioBytes, err := json.Marshal(bioMap)
		if err != nil {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid biography")
		}
		author.Biography = json.RawMessage(bioBytes)
	} else {
		author.Biography = json.RawMessage("{}")
	}

	if socialLinksJSON != nil && *socialLinksJSON != "" {
		if !json.Valid([]byte(*socialLinksJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid social links JSON")
		}
		author.SocialLinks = json.RawMessage(*socialLinksJSON)
	} else {
		author.SocialLinks = json.RawMessage("{}")
	}

	if err := s.authorRepo.Update(ctx, author); err != nil {
		return nil, err
	}

	return s.authorRepo.GetByID(ctx, id)
}

// DeleteAuthor xóa author (soft delete)
func (s *authorServiceImpl) DeleteAuthor(ctx context.Context, id uuid.UUID) error {
	author, err := s.authorRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "author not found")
	}

	if author.NovelCount > 0 {
		return pkgerrors.Conflict(I18nInUse, "author is in use by novels")
	}

	return s.authorRepo.Delete(ctx, id)
}

// GetAuthorByID lấy thông tin author theo ID
func (s *authorServiceImpl) GetAuthorByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	return s.authorRepo.GetByID(ctx, id)
}

// GetAuthorBySlug lấy thông tin author theo slug
func (s *authorServiceImpl) GetAuthorBySlug(ctx context.Context, slug string) (*domain.Author, error) {
	return s.authorRepo.GetBySlug(ctx, slug)
}

// ListAuthors lấy danh sách authors với pagination, search và sort
func (s *authorServiceImpl) ListAuthors(ctx context.Context, page, limit int, search, sortBy, sortOrder string, isVerified *bool) ([]*domain.Author, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	validSortFields := map[string]string{
		"name":    "name",
		"views":   "total_views",
		"novels":  "novel_count",
		"created": "created_at",
	}

	sortField := "created_at"
	if sortBy != "" {
		if field, ok := validSortFields[sortBy]; ok {
			sortField = field
		}
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

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

// ListSelection lấy danh sách authors rút gọn
func (s *authorServiceImpl) ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Author, int64, error) {
	offset := (page - 1) * limit
	return s.authorRepo.ListSelection(ctx, offset, limit, search)
}

// MergeAuthors gộp nhiều authors thành một
func (s *authorServiceImpl) MergeAuthors(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	if len(sourceIDs) == 0 {
		return pkgerrors.BadRequest(I18nInvalidInput, "source IDs are required")
	}

	_, err := s.authorRepo.GetByID(ctx, targetID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "target author not found")
	}

	for _, id := range sourceIDs {
		if id == targetID {
			return pkgerrors.BadRequest(I18nInvalidInput, "cannot merge author with itself")
		}
	}

	return s.authorRepo.Merge(ctx, targetID, sourceIDs, mergedBy)
}

// PreviewMergeAuthors xem trước kết quả gộp authors
func (s *authorServiceImpl) PreviewMergeAuthors(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	_, err := s.authorRepo.GetByID(ctx, targetID)
	if err != nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "target author not found")
	}

	if len(sourceIDs) == 0 {
		return []*domain.Novel{}, nil
	}

	for _, id := range sourceIDs {
		if id == targetID {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "cannot merge author with itself")
		}
	}

	return s.authorRepo.GetMergePreview(ctx, targetID, sourceIDs)
}

// AddNovelAuthors thêm authors cho novel
// Used by CreateNovelUseCase for orchestrated creation
func (s *authorServiceImpl) AddNovelAuthors(ctx context.Context, novelID uuid.UUID, authorIDs []uuid.UUID) error {
	for i, authorID := range authorIDs {
		// Verify author exists
		if _, err := s.authorRepo.GetByID(ctx, authorID); err != nil {
			return fmt.Errorf("author %s not found: %w", authorID, err)
		}
		if err := s.authorRepo.AddNovelAuthor(ctx, novelID, authorID, i); err != nil {
			return fmt.Errorf("failed to add author %s: %w", authorID, err)
		}
	}
	return nil
}
