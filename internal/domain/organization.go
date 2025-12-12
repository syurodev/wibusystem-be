package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// OrganizationSettings định nghĩa settings của organization
type OrganizationSettings struct {
	BypassInviteApproval bool `json:"bypass_invite_approval"`
}

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
	ReportCount           int

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

// GetSettings parses settings JSON to OrganizationSettings
func (o *Organization) GetSettings() OrganizationSettings {
	var settings OrganizationSettings
	if o.Settings != nil {
		_ = json.Unmarshal(o.Settings, &settings)
	}
	return settings
}

// OrganizationMembership là domain model cho thành viên của organization
type OrganizationMembership struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	User           *User         // Optional: loaded by JOIN
	Organization   *Organization // Optional: loaded by JOIN

	Status   string
	Role     OrganizationMemberRole
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

// OrganizationPendingInvite là domain model cho pending invite
type OrganizationPendingInvite struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	InvitedBy      uuid.UUID
	Status         string // pending, approved, rejected, expired
	ApprovedBy     *uuid.UUID
	ProcessedAt    *time.Time
	ExpiresAt      time.Time
	CreatedAt      time.Time

	// Relations (optional)
	User        *User
	Inviter     *User
	Organization *Organization
}

// OrganizationReport là domain model cho report
type OrganizationReport struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	ReporterID     uuid.UUID
	Reason         string
	Description    *string
	OrgResponse    *string
	OrgRespondedBy *uuid.UUID
	OrgRespondedAt *time.Time
	Status         string // pending, org_responded, reviewing, resolved, dismissed
	ResolvedBy     *uuid.UUID
	ResolvedAt     *time.Time
	ResolutionNote *string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations (optional)
	Reporter     *User
	Organization *Organization
	Responder    *User
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

	// SlugExists kiểm tra slug đã tồn tại chưa
	SlugExists(ctx context.Context, slug string) (bool, error)

	// --- Membership ---

	// GetMembers lấy danh sách thành viên của organization
	GetMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*OrganizationMembership, int64, error)

	// GetMembership lấy membership của user trong org
	GetMembership(ctx context.Context, userID, orgID uuid.UUID) (*OrganizationMembership, error)

	// AddMember thêm thành viên vào organization
	AddMember(ctx context.Context, membership *OrganizationMembership) error

	// UpdateMember cập nhật thông tin thành viên
	UpdateMember(ctx context.Context, membership *OrganizationMembership) error

	// RemoveMember xóa thành viên khỏi organization
	RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error

	// GetUserOrganizations lấy danh sách organizations của user
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*OrganizationMembership, error)

	// CountUserMemberships đếm số memberships của user (không bao gồm owner)
	CountUserMemberships(ctx context.Context, userID uuid.UUID) (int, error)

	// IsUserOwnerOfAnyOrg kiểm tra user đã là owner của org nào chưa
	IsUserOwnerOfAnyOrg(ctx context.Context, userID uuid.UUID) (bool, error)

	// UpdateStatistics cập nhật thống kê organization
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats OrganizationStatisticsUpdate) error

	// --- Pending Invites ---

	// CreatePendingInvite tạo pending invite mới
	CreatePendingInvite(ctx context.Context, invite *OrganizationPendingInvite) error

	// GetPendingInvite lấy pending invite theo ID
	GetPendingInvite(ctx context.Context, id uuid.UUID) (*OrganizationPendingInvite, error)

	// GetPendingInviteByUserAndOrg lấy pending invite của user trong org
	GetPendingInviteByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*OrganizationPendingInvite, error)

	// ListPendingInvites lấy danh sách pending invites của org
	ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]*OrganizationPendingInvite, error)

	// UpdatePendingInvite cập nhật pending invite
	UpdatePendingInvite(ctx context.Context, invite *OrganizationPendingInvite) error

	// DeletePendingInvite xóa pending invite
	DeletePendingInvite(ctx context.Context, id uuid.UUID) error

	// --- Reports ---

	// CreateReport tạo report mới
	CreateReport(ctx context.Context, report *OrganizationReport) error

	// GetReport lấy report theo ID
	GetReport(ctx context.Context, id uuid.UUID) (*OrganizationReport, error)

	// GetReportByUserAndOrg lấy report của user về org
	GetReportByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*OrganizationReport, error)

	// ListReportsByOrg lấy danh sách reports của org
	ListReportsByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationReport, error)

	// UpdateReport cập nhật report
	UpdateReport(ctx context.Context, report *OrganizationReport) error
}

// OrganizationStatisticsUpdate chứa thông tin thống kê để update
type OrganizationStatisticsUpdate struct {
	MemberCount           *int
	ActiveProjects        *int
	CompletedTranslations *int
}

