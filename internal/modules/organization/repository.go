package organization

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"system/internal/domain"
)

// organizationRepository triển khai OrganizationRepository sử dụng pgx
type organizationRepository struct {
	pool *pgxpool.Pool
}

// NewRepository tạo một instance mới của organizationRepository
func NewRepository(pool *pgxpool.Pool) domain.OrganizationRepository {
	return &organizationRepository{pool: pool}
}

// GetByID lấy organization từ database theo ID
func (r *organizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, status, description, avatar_url, settings,
		       is_recruiting, can_translate, can_proofread, can_edit,
		       member_count, active_projects, completed_translations, report_count,
		       metadata, created_by, updated_by, deleted_by, version,
		       created_at, updated_at, deleted_at
		FROM identify.organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
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
	query := `
		SELECT id, name, slug, status, description, avatar_url, settings,
		       is_recruiting, can_translate, can_proofread, can_edit,
		       member_count, active_projects, completed_translations, report_count,
		       metadata, created_by, updated_by, deleted_by, version,
		       created_at, updated_at, deleted_at
		FROM identify.organizations
		WHERE slug = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, slug)
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
	query := `SELECT EXISTS(SELECT 1 FROM identify.organizations WHERE slug = $1 AND deleted_at IS NULL)`
	err := r.pool.QueryRow(ctx, query, slug).Scan(&exists)
	return exists, err
}

// List lấy danh sách organizations với filter
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

	// Main query
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
	query := `
		INSERT INTO identify.organizations (
			id, name, slug, status, description, avatar_url, settings, 
			is_recruiting, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	if org.Settings == nil {
		org.Settings = json.RawMessage(`{"bypass_invite_approval": false}`)
	}

	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Name, org.Slug, org.Status, org.Description,
		org.AvatarURL, org.Settings, org.IsRecruiting, org.CreatedBy,
	)

	return err
}

// Update cập nhật organization
func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	query := `
		UPDATE identify.organizations
		SET name = $2, description = $3, avatar_url = $4, settings = $5,
		    is_recruiting = $6, updated_by = $7, version = version + 1
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		org.ID, org.Name, org.Description, org.AvatarURL, org.Settings,
		org.IsRecruiting, org.UpdatedBy,
	)

	return err
}

// Delete xóa mềm organization
func (r *organizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE identify.organizations
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

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
	query := `
		SELECT user_id, organization_id, status, role, is_active,
		       contribution_count, quality_score, metadata,
		       invited_by, invited_at, joined_at, left_at,
		       created_by, updated_by, deleted_by, version,
		       created_at, updated_at, deleted_at
		FROM identify.user_organization_memberships
		WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, userID, orgID)
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
	query := `
		INSERT INTO identify.user_organization_memberships (
			user_id, organization_id, status, role, is_active, 
			invited_by, invited_at, joined_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		membership.UserID, membership.OrganizationID, membership.Status, membership.Role, membership.IsActive,
		membership.InvitedBy, membership.InvitedAt, membership.JoinedAt, membership.CreatedBy,
	)

	if err == nil {
		// Update member_count
		r.pool.Exec(ctx, `UPDATE identify.organizations SET member_count = member_count + 1 WHERE id = $1`, membership.OrganizationID)
	}

	return err
}

// UpdateMember cập nhật thông tin thành viên
func (r *organizationRepository) UpdateMember(ctx context.Context, membership *domain.OrganizationMembership) error {
	query := `
		UPDATE identify.user_organization_memberships
		SET status = $3, role = $4, is_active = $5, updated_by = $6, version = version + 1
		WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		membership.UserID, membership.OrganizationID, membership.Status, membership.Role,
		membership.IsActive, membership.UpdatedBy,
	)

	return err
}

// RemoveMember xóa thành viên khỏi organization
func (r *organizationRepository) RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error {
	query := `
		UPDATE identify.user_organization_memberships
		SET deleted_at = NOW(), left_at = NOW()
		WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, userID, orgID)
	if err != nil {
		return err
	}

	if result.RowsAffected() > 0 {
		r.pool.Exec(ctx, `UPDATE identify.organizations SET member_count = member_count - 1 WHERE id = $1 AND member_count > 0`, orgID)
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
	query := `SELECT COUNT(*) FROM identify.user_organization_memberships WHERE user_id = $1 AND role::text != 'owner' AND deleted_at IS NULL`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// IsUserOwnerOfAnyOrg kiểm tra user đã là owner của org nào chưa
func (r *organizationRepository) IsUserOwnerOfAnyOrg(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM identify.user_organization_memberships WHERE user_id = $1 AND role::text = 'owner' AND deleted_at IS NULL)`
	err := r.pool.QueryRow(ctx, query, userID).Scan(&exists)
	return exists, err
}

// UpdateStatistics cập nhật thống kê organization
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

// CreatePendingInvite tạo pending invite mới
func (r *organizationRepository) CreatePendingInvite(ctx context.Context, invite *domain.OrganizationPendingInvite) error {
	query := `
		INSERT INTO identify.organization_pending_invites (
			id, organization_id, user_id, invited_by, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		invite.ID, invite.OrganizationID, invite.UserID, invite.InvitedBy, invite.Status, invite.ExpiresAt,
	)

	return err
}

// GetPendingInvite lấy pending invite theo ID
func (r *organizationRepository) GetPendingInvite(ctx context.Context, id uuid.UUID) (*domain.OrganizationPendingInvite, error) {
	query := `
		SELECT id, organization_id, user_id, invited_by, status, approved_by, processed_at, expires_at, created_at
		FROM identify.organization_pending_invites
		WHERE id = $1
	`

	rows, err := r.pool.Query(ctx, query, id)
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
	query := `
		SELECT id, organization_id, user_id, invited_by, status, approved_by, processed_at, expires_at, created_at
		FROM identify.organization_pending_invites
		WHERE user_id = $1 AND organization_id = $2 AND status = 'pending'
	`

	rows, err := r.pool.Query(ctx, query, userID, orgID)
	if err != nil {
		return nil, err
	}

	invite, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationPendingInvite])
	if err != nil {
		return nil, err
	}

	return &invite, nil
}

// ListPendingInvites lấy danh sách pending invites của org
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
	query := `
		UPDATE identify.organization_pending_invites
		SET status = $2, approved_by = $3, processed_at = $4
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		invite.ID, invite.Status, invite.ApprovedBy, invite.ProcessedAt,
	)

	return err
}

// DeletePendingInvite xóa pending invite
func (r *organizationRepository) DeletePendingInvite(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM identify.organization_pending_invites WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// CreateReport tạo report mới
func (r *organizationRepository) CreateReport(ctx context.Context, report *domain.OrganizationReport) error {
	query := `
		INSERT INTO identify.organization_reports (
			id, organization_id, reporter_id, reason, description, status
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		report.ID, report.OrganizationID, report.ReporterID, report.Reason, report.Description, report.Status,
	)

	return err
}

// GetReport lấy report theo ID
func (r *organizationRepository) GetReport(ctx context.Context, id uuid.UUID) (*domain.OrganizationReport, error) {
	query := `
		SELECT id, organization_id, reporter_id, reason, description,
		       org_response, org_responded_by, org_responded_at,
		       status, resolved_by, resolved_at, resolution_note,
		       created_at, updated_at
		FROM identify.organization_reports
		WHERE id = $1
	`

	rows, err := r.pool.Query(ctx, query, id)
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
	query := `
		SELECT id, organization_id, reporter_id, reason, description,
		       org_response, org_responded_by, org_responded_at,
		       status, resolved_by, resolved_at, resolution_note,
		       created_at, updated_at
		FROM identify.organization_reports
		WHERE reporter_id = $1 AND organization_id = $2 AND status IN ('pending', 'org_responded')
	`

	rows, err := r.pool.Query(ctx, query, userID, orgID)
	if err != nil {
		return nil, err
	}

	report, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OrganizationReport])
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// ListReportsByOrg lấy danh sách reports của org
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
	query := `
		UPDATE identify.organization_reports
		SET org_response = $2, org_responded_by = $3, org_responded_at = $4, 
		    status = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		report.ID, report.OrgResponse, report.OrgRespondedBy, report.OrgRespondedAt,
		report.Status, now,
	)

	return err
}
