package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// OrganizationStatus định nghĩa trạng thái của organization
type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
	OrganizationStatusArchived  OrganizationStatus = "archived"
)

// OrganizationMemberRole định nghĩa role của thành viên trong organization
type OrganizationMemberRole string

const (
	OrganizationMemberRoleOwner   OrganizationMemberRole = "owner"
	OrganizationMemberRoleAdmin   OrganizationMemberRole = "admin"
	OrganizationMemberRoleMember  OrganizationMemberRole = "member"
)

// Organization là domain model cho nhóm dịch
type Organization struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Status      OrganizationStatus
	Description json.RawMessage
	AvatarURL   *string
	Settings    json.RawMessage

	// Capabilities
	IsRecruiting bool
	CanTranslate bool
	CanProofread bool
	CanEdit      bool

	// Statistics
	MemberCount           int
	ActiveProjects        int
	CompletedTranslations int

	// Metadata
	Metadata json.RawMessage

	// Audit
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// OrganizationMembership là domain model cho thành viên của organization
type OrganizationMembership struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	User           *User         // Optional: loaded by JOIN
	Organization   *Organization // Optional: loaded by JOIN

	Status  string
	Role    OrganizationMemberRole
	IsActive bool

	// Statistics
	ContributionCount int
	QualityScore      float64

	// Metadata
	Metadata json.RawMessage

	// Invitation
	InvitedBy *uuid.UUID
	InvitedAt *time.Time
	JoinedAt  *time.Time
	LeftAt    *time.Time

	// Audit
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// OrganizationFilter định nghĩa các filter cho việc query organizations
type OrganizationFilter struct {
	SearchQuery  *string
	Status       *OrganizationStatus
	IsRecruiting *bool
	CanTranslate *bool
	SortBy       string // "name", "member_count", "completed_translations", "created_at"
	SortOrder    string // "asc", "desc"
	Limit        int
	Offset       int
}

// OrganizationRepository định nghĩa interface cho việc truy cập dữ liệu organization
type OrganizationRepository interface {
	// GetByID lấy organization theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)

	// GetBySlug lấy organization theo slug
	GetBySlug(ctx context.Context, slug string) (*Organization, error)

	// List lấy danh sách organizations với filter
	List(ctx context.Context, filter OrganizationFilter) ([]*Organization, int64, error)

	// Create tạo organization mới
	Create(ctx context.Context, org *Organization) error

	// Update cập nhật organization
	Update(ctx context.Context, org *Organization) error

	// Delete xóa mềm organization
	Delete(ctx context.Context, id uuid.UUID) error

	// GetMembers lấy danh sách thành viên của organization
	GetMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*OrganizationMembership, int64, error)

	// AddMember thêm thành viên vào organization
	AddMember(ctx context.Context, membership *OrganizationMembership) error

	// UpdateMember cập nhật thông tin thành viên
	UpdateMember(ctx context.Context, membership *OrganizationMembership) error

	// RemoveMember xóa thành viên khỏi organization
	RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error

	// GetUserOrganizations lấy danh sách organizations của user
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*OrganizationMembership, error)

	// UpdateStatistics cập nhật thống kê organization
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats OrganizationStatisticsUpdate) error
}

// OrganizationStatisticsUpdate chứa thông tin thống kê để update
type OrganizationStatisticsUpdate struct {
	MemberCount           *int
	ActiveProjects        *int
	CompletedTranslations *int
}
