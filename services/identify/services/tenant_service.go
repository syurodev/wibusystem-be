package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services/interfaces"
)

// TenantService implements tenant-related business logic
type TenantService struct {
	repos *repositories.Repositories
}

// NewTenantService creates a new tenant service
func NewTenantService(repos *repositories.Repositories) interfaces.TenantServiceInterface {
	return &TenantService{
		repos: repos,
	}
}

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(ctx context.Context, req d.CreateTenantRequest) (*m.Tenant, error) {
	if err := s.ValidateTenantData(req); err != nil {
		return nil, err
	}

	// Check if slug already exists
	if req.Slug != "" {
		exists, err := s.repos.Tenant.SlugExists(ctx, req.Slug)
		if err != nil {
			return nil, fmt.Errorf("failed to check slug existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("tenant with slug %s already exists", req.Slug)
		}
	}

	// Convert settings to TenantSettings type
	var settings *m.TenantSettings
	if req.Settings != nil {
		ts := m.TenantSettings(req.Settings)
		settings = &ts
	}

	// Create tenant
	tenant := &m.Tenant{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Settings:    settings,
		Status:      m.TenantStatusActive,
	}

	if err := s.repos.Tenant.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// GetTenantByID retrieves a tenant by ID
func (s *TenantService) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*m.Tenant, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant ID cannot be nil")
	}

	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return tenant, nil
}

// GetTenantBySlug retrieves a tenant by slug
func (s *TenantService) GetTenantBySlug(ctx context.Context, slug string) (*m.Tenant, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	tenant, err := s.repos.Tenant.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}

	return tenant, nil
}

// ListTenants retrieves paginated list of all tenants (admin only)
func (s *TenantService) ListTenants(ctx context.Context, page, pageSize int) ([]*m.Tenant, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	tenants, total, err := s.repos.Tenant.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, total, nil
}

// GetUserTenants retrieves tenants for a specific user
func (s *TenantService) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*m.Tenant, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	tenants, err := s.repos.Tenant.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	return tenants, nil
}

// UpdateTenant updates tenant information
func (s *TenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, req d.UpdateTenantRequest) (*m.Tenant, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant ID cannot be nil")
	}

	// Get existing tenant
	tenant, err := s.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Validate and update name if provided
	if req.Name != nil {
		if err := s.validateTenantName(*req.Name); err != nil {
			return nil, err
		}
		tenant.Name = *req.Name
	}

	// Validate and update slug if provided
	if req.Slug != nil && *req.Slug != tenant.Slug {
		if err := s.validateTenantSlug(*req.Slug); err != nil {
			return nil, err
		}
		exists, err := s.repos.Tenant.SlugExists(ctx, *req.Slug)
		if err != nil {
			return nil, fmt.Errorf("failed to check slug existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("slug %s is already taken", *req.Slug)
		}
		tenant.Slug = *req.Slug
	}

	// Update description if provided
	if req.Description != nil {
		tenant.Description = req.Description
	}

	// Update settings if provided
	if req.Settings != nil {
		ts := m.TenantSettings(req.Settings)
		tenant.Settings = &ts
	}

	if err := s.repos.Tenant.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// DeleteTenant soft deletes a tenant
func (s *TenantService) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return fmt.Errorf("tenant ID cannot be nil")
	}

	if err := s.repos.Tenant.Delete(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// AddUserToTenant adds a user to a tenant with specific role
func (s *TenantService) AddUserToTenant(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return fmt.Errorf("tenant ID and user ID cannot be nil")
	}

	if err := s.validateRole(role); err != nil {
		return err
	}

	// TODO: Implement membership management when repository methods are available
	return fmt.Errorf("membership management not implemented yet")
}

// RemoveUserFromTenant removes a user from a tenant
func (s *TenantService) RemoveUserFromTenant(ctx context.Context, tenantID, userID uuid.UUID) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return fmt.Errorf("tenant ID and user ID cannot be nil")
	}

	// TODO: Implement membership management when repository methods are available
	return fmt.Errorf("membership management not implemented yet")
}

// GetTenantMembers retrieves all members of a tenant with their roles
func (s *TenantService) GetTenantMembers(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*m.Membership, int64, error) {
	if tenantID == uuid.Nil {
		return nil, 0, fmt.Errorf("tenant ID cannot be nil")
	}

	// TODO: Implement membership management when repository methods are available
	return nil, 0, fmt.Errorf("membership management not implemented yet")
}

// UpdateUserRole updates a user's role in a tenant
func (s *TenantService) UpdateUserRole(ctx context.Context, tenantID, userID uuid.UUID, newRole string) error {
	if tenantID == uuid.Nil || userID == uuid.Nil {
		return fmt.Errorf("tenant ID and user ID cannot be nil")
	}

	if err := s.validateRole(newRole); err != nil {
		return err
	}

	// TODO: Implement membership management when repository methods are available
	return fmt.Errorf("membership management not implemented yet")
}

// ValidateTenantData validates tenant data before creation/update
func (s *TenantService) ValidateTenantData(req d.CreateTenantRequest) error {
	if err := s.validateTenantName(req.Name); err != nil {
		return err
	}

	if req.Slug != "" {
		if err := s.validateTenantSlug(req.Slug); err != nil {
			return err
		}
	}

	return nil
}

// CheckSlugExists checks if tenant slug is already taken
func (s *TenantService) CheckSlugExists(ctx context.Context, slug string) (bool, error) {
	if slug == "" {
		return false, fmt.Errorf("slug cannot be empty")
	}

	return s.repos.Tenant.SlugExists(ctx, slug)
}

// CheckUserTenantAccess checks if user has access to a tenant
func (s *TenantService) CheckUserTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	if userID == uuid.Nil || tenantID == uuid.Nil {
		return false, fmt.Errorf("user ID and tenant ID cannot be nil")
	}

	// TODO: Implement membership management when repository methods are available
	// For now, return true to allow access
	return true, nil
}

// Private validation methods

func (s *TenantService) validateTenantName(name string) error {
	if name == "" {
		return fmt.Errorf("tenant name is required")
	}

	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return fmt.Errorf("tenant name must be at least 2 characters long")
	}

	if len(name) > 100 {
		return fmt.Errorf("tenant name must not exceed 100 characters")
	}

	return nil
}

func (s *TenantService) validateTenantSlug(slug string) error {
	if slug == "" {
		return nil // Slug is optional
	}

	if len(slug) < 2 {
		return fmt.Errorf("tenant slug must be at least 2 characters long")
	}

	if len(slug) > 50 {
		return fmt.Errorf("tenant slug must not exceed 50 characters")
	}

	// Slug can contain lowercase letters, numbers, and hyphens
	slugRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("tenant slug can only contain lowercase letters, numbers, and hyphens")
	}

	// Slug cannot start or end with hyphen
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return fmt.Errorf("tenant slug cannot start or end with hyphen")
	}

	return nil
}

func (s *TenantService) validateRole(role string) error {
	if role == "" {
		return fmt.Errorf("role is required")
	}

	validRoles := map[string]bool{
		"owner":  true,
		"admin":  true,
		"member": true,
		"viewer": true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s. Valid roles are: owner, admin, member, viewer", role)
	}

	return nil
}
