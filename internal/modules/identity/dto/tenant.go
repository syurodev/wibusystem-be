// Package dto contains Data Transfer Objects for the Identity module.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateTenantRequest represents a tenant creation request.
type CreateTenantRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Slug        string  `json:"slug" binding:"required,min=1,max=100,alphanum"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// CreateTenantResponse represents a tenant creation response.
type CreateTenantResponse struct {
	Tenant  TenantResponse `json:"tenant"`
	Message string         `json:"message"`
}

// UpdateTenantRequest represents a tenant update request.
type UpdateTenantRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=1000"`
}

// UpdateTenantResponse represents a tenant update response.
type UpdateTenantResponse struct {
	Tenant  TenantResponse `json:"tenant"`
	Message string         `json:"message"`
}

// GetTenantResponse represents a get tenant response.
type GetTenantResponse struct {
	Tenant TenantResponse `json:"tenant"`
}

// DeleteTenantRequest represents a tenant deletion request.
type DeleteTenantRequest struct {
	ConfirmDeletion bool `json:"confirm_deletion" binding:"required"`
}

// DeleteTenantResponse represents a tenant deletion response.
type DeleteTenantResponse struct {
	Message string `json:"message"`
}

// ListTenantsRequest represents a list tenants request.
type ListTenantsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=active suspended"`
	Search   string `form:"search" binding:"omitempty,max=255"`
}

// ListTenantsResponse represents a list tenants response.
type ListTenantsResponse struct {
	Tenants    []TenantResponse `json:"tenants"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// TenantResponse represents a tenant in API responses.
type TenantResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	OwnerID     uuid.UUID  `json:"owner_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// AddTenantMemberRequest represents a request to add a member to a tenant.
type AddTenantMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Role   string    `json:"role" binding:"required,oneof=owner admin member viewer"`
}

// AddTenantMemberResponse represents a response for adding a tenant member.
type AddTenantMemberResponse struct {
	Member  TenantMemberResponse `json:"member"`
	Message string               `json:"message"`
}

// UpdateTenantMemberRequest represents a request to update a tenant member's role.
type UpdateTenantMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member viewer"`
}

// UpdateTenantMemberResponse represents a response for updating a tenant member.
type UpdateTenantMemberResponse struct {
	Member  TenantMemberResponse `json:"member"`
	Message string               `json:"message"`
}

// RemoveTenantMemberResponse represents a response for removing a tenant member.
type RemoveTenantMemberResponse struct {
	Message string `json:"message"`
}

// ListTenantMembersRequest represents a request to list tenant members.
type ListTenantMembersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=owner admin member viewer"`
}

// ListTenantMembersResponse represents a response for listing tenant members.
type ListTenantMembersResponse struct {
	Members    []TenantMemberResponse `json:"members"`
	Total      int                    `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// TenantMemberResponse represents a tenant member in API responses.
type TenantMemberResponse struct {
	ID        uuid.UUID    `json:"id"`
	TenantID  uuid.UUID    `json:"tenant_id"`
	UserID    uuid.UUID    `json:"user_id"`
	Role      string       `json:"role"`
	User      UserResponse `json:"user"`
	JoinedAt  time.Time    `json:"joined_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// TransferOwnershipRequest represents a request to transfer tenant ownership.
type TransferOwnershipRequest struct {
	NewOwnerID uuid.UUID `json:"new_owner_id" binding:"required"`
}

// TransferOwnershipResponse represents a response for transferring ownership.
type TransferOwnershipResponse struct {
	Tenant  TenantResponse `json:"tenant"`
	Message string         `json:"message"`
}

// GetTenantStatsResponse represents tenant statistics.
type GetTenantStatsResponse struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	MemberCount  int       `json:"member_count"`
	AdminCount   int       `json:"admin_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}
