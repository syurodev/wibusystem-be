package dto

import (
	"time"

	"github.com/google/uuid"
)

// OwnershipInfo represents ownership information embedded in requests
type OwnershipInfo struct {
	OwnershipType     string    `json:"ownership_type" validate:"required,oneof=PERSONAL TENANT COLLABORATIVE"` // PERSONAL, TENANT, COLLABORATIVE
	PrimaryOwnerID    uuid.UUID `json:"primary_owner_id" validate:"required,uuid"`                             // User ID or Tenant ID
	OriginalCreatorID uuid.UUID `json:"-"`                                                                     // Set from auth context - not from request body
	AccessLevel       string    `json:"access_level" validate:"required,oneof=PRIVATE TENANT_ONLY PUBLIC"`    // PRIVATE, TENANT_ONLY, PUBLIC
}

// CreateTransferRequest represents a request to initiate ownership transfer
type CreateTransferRequest struct {
	ContentType      string    `json:"content_type" validate:"required,oneof=NOVEL VOLUME CHAPTER"` // NOVEL, VOLUME, CHAPTER
	ContentID        uuid.UUID `json:"content_id" validate:"required,uuid"`                         // ID of content to transfer
	ToOwnerID        uuid.UUID `json:"to_owner_id" validate:"required,uuid"`                        // New owner (user or tenant ID)
	ToOwnerType      string    `json:"to_owner_type" validate:"required,oneof=user tenant"`         // 'user' or 'tenant'
	NewOwnershipType string    `json:"new_ownership_type" validate:"required,oneof=PERSONAL TENANT COLLABORATIVE"`
	NewAccessLevel   string    `json:"new_access_level" validate:"required,oneof=PRIVATE TENANT_ONLY PUBLIC"`
	TransferReason   string    `json:"transfer_reason,omitempty" validate:"max=1000"` // Optional reason
	ExpiresInDays    int       `json:"expires_in_days,omitempty" validate:"min=1,max=90"` // Days until expiration (default 7)
}

// ApproveTransferRequest represents approval of a transfer
type ApproveTransferRequest struct {
	ApproverNotes string `json:"approver_notes,omitempty" validate:"max=1000"` // Optional notes from approver
}

// TransferResponse represents a transfer object in responses
type TransferResponse struct {
	ID               uuid.UUID              `json:"id"`
	ContentType      string                 `json:"content_type"`
	ContentID        uuid.UUID              `json:"content_id"`
	FromOwnerID      uuid.UUID              `json:"from_owner_id"`
	FromOwnerType    string                 `json:"from_owner_type"`
	ToOwnerID        uuid.UUID              `json:"to_owner_id"`
	ToOwnerType      string                 `json:"to_owner_type"`
	NewOwnershipType string                 `json:"new_ownership_type"`
	NewAccessLevel   string                 `json:"new_access_level"`
	Status           string                 `json:"status"`
	TransferReason   *string                `json:"transfer_reason,omitempty"`
	InitiatedAt      time.Time              `json:"initiated_at"`
	ApprovedAt       *time.Time             `json:"approved_at,omitempty"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
	Conditions       map[string]interface{} `json:"conditions,omitempty"`
}

// CreateCollaboratorRequest represents a request to add a collaborator
type CreateCollaboratorRequest struct {
	ContentType         string   `json:"content_type" validate:"required,oneof=NOVEL VOLUME CHAPTER"` // NOVEL, VOLUME, CHAPTER
	ContentID           uuid.UUID `json:"content_id" validate:"required,uuid"`                         // ID of content
	CollaboratorID      uuid.UUID `json:"collaborator_id" validate:"required,uuid"`                    // User ID of collaborator
	Role                string   `json:"role,omitempty" validate:"max=100"`                           // Optional role label
	Permissions         []string `json:"permissions" validate:"required,min=1,dive,oneof=READ EDIT PUBLISH DELETE MANAGE_CHAPTERS MANAGE_PRICING VIEW_ANALYTICS MANAGE_COLLAB"`
	RevenueSharePercent *float64 `json:"revenue_share_percent,omitempty" validate:"omitempty,min=0,max=100"` // 0-100
	CollaborationNotes  string   `json:"collaboration_notes,omitempty" validate:"max=1000"`
}

// UpdateCollaboratorRequest represents updating a collaborator's permissions
type UpdateCollaboratorRequest struct {
	Permissions         []string `json:"permissions,omitempty" validate:"omitempty,min=1,dive,oneof=READ EDIT PUBLISH DELETE MANAGE_CHAPTERS MANAGE_PRICING VIEW_ANALYTICS MANAGE_COLLAB"`
	RevenueSharePercent *float64 `json:"revenue_share_percent,omitempty" validate:"omitempty,min=0,max=100"`
	Status              string   `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE"`
}

// CollaboratorResponse represents a collaborator in responses
type CollaboratorResponse struct {
	ID                  uuid.UUID  `json:"id"`
	ContentType         string     `json:"content_type"`
	ContentID           uuid.UUID  `json:"content_id"`
	CollaboratorID      uuid.UUID  `json:"collaborator_id"`
	Role                *string    `json:"role,omitempty"`
	Permissions         []string   `json:"permissions"`
	Status              string     `json:"status"`
	RevenueSharePercent *float64   `json:"revenue_share_percent,omitempty"`
	InvitedAt           time.Time  `json:"invited_at"`
	AcceptedAt          *time.Time `json:"accepted_at,omitempty"`
}

// ListTransfersRequest for filtering transfers
type ListTransfersRequest struct {
	ContentType string     `form:"content_type" validate:"omitempty,oneof=NOVEL VOLUME CHAPTER"`
	ContentID   *uuid.UUID `form:"content_id" validate:"omitempty,uuid"`
	Status      string     `form:"status" validate:"omitempty,oneof=PENDING APPROVED COMPLETED REJECTED CANCELLED EXPIRED"`
	Page        int        `form:"page" validate:"min=1"`
	PageSize    int        `form:"page_size" validate:"min=1,max=100"`
}

// ListCollaboratorsRequest for filtering collaborators
type ListCollaboratorsRequest struct {
	ContentType string     `form:"content_type" validate:"omitempty,oneof=NOVEL VOLUME CHAPTER"`
	ContentID   *uuid.UUID `form:"content_id" validate:"omitempty,uuid"`
	Status      string     `form:"status" validate:"omitempty,oneof=PENDING ACTIVE INACTIVE REMOVED"`
	Page        int        `form:"page" validate:"min=1"`
	PageSize    int        `form:"page_size" validate:"min=1,max=100"`
}
