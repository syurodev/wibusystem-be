package organization

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// OrganizationService interface định nghĩa business logic cho Organization module
type OrganizationService interface {
	// Organization CRUD
	CreateOrganization(ctx context.Context, name, slug string, description json.RawMessage, createdBy uuid.UUID) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, orgID uuid.UUID, name *string, description json.RawMessage, avatarURL *string, isRecruiting *bool, updatedBy uuid.UUID) (*domain.Organization, error)
	DeleteOrganization(ctx context.Context, orgID uuid.UUID, deletedBy uuid.UUID) error
	GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	ListOrganizations(ctx context.Context, filter domain.OrganizationFilter) ([]*domain.Organization, int64, error)

	// Settings
	UpdateSettings(ctx context.Context, orgID uuid.UUID, bypassInviteApproval *bool, updatedBy uuid.UUID) error

	// Membership
	GetMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.OrganizationMembership, int64, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) (*domain.OrganizationMembership, []*domain.OrganizationMembership, error)
	InviteMember(ctx context.Context, orgID, inviterID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, orgID, removedByID, userID uuid.UUID) error
	UpdateMemberRole(ctx context.Context, orgID, updatedByID, userID uuid.UUID, role string) error
	LeaveOrganization(ctx context.Context, orgID, userID uuid.UUID) error

	// Pending Invites
	ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationPendingInvite, error)
	ProcessPendingInvite(ctx context.Context, inviteID uuid.UUID, approvedBy uuid.UUID, action string) error

	// Reports
	ReportOrganization(ctx context.Context, orgID, reporterID uuid.UUID, reason string, description *string) error
	ListReports(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationReport, error)
	RespondToReport(ctx context.Context, reportID uuid.UUID, responderID uuid.UUID, response string) error
	HasUserReported(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
}
