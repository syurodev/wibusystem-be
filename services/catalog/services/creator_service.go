package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services/interfaces"
)

// CreatorService implements creator-related business logic
type CreatorService struct {
	repos *repositories.Repositories
}

// NewCreatorService creates a new creator service
func NewCreatorService(repos *repositories.Repositories) interfaces.CreatorServiceInterface {
	return &CreatorService{
		repos: repos,
	}
}

// CreateCreator creates a new creator with validation
func (s *CreatorService) CreateCreator(ctx context.Context, req d.CreateCreatorRequest) (*m.Creator, error) {
	if err := s.ValidateCreatorName(req.Name); err != nil {
		return nil, err
	}

	// Check if creator name already exists
	exists, err := s.CheckCreatorExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check creator existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("creator with name '%s' already exists", req.Name)
	}

	// Create creator
	creator := &m.Creator{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
	}

	if err := s.repos.Creator.Create(ctx, creator); err != nil {
		return nil, fmt.Errorf("failed to create creator: %w", err)
	}

	return creator, nil
}

// GetCreatorByID retrieves a creator by ID
func (s *CreatorService) GetCreatorByID(ctx context.Context, creatorID uuid.UUID) (*m.Creator, error) {
	if creatorID == uuid.Nil {
		return nil, fmt.Errorf("creator ID cannot be nil")
	}

	creator, err := s.repos.Creator.GetByID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator by ID: %w", err)
	}

	return creator, nil
}

// ListCreators retrieves paginated list of creators with optional search
func (s *CreatorService) ListCreators(ctx context.Context, req d.ListCreatorsRequest) ([]*m.Creator, int64, error) {
	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Clean search term
	search := strings.TrimSpace(req.Search)

	creators, total, err := s.repos.Creator.List(ctx, req.PageSize, offset, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list creators: %w", err)
	}

	return creators, total, nil
}

// UpdateCreator updates creator information
func (s *CreatorService) UpdateCreator(ctx context.Context, creatorID uuid.UUID, req d.UpdateCreatorRequest) (*m.Creator, error) {
	if creatorID == uuid.Nil {
		return nil, fmt.Errorf("creator ID cannot be nil")
	}

	// Get existing creator
	creator, err := s.repos.Creator.GetByID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if err := s.ValidateCreatorName(name); err != nil {
			return nil, err
		}

		// Check if new name already exists (if different from current)
		if strings.ToLower(name) != strings.ToLower(creator.Name) {
			exists, err := s.CheckCreatorExists(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("failed to check creator existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("creator with name '%s' already exists", name)
			}
		}

		creator.Name = name
	}

	if req.Description != nil {
		creator.Description = req.Description
	}

	// Update creator
	if err := s.repos.Creator.Update(ctx, creator); err != nil {
		return nil, fmt.Errorf("failed to update creator: %w", err)
	}

	return creator, nil
}

// DeleteCreator removes a creator
func (s *CreatorService) DeleteCreator(ctx context.Context, creatorID uuid.UUID) error {
	if creatorID == uuid.Nil {
		return fmt.Errorf("creator ID cannot be nil")
	}

	// Check if creator exists
	_, err := s.repos.Creator.GetByID(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("creator not found: %w", err)
	}

	// Delete creator
	if err := s.repos.Creator.Delete(ctx, creatorID); err != nil {
		return fmt.Errorf("failed to delete creator: %w", err)
	}

	return nil
}

// ValidateCreatorName validates creator name
func (s *CreatorService) ValidateCreatorName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("creator name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("creator name must not exceed 255 characters")
	}

	if len(name) < 1 {
		return fmt.Errorf("creator name must be at least 1 character long")
	}

	return nil
}

// CheckCreatorExists checks if a creator name already exists
func (s *CreatorService) CheckCreatorExists(ctx context.Context, name string) (bool, error) {
	_, err := s.repos.Creator.GetByName(ctx, name)
	if err != nil {
		// If error is "not found", creator doesn't exist
		if strings.Contains(err.Error(), "no rows") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}