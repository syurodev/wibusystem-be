// ============================================================================
// Organization Repository
// ============================================================================
//
// Repository này triển khai OrganizationRepository interface từ domain package.
// Quản lý nhóm dịch (translation teams) trong hệ thống.
//
// Organization CRUD:
//   - GetByID: Lấy org theo ID
//   - GetBySlug: Lấy org theo slug
//   - SlugExists: Kiểm tra slug tồn tại
//   - List: Lấy danh sách với filter động
//   - Create: Tạo org mới
//   - Update: Cập nhật org
//   - Delete: Soft delete org
//
// Membership Operations:
//   - GetMembers: Lấy danh sách members với User info
//   - GetMembership: Lấy membership của user
//   - AddMember: Thêm member
//   - UpdateMember: Cập nhật role/status
//   - RemoveMember: Xóa member (soft delete)
//   - GetUserOrganizations: Lấy orgs của user
//   - CountUserMemberships: Đếm số memberships
//   - IsUserOwnerOfAnyOrg: Kiểm tra user là owner
//
// Pending Invites:
//   - CreatePendingInvite: Tạo invite mới
//   - GetPendingInvite: Lấy invite theo ID
//   - GetPendingInviteByUserAndOrg: Lấy invite theo user+org
//   - ListPendingInvites: Lấy danh sách invites
//   - UpdatePendingInvite: Cập nhật status
//   - DeletePendingInvite: Xóa invite
//
// Reports:
//   - CreateReport: Tạo report mới
//   - GetReport: Lấy report theo ID
//   - GetReportByUserAndOrg: Lấy report theo user+org
//   - ListReportsByOrg: Lấy danh sách reports
//   - UpdateReport: Cập nhật response
//
// SQL queries ổn định được load từ queries/ sử dụng go:embed.
// Các queries động (List với filter) được build trong code.
//
// ============================================================================

package organization

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"system/internal/domain"
)

// =============================================================================
// SQL Queries - Organization CRUD
// =============================================================================

//go:embed queries/get_by_id.sql
var getByIDQuery string

//go:embed queries/get_by_slug.sql
var getBySlugQuery string

//go:embed queries/slug_exists.sql
var slugExistsQuery string

//go:embed queries/create.sql
var createQuery string

//go:embed queries/update.sql
var updateQuery string

//go:embed queries/delete.sql
var deleteQuery string

// =============================================================================
// SQL Queries - Membership
// =============================================================================

//go:embed queries/membership/get.sql
var getMembershipQuery string

//go:embed queries/membership/add.sql
var addMemberQuery string

//go:embed queries/membership/update.sql
var updateMemberQuery string

//go:embed queries/membership/remove.sql
var removeMemberQuery string

//go:embed queries/membership/count_user.sql
var countUserMembershipsQuery string

//go:embed queries/membership/is_owner.sql
var isUserOwnerQuery string

//go:embed queries/membership/increment_count.sql
var incrementMemberCountQuery string

//go:embed queries/membership/decrement_count.sql
var decrementMemberCountQuery string

// =============================================================================
// SQL Queries - Pending Invites
// =============================================================================

//go:embed queries/invite/create.sql
var createInviteQuery string

//go:embed queries/invite/get_by_id.sql
var getInviteByIDQuery string

//go:embed queries/invite/get_by_user_org.sql
var getInviteByUserOrgQuery string

//go:embed queries/invite/update.sql
var updateInviteQuery string

//go:embed queries/invite/delete.sql
var deleteInviteQuery string

// =============================================================================
// SQL Queries - Reports
// =============================================================================

//go:embed queries/report/create.sql
var createReportQuery string

//go:embed queries/report/get_by_id.sql
var getReportByIDQuery string

//go:embed queries/report/get_by_user_org.sql
var getReportByUserOrgQuery string

//go:embed queries/report/update.sql
var updateReportQuery string

// =============================================================================
// Repository Implementation
// =============================================================================

// organizationRepository triển khai OrganizationRepository sử dụng pgx
type organizationRepository struct {
	pool *pgxpool.Pool
}

// NewRepository tạo một instance mới của organizationRepository
func NewRepository(pool *pgxpool.Pool) domain.OrganizationRepository {
	return &organizationRepository{pool: pool}
}

// =============================================================================
// Organization CRUD
// =============================================================================

// GetByID lấy organization từ database theo ID
func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	rows, err := r.pool.Query(ctx, getByIDQuery, id)
	if err != nil {
		return nil, err
	}

	org, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Organization])
	if err != nil {
		return nil, err
	}

	return &org, nil
}

// GetBySlug lấy organization từ database theo slug
func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	rows, err := r.pool.Query(ctx, getBySlugQuery, slug)
	if err != nil {
		return nil, err
	}

	org, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Organization])
	if err != nil {
		return nil, err
	}

	return &org, nil
}

// SlugExists kiểm tra slug đã tồn tại chưa
func (r *organizationRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, slugExistsQuery, slug).Scan(&exists)
	return exists, err
}

// List lấy danh sách organizations với filter (dynamic query)
func (r *organizationRepository) List(ctx context.Context, filter domain.OrganizationFilter) ([]*domain.Organization, int64, error) {
	var whereClauses []string
	var args []any
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIdx++
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, string(*filter.Status))
		argIdx++
	}

	if filter.IsRecruiting != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_recruiting = $%d", argIdx))
		args = append(args, *filter.IsRecruiting)
		argIdx++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM identify.organizations WHERE %s", whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = filter.SortBy
		if filter.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	// Main query - using same SELECT fields as get_by_id.sql
	query := fmt.Sprintf(`
		SELECT id, name, slug, status, description, avatar_url, settings,
		       is_recruiting, can_translate, can_proofread, can_edit,
		       member_count, active_projects, completed_translations, report_count,
		       metadata, created_by, updated_by, deleted_by, version,
		       created_at, updated_at, deleted_at
		FROM identify.organizations
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	orgs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Organization])
	if err != nil {
		return nil, 0, err
	}

	return orgs, total, nil
}

// Create tạo organization mới
func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	if org.Settings == nil {
		org.Settings = json.RawMessage(`{"bypass_invite_approval": false}`)
	}

	_, err := r.pool.Exec(ctx, createQuery,
		org.ID, org.Name, org.Slug, org.Status, org.Description,
		org.AvatarURL, org.Settings, org.IsRecruiting, org.CreatedBy,
	)

	return err
}

// Update cập nhật organization
func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	_, err := r.pool.Exec(ctx, updateQuery,
		org.ID, org.Name, org.Description, org.AvatarURL, org.Settings,
		org.IsRecruiting, org.UpdatedBy,
	)

	return err
}

// Delete xóa mềm organization
func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, deleteQuery, id)
	return err
}

// =============================================================================
// Membership Operations
// =============================================================================

// GetMembers lấy danh sách thành viên của organization
func (r *organizationRepository) GetMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.OrganizationMembership, int64, error) {
	countQuery := `SELECT COUNT(*) FROM identify.user_organization_memberships WHERE organization_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT m.user_id, m.organization_id, m.status, m.role, m.is_active,
		       m.contribution_count, m.quality_score, m.metadata,
		       m.invited_by, m.invited_at, m.joined_at, m.left_at,
		       m.created_by, m.updated_by, m.deleted_by, m.version,
		       m.created_at, m.updated_at, m.deleted_at,
		       u.display_name, u.username, u.avatar_url
		FROM identify.user_organization_memberships m
		LEFT JOIN identify.users u ON m.user_id = u.id
		WHERE m.organization_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.role::text = 'owner' DESC, m.role::text = 'admin' DESC, m.joined_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var members []*domain.OrganizationMembership
	for rows.Next() {
		var m domain.OrganizationMembership
		var displayName, username, avatarURL *string

		err := rows.Scan(
			&m.UserID, &m.OrganizationID, &m.Status, &m.Role, &m.IsActive,
			&m.ContributionCount, &m.QualityScore, &m.Metadata,
			&m.InvitedBy, &m.InvitedAt, &m.JoinedAt, &m.LeftAt,
			&m.CreatedBy, &m.UpdatedBy, &m.DeletedBy, &m.Version,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
			&displayName, &username, &avatarURL,
		)
		if err != nil {
			return nil, 0, err
		}

		m.User = &domain.User{}
		if displayName != nil {
			m.User.DisplayName = displayName
		}
		if username != nil {
			m.User.Username = username
		}
		if avatarURL != nil {
			m.User.AvatarURL = avatarURL
		}

		members = append(members, &m)
	}

	return members, total, rows.Err()
}

// GetMembership lấy membership của user trong org
func (r *organizationRepository) GetMembership(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationMembership, error) {
	rows, err := r.pool.Query(ctx, getMembershipQuery, userID, orgID)
	if err != nil {
		return nil, err
	}

	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationMembership])
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// AddMember thêm thành viên vào organization
func (r *organizationRepository) AddMember(ctx context.Context, membership *domain.OrganizationMembership) error {
	_, err := r.pool.Exec(ctx, addMemberQuery,
		membership.UserID, membership.OrganizationID, membership.Status, membership.Role, membership.IsActive,
		membership.InvitedBy, membership.InvitedAt, membership.JoinedAt, membership.CreatedBy,
	)

	if err == nil {
		r.pool.Exec(ctx, incrementMemberCountQuery, membership.OrganizationID)
	}

	return err
}

// UpdateMember cập nhật thông tin thành viên
func (r *organizationRepository) UpdateMember(ctx context.Context, membership *domain.OrganizationMembership) error {
	_, err := r.pool.Exec(ctx, updateMemberQuery,
		membership.UserID, membership.OrganizationID, membership.Status, membership.Role,
		membership.IsActive, membership.UpdatedBy,
	)

	return err
}

// RemoveMember xóa thành viên khỏi organization
func (r *organizationRepository) RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, removeMemberQuery, userID, orgID)
	if err != nil {
		return err
	}

	if result.RowsAffected() > 0 {
		r.pool.Exec(ctx, decrementMemberCountQuery, orgID)
	}

	return nil
}

// GetUserOrganizations lấy danh sách organizations của user
func (r *organizationRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*domain.OrganizationMembership, error) {
	query := `
		SELECT m.user_id, m.organization_id, m.status, m.role, m.is_active,
		       m.contribution_count, m.quality_score, m.metadata,
		       m.invited_by, m.invited_at, m.joined_at, m.left_at,
		       m.created_by, m.updated_by, m.deleted_by, m.version,
		       m.created_at, m.updated_at, m.deleted_at,
		       o.id, o.name, o.slug, o.status, o.avatar_url, o.member_count
		FROM identify.user_organization_memberships m
		JOIN identify.organizations o ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.deleted_at IS NULL AND o.deleted_at IS NULL
		ORDER BY m.role::text = 'owner' DESC, m.joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []*domain.OrganizationMembership
	for rows.Next() {
		var m domain.OrganizationMembership
		var org domain.Organization

		err := rows.Scan(
			&m.UserID, &m.OrganizationID, &m.Status, &m.Role, &m.IsActive,
			&m.ContributionCount, &m.QualityScore, &m.Metadata,
			&m.InvitedBy, &m.InvitedAt, &m.JoinedAt, &m.LeftAt,
			&m.CreatedBy, &m.UpdatedBy, &m.DeletedBy, &m.Version,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
			&org.ID, &org.Name, &org.Slug, &org.Status, &org.AvatarURL, &org.MemberCount,
		)
		if err != nil {
			return nil, err
		}

		m.Organization = &org
		memberships = append(memberships, &m)
	}

	return memberships, rows.Err()
}

// CountUserMemberships đếm số memberships của user (không bao gồm owner)
func (r *organizationRepository) CountUserMemberships(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, countUserMembershipsQuery, userID).Scan(&count)
	return count, err
}

// IsUserOwnerOfAnyOrg kiểm tra user đã là owner của org nào chưa
func (r *organizationRepository) IsUserOwnerOfAnyOrg(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, isUserOwnerQuery, userID).Scan(&exists)
	return exists, err
}

// UpdateStatistics cập nhật thống kê organization (dynamic query)
func (r *organizationRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.OrganizationStatisticsUpdate) error {
	var setClauses []string
	var args []any
	argIdx := 2

	args = append(args, id)

	if stats.MemberCount != nil {
		setClauses = append(setClauses, fmt.Sprintf("member_count = $%d", argIdx))
		args = append(args, *stats.MemberCount)
		argIdx++
	}

	if stats.ActiveProjects != nil {
		setClauses = append(setClauses, fmt.Sprintf("active_projects = $%d", argIdx))
		args = append(args, *stats.ActiveProjects)
		argIdx++
	}

	if stats.CompletedTranslations != nil {
		setClauses = append(setClauses, fmt.Sprintf("completed_translations = $%d", argIdx))
		args = append(args, *stats.CompletedTranslations)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE identify.organizations
		SET %s
		WHERE id = $1 AND deleted_at IS NULL
	`, strings.Join(setClauses, ", "))

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// =============================================================================
// Pending Invites
// =============================================================================

// CreatePendingInvite tạo pending invite mới
func (r *organizationRepository) CreatePendingInvite(ctx context.Context, invite *domain.OrganizationPendingInvite) error {
	_, err := r.pool.Exec(ctx, createInviteQuery,
		invite.ID, invite.OrganizationID, invite.UserID, invite.InvitedBy, invite.Status, invite.ExpiresAt,
	)

	return err
}

// GetPendingInvite lấy pending invite theo ID
func (r *organizationRepository) GetPendingInvite(ctx context.Context, id uuid.UUID) (*domain.OrganizationPendingInvite, error) {
	rows, err := r.pool.Query(ctx, getInviteByIDQuery, id)
	if err != nil {
		return nil, err
	}

	invite, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationPendingInvite])
	if err != nil {
		return nil, err
	}

	return &invite, nil
}

// GetPendingInviteByUserAndOrg lấy pending invite của user trong org
func (r *organizationRepository) GetPendingInviteByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationPendingInvite, error) {
	rows, err := r.pool.Query(ctx, getInviteByUserOrgQuery, userID, orgID)
	if err != nil {
		return nil, err
	}

	invite, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationPendingInvite])
	if err != nil {
		return nil, err
	}

	return &invite, nil
}

// ListPendingInvites lấy danh sách pending invites của org (với User info)
func (r *organizationRepository) ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationPendingInvite, error) {
	query := `
		SELECT p.id, p.organization_id, p.user_id, p.invited_by, p.status, 
		       p.approved_by, p.processed_at, p.expires_at, p.created_at,
		       u.display_name, u.username, u.avatar_url,
		       inv.display_name as inviter_name
		FROM identify.organization_pending_invites p
		LEFT JOIN identify.users u ON p.user_id = u.id
		LEFT JOIN identify.users inv ON p.invited_by = inv.id
		WHERE p.organization_id = $1 AND p.status = 'pending' AND p.expires_at > NOW()
		ORDER BY p.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*domain.OrganizationPendingInvite
	for rows.Next() {
		var invite domain.OrganizationPendingInvite
		var displayName, username, avatarURL, inviterName *string

		err := rows.Scan(
			&invite.ID, &invite.OrganizationID, &invite.UserID, &invite.InvitedBy, &invite.Status,
			&invite.ApprovedBy, &invite.ProcessedAt, &invite.ExpiresAt, &invite.CreatedAt,
			&displayName, &username, &avatarURL, &inviterName,
		)
		if err != nil {
			return nil, err
		}

		invite.User = &domain.User{}
		if displayName != nil {
			invite.User.DisplayName = displayName
		}
		if username != nil {
			invite.User.Username = username
		}
		if avatarURL != nil {
			invite.User.AvatarURL = avatarURL
		}
		if inviterName != nil {
			invite.Inviter = &domain.User{DisplayName: inviterName}
		}

		invites = append(invites, &invite)
	}

	return invites, rows.Err()
}

// UpdatePendingInvite cập nhật pending invite
func (r *organizationRepository) UpdatePendingInvite(ctx context.Context, invite *domain.OrganizationPendingInvite) error {
	_, err := r.pool.Exec(ctx, updateInviteQuery,
		invite.ID, invite.Status, invite.ApprovedBy, invite.ProcessedAt,
	)

	return err
}

// DeletePendingInvite xóa pending invite
func (r *organizationRepository) DeletePendingInvite(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, deleteInviteQuery, id)
	return err
}

// =============================================================================
// Reports
// =============================================================================

// CreateReport tạo report mới
func (r *organizationRepository) CreateReport(ctx context.Context, report *domain.OrganizationReport) error {
	_, err := r.pool.Exec(ctx, createReportQuery,
		report.ID, report.OrganizationID, report.ReporterID, report.Reason, report.Description, report.Status,
	)

	return err
}

// GetReport lấy report theo ID
func (r *organizationRepository) GetReport(ctx context.Context, id uuid.UUID) (*domain.OrganizationReport, error) {
	rows, err := r.pool.Query(ctx, getReportByIDQuery, id)
	if err != nil {
		return nil, err
	}

	report, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationReport])
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// GetReportByUserAndOrg lấy report của user về org
func (r *organizationRepository) GetReportByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationReport, error) {
	rows, err := r.pool.Query(ctx, getReportByUserOrgQuery, userID, orgID)
	if err != nil {
		return nil, err
	}

	report, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationReport])
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// ListReportsByOrg lấy danh sách reports của org (với Reporter info)
func (r *organizationRepository) ListReportsByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationReport, error) {
	query := `
		SELECT r.id, r.organization_id, r.reporter_id, r.reason, r.description,
		       r.org_response, r.org_responded_by, r.org_responded_at,
		       r.status, r.resolved_by, r.resolved_at, r.resolution_note,
		       r.created_at, r.updated_at,
		       u.display_name as reporter_name
		FROM identify.organization_reports r
		LEFT JOIN identify.users u ON r.reporter_id = u.id
		WHERE r.organization_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*domain.OrganizationReport
	for rows.Next() {
		var report domain.OrganizationReport
		var reporterName *string

		err := rows.Scan(
			&report.ID, &report.OrganizationID, &report.ReporterID, &report.Reason, &report.Description,
			&report.OrgResponse, &report.OrgRespondedBy, &report.OrgRespondedAt,
			&report.Status, &report.ResolvedBy, &report.ResolvedAt, &report.ResolutionNote,
			&report.CreatedAt, &report.UpdatedAt,
			&reporterName,
		)
		if err != nil {
			return nil, err
		}

		if reporterName != nil {
			report.Reporter = &domain.User{DisplayName: reporterName}
		}

		reports = append(reports, &report)
	}

	return reports, rows.Err()
}

// UpdateReport cập nhật report
func (r *organizationRepository) UpdateReport(ctx context.Context, report *domain.OrganizationReport) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, updateReportQuery,
		report.ID, report.OrgResponse, report.OrgRespondedBy, report.OrgRespondedAt,
		report.Status, now,
	)

	return err
}
