// ============================================================================
// Organization Repository (Ent Implementation)
// ============================================================================
//
// Repository này triển khai OrganizationRepository interface sử dụng Ent ORM.
// Bao gồm 21+ methods chia thành 4 groups: CRUD, Membership, Invites, Reports.
//
// ============================================================================

package organization

import (
	"context"
	"database/sql"
	"encoding/json"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/organization"
	"system/internal/ent/generated/orgmember"
	"system/internal/ent/generated/orgpendinginvite"
	"system/internal/ent/generated/orgreport"
)

// entRepository triển khai OrganizationRepository sử dụng Ent
type entRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewEntRepository tạo một instance mới
func NewEntRepository(client *ent.Client, db *sql.DB) domain.OrganizationRepository {
	return &entRepository{client: client, db: db}
}

// =============================================================================
// Organization CRUD
// =============================================================================

// GetByID lấy organization theo ID
func (r *entRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	o, err := database.GetClientFromContext(ctx, r.client).Organization.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entOrgToDomain(o), nil
}

// GetBySlug lấy organization theo slug
func (r *entRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	o, err := database.GetClientFromContext(ctx, r.client).Organization.Query().
		Where(organization.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entOrgToDomain(o), nil
}

// SlugExists kiểm tra slug đã tồn tại chưa
func (r *entRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	return database.GetClientFromContext(ctx, r.client).Organization.Query().
		Where(organization.SlugEQ(slug)).
		Exist(ctx)
}

// List lấy danh sách organizations với filter
func (r *entRepository) List(ctx context.Context, filter domain.OrganizationFilter) ([]*domain.Organization, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).Organization.Query().
		Where(organization.DeletedAtIsNil())

	// Apply filters
	if filter.Status != nil {
		query = query.Where(organization.StatusEQ(organization.Status(*filter.Status)))
	}
	if filter.IsRecruiting != nil {
		query = query.Where(organization.IsRecruitingEQ(*filter.IsRecruiting))
	}
	if filter.CanTranslate != nil {
		query = query.Where(organization.CanTranslateEQ(*filter.CanTranslate))
	}
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		query = query.Where(organization.Or(
			organization.NameContainsFold(*filter.SearchQuery),
			organization.SlugContainsFold(*filter.SearchQuery),
		))
	}

	// Count total
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filter.SortBy {
	case "name":
		if filter.SortOrder == "desc" {
			query = query.Order(ent.Desc(organization.FieldName))
		} else {
			query = query.Order(ent.Asc(organization.FieldName))
		}
	case "member_count":
		if filter.SortOrder == "desc" {
			query = query.Order(ent.Desc(organization.FieldMemberCount))
		} else {
			query = query.Order(ent.Asc(organization.FieldMemberCount))
		}
	case "completed_translations":
		if filter.SortOrder == "desc" {
			query = query.Order(ent.Desc(organization.FieldCompletedTranslations))
		} else {
			query = query.Order(ent.Asc(organization.FieldCompletedTranslations))
		}
	default:
		if filter.SortOrder == "asc" {
			query = query.Order(ent.Asc(organization.FieldCreatedAt))
		} else {
			query = query.Order(ent.Desc(organization.FieldCreatedAt))
		}
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	orgs, err := query.All(ctx)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.Organization, len(orgs))
	for i, o := range orgs {
		results[i] = entOrgToDomain(o)
	}
	return results, int64(total), nil
}

// Create tạo organization mới
func (r *entRepository) Create(ctx context.Context, org *domain.Organization) error {
	builder := database.GetClientFromContext(ctx, r.client).Organization.Create().
		SetID(org.ID).
		SetName(org.Name).
		SetSlug(org.Slug).
		SetStatus(organization.Status(org.Status)).
		SetIsRecruiting(org.IsRecruiting).
		SetCanTranslate(org.CanTranslate).
		SetCanProofread(org.CanProofread).
		SetCanEdit(org.CanEdit).
		SetMemberCount(org.MemberCount).
		SetActiveProjects(org.ActiveProjects).
		SetCompletedTranslations(org.CompletedTranslations).
		SetReportCount(org.ReportCount).
		SetVersion(org.Version)

	if org.Description != nil {
		builder.SetDescription(org.Description)
	}
	if org.AvatarURL != nil {
		builder.SetAvatarURL(*org.AvatarURL)
	}
	if org.Settings != nil {
		builder.SetSettings(org.Settings)
	}
	if org.Metadata != nil {
		builder.SetMetadata(org.Metadata)
	}
	if org.CreatedBy != nil {
		builder.SetCreatedBy(*org.CreatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	org.CreatedAt = created.CreatedAt
	org.UpdatedAt = created.UpdatedAt
	return nil
}

// Update cập nhật organization
func (r *entRepository) Update(ctx context.Context, org *domain.Organization) error {
	builder := database.GetClientFromContext(ctx, r.client).Organization.UpdateOneID(org.ID).
		SetName(org.Name).
		SetSlug(org.Slug).
		SetStatus(organization.Status(org.Status)).
		SetIsRecruiting(org.IsRecruiting).
		SetCanTranslate(org.CanTranslate).
		SetCanProofread(org.CanProofread).
		SetCanEdit(org.CanEdit).
		SetVersion(org.Version + 1)

	if org.Description != nil {
		builder.SetDescription(org.Description)
	}
	if org.AvatarURL != nil {
		builder.SetAvatarURL(*org.AvatarURL)
	} else {
		builder.ClearAvatarURL()
	}
	if org.Settings != nil {
		builder.SetSettings(org.Settings)
	}
	if org.UpdatedBy != nil {
		builder.SetUpdatedBy(*org.UpdatedBy)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa mềm organization
func (r *entRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).Organization.UpdateOneID(id).
		SetDeletedAt(now).
		SetStatus(organization.StatusDeleted).
		Save(ctx)
	return err
}

// UpdateStatistics cập nhật thống kê organization
func (r *entRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.OrganizationStatisticsUpdate) error {
	builder := database.GetClientFromContext(ctx, r.client).Organization.UpdateOneID(id)

	if stats.MemberCount != nil {
		builder.SetMemberCount(*stats.MemberCount)
	}
	if stats.ActiveProjects != nil {
		builder.SetActiveProjects(*stats.ActiveProjects)
	}
	if stats.CompletedTranslations != nil {
		builder.SetCompletedTranslations(*stats.CompletedTranslations)
	}

	_, err := builder.Save(ctx)
	return err
}

// =============================================================================
// Membership Operations
// =============================================================================

// GetMembers lấy danh sách thành viên của organization
func (r *entRepository) GetMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.OrganizationMembership, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).OrgMember.Query().
		Where(
			orgmember.OrganizationIDEQ(orgID),
			orgmember.DeletedAtIsNil(),
		)

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	members, err := query.
		Order(ent.Desc(orgmember.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.OrganizationMembership, len(members))
	for i, m := range members {
		results[i] = entMemberToDomain(m)
	}
	return results, int64(total), nil
}

// GetMembership lấy membership của user trong org
func (r *entRepository) GetMembership(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationMembership, error) {
	m, err := database.GetClientFromContext(ctx, r.client).OrgMember.Query().
		Where(
			orgmember.UserIDEQ(userID),
			orgmember.OrganizationIDEQ(orgID),
			orgmember.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entMemberToDomain(m), nil
}

// AddMember thêm thành viên vào organization
func (r *entRepository) AddMember(ctx context.Context, membership *domain.OrganizationMembership) error {
	builder := database.GetClientFromContext(ctx, r.client).OrgMember.Create().
		SetUserID(membership.UserID).
		SetOrganizationID(membership.OrganizationID).
		SetStatus(membership.Status).
		SetRole(orgmember.Role(membership.Role)).
		SetIsActive(membership.IsActive).
		SetContributionCount(membership.ContributionCount).
		SetQualityScore(membership.QualityScore).
		SetVersion(1)

	if membership.Metadata != nil {
		builder.SetMetadata(membership.Metadata)
	}
	if membership.InvitedBy != nil {
		builder.SetInvitedBy(*membership.InvitedBy)
	}
	if membership.InvitedAt != nil {
		builder.SetInvitedAt(*membership.InvitedAt)
	}
	if membership.JoinedAt != nil {
		builder.SetJoinedAt(*membership.JoinedAt)
	}
	if membership.CreatedBy != nil {
		builder.SetCreatedBy(*membership.CreatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	membership.CreatedAt = created.CreatedAt
	membership.UpdatedAt = created.UpdatedAt
	return nil
}

// UpdateMember cập nhật thông tin thành viên
func (r *entRepository) UpdateMember(ctx context.Context, membership *domain.OrganizationMembership) error {
	_, err := database.GetClientFromContext(ctx, r.client).OrgMember.Update().
		Where(
			orgmember.UserIDEQ(membership.UserID),
			orgmember.OrganizationIDEQ(membership.OrganizationID),
		).
		SetStatus(membership.Status).
		SetRole(orgmember.Role(membership.Role)).
		SetIsActive(membership.IsActive).
		SetContributionCount(membership.ContributionCount).
		SetQualityScore(membership.QualityScore).
		SetVersion(membership.Version + 1).
		Save(ctx)
	return err
}

// RemoveMember xóa thành viên khỏi organization (soft delete)
func (r *entRepository) RemoveMember(ctx context.Context, userID, orgID uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).OrgMember.Update().
		Where(
			orgmember.UserIDEQ(userID),
			orgmember.OrganizationIDEQ(orgID),
		).
		SetDeletedAt(now).
		SetLeftAt(now).
		SetIsActive(false).
		Save(ctx)
	return err
}

// GetUserOrganizations lấy danh sách organizations của user
func (r *entRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*domain.OrganizationMembership, error) {
	members, err := database.GetClientFromContext(ctx, r.client).OrgMember.Query().
		Where(
			orgmember.UserIDEQ(userID),
			orgmember.DeletedAtIsNil(),
			orgmember.IsActiveEQ(true),
		).
		Order(ent.Desc(orgmember.FieldJoinedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.OrganizationMembership, len(members))
	for i, m := range members {
		results[i] = entMemberToDomain(m)
	}
	return results, nil
}

// CountUserMemberships đếm số memberships của user (không bao gồm owner)
func (r *entRepository) CountUserMemberships(ctx context.Context, userID uuid.UUID) (int, error) {
	return database.GetClientFromContext(ctx, r.client).OrgMember.Query().
		Where(
			orgmember.UserIDEQ(userID),
			orgmember.DeletedAtIsNil(),
			orgmember.RoleNEQ(orgmember.RoleOwner),
		).
		Count(ctx)
}

// IsUserOwnerOfAnyOrg kiểm tra user đã là owner của org nào chưa
func (r *entRepository) IsUserOwnerOfAnyOrg(ctx context.Context, userID uuid.UUID) (bool, error) {
	return database.GetClientFromContext(ctx, r.client).OrgMember.Query().
		Where(
			orgmember.UserIDEQ(userID),
			orgmember.DeletedAtIsNil(),
			orgmember.RoleEQ(orgmember.RoleOwner),
		).
		Exist(ctx)
}

// =============================================================================
// Pending Invites
// =============================================================================

// CreatePendingInvite tạo pending invite mới
func (r *entRepository) CreatePendingInvite(ctx context.Context, invite *domain.OrganizationPendingInvite) error {
	created, err := database.GetClientFromContext(ctx, r.client).OrgPendingInvite.Create().
		SetID(invite.ID).
		SetOrganizationID(invite.OrganizationID).
		SetUserID(invite.UserID).
		SetInvitedBy(invite.InvitedBy).
		SetStatus(orgpendinginvite.StatusPending).
		SetExpiresAt(invite.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}
	invite.CreatedAt = created.CreatedAt
	return nil
}

// GetPendingInvite lấy pending invite theo ID
func (r *entRepository) GetPendingInvite(ctx context.Context, id uuid.UUID) (*domain.OrganizationPendingInvite, error) {
	inv, err := database.GetClientFromContext(ctx, r.client).OrgPendingInvite.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entInviteToDomain(inv), nil
}

// GetPendingInviteByUserAndOrg lấy pending invite của user trong org
func (r *entRepository) GetPendingInviteByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationPendingInvite, error) {
	inv, err := database.GetClientFromContext(ctx, r.client).OrgPendingInvite.Query().
		Where(
			orgpendinginvite.UserIDEQ(userID),
			orgpendinginvite.OrganizationIDEQ(orgID),
			orgpendinginvite.StatusEQ(orgpendinginvite.StatusPending),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entInviteToDomain(inv), nil
}

// ListPendingInvites lấy danh sách pending invites của org
func (r *entRepository) ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationPendingInvite, error) {
	invites, err := database.GetClientFromContext(ctx, r.client).OrgPendingInvite.Query().
		Where(
			orgpendinginvite.OrganizationIDEQ(orgID),
			orgpendinginvite.StatusEQ(orgpendinginvite.StatusPending),
		).
		Order(ent.Desc(orgpendinginvite.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.OrganizationPendingInvite, len(invites))
	for i, inv := range invites {
		results[i] = entInviteToDomain(inv)
	}
	return results, nil
}

// UpdatePendingInvite cập nhật pending invite
func (r *entRepository) UpdatePendingInvite(ctx context.Context, invite *domain.OrganizationPendingInvite) error {
	builder := database.GetClientFromContext(ctx, r.client).OrgPendingInvite.UpdateOneID(invite.ID).
		SetStatus(orgpendinginvite.Status(invite.Status))

	if invite.ApprovedBy != nil {
		builder.SetApprovedBy(*invite.ApprovedBy)
	}
	if invite.ProcessedAt != nil {
		builder.SetProcessedAt(*invite.ProcessedAt)
	}

	_, err := builder.Save(ctx)
	return err
}

// DeletePendingInvite xóa pending invite
func (r *entRepository) DeletePendingInvite(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).OrgPendingInvite.DeleteOneID(id).Exec(ctx)
}

// =============================================================================
// Reports
// =============================================================================

// CreateReport tạo report mới
func (r *entRepository) CreateReport(ctx context.Context, report *domain.OrganizationReport) error {
	builder := database.GetClientFromContext(ctx, r.client).OrgReport.Create().
		SetID(report.ID).
		SetOrganizationID(report.OrganizationID).
		SetReporterID(report.ReporterID).
		SetReason(report.Reason).
		SetStatus(orgreport.StatusPending)

	if report.Description != nil {
		builder.SetDescription(*report.Description)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	report.CreatedAt = created.CreatedAt
	report.UpdatedAt = created.UpdatedAt
	return nil
}

// GetReport lấy report theo ID
func (r *entRepository) GetReport(ctx context.Context, id uuid.UUID) (*domain.OrganizationReport, error) {
	rpt, err := database.GetClientFromContext(ctx, r.client).OrgReport.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entReportToDomain(rpt), nil
}

// GetReportByUserAndOrg lấy report của user về org
func (r *entRepository) GetReportByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*domain.OrganizationReport, error) {
	rpt, err := database.GetClientFromContext(ctx, r.client).OrgReport.Query().
		Where(
			orgreport.ReporterIDEQ(userID),
			orgreport.OrganizationIDEQ(orgID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entReportToDomain(rpt), nil
}

// ListReportsByOrg lấy danh sách reports của org
func (r *entRepository) ListReportsByOrg(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationReport, error) {
	reports, err := database.GetClientFromContext(ctx, r.client).OrgReport.Query().
		Where(orgreport.OrganizationIDEQ(orgID)).
		Order(ent.Desc(orgreport.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.OrganizationReport, len(reports))
	for i, rpt := range reports {
		results[i] = entReportToDomain(rpt)
	}
	return results, nil
}

// UpdateReport cập nhật report
func (r *entRepository) UpdateReport(ctx context.Context, report *domain.OrganizationReport) error {
	builder := database.GetClientFromContext(ctx, r.client).OrgReport.UpdateOneID(report.ID).
		SetStatus(orgreport.Status(report.Status))

	if report.OrgResponse != nil {
		builder.SetOrgResponse(*report.OrgResponse)
	}
	if report.OrgRespondedBy != nil {
		builder.SetOrgRespondedBy(*report.OrgRespondedBy)
	}
	if report.OrgRespondedAt != nil {
		builder.SetOrgRespondedAt(*report.OrgRespondedAt)
	}
	if report.ResolvedBy != nil {
		builder.SetResolvedBy(*report.ResolvedBy)
	}
	if report.ResolvedAt != nil {
		builder.SetResolvedAt(*report.ResolvedAt)
	}
	if report.ResolutionNote != nil {
		builder.SetResolutionNote(*report.ResolutionNote)
	}

	_, err := builder.Save(ctx)
	return err
}

// =============================================================================
// Helper Functions
// =============================================================================

func entOrgToDomain(o *ent.Organization) *domain.Organization {
	return &domain.Organization{
		ID:                    o.ID,
		Name:                  o.Name,
		Slug:                  o.Slug,
		Status:                domain.OrganizationStatus(o.Status),
		Description:           o.Description,
		AvatarURL:             o.AvatarURL,
		Settings:              o.Settings,
		IsRecruiting:          o.IsRecruiting,
		CanTranslate:          o.CanTranslate,
		CanProofread:          o.CanProofread,
		CanEdit:               o.CanEdit,
		MemberCount:           o.MemberCount,
		ActiveProjects:        o.ActiveProjects,
		CompletedTranslations: o.CompletedTranslations,
		ReportCount:           o.ReportCount,
		Metadata:              o.Metadata,
		CreatedBy:             o.CreatedBy,
		UpdatedBy:             o.UpdatedBy,
		DeletedBy:             o.DeletedBy,
		Version:               o.Version,
		CreatedAt:             o.CreatedAt,
		UpdatedAt:             o.UpdatedAt,
		DeletedAt:             o.DeletedAt,
	}
}

func entMemberToDomain(m *ent.OrgMember) *domain.OrganizationMembership {
	var metadata json.RawMessage
	if m.Metadata != nil {
		metadata = m.Metadata
	}

	return &domain.OrganizationMembership{
		UserID:            m.UserID,
		OrganizationID:    m.OrganizationID,
		Status:            m.Status,
		Role:              domain.OrganizationMemberRole(m.Role),
		IsActive:          m.IsActive,
		ContributionCount: m.ContributionCount,
		QualityScore:      m.QualityScore,
		Metadata:          metadata,
		InvitedBy:         m.InvitedBy,
		InvitedAt:         m.InvitedAt,
		JoinedAt:          m.JoinedAt,
		LeftAt:            m.LeftAt,
		CreatedBy:         m.CreatedBy,
		UpdatedBy:         m.UpdatedBy,
		DeletedBy:         m.DeletedBy,
		Version:           m.Version,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		DeletedAt:         m.DeletedAt,
	}
}

func entInviteToDomain(inv *ent.OrgPendingInvite) *domain.OrganizationPendingInvite {
	return &domain.OrganizationPendingInvite{
		ID:             inv.ID,
		OrganizationID: inv.OrganizationID,
		UserID:         inv.UserID,
		InvitedBy:      inv.InvitedBy,
		Status:         string(inv.Status),
		ApprovedBy:     inv.ApprovedBy,
		ProcessedAt:    inv.ProcessedAt,
		ExpiresAt:      inv.ExpiresAt,
		CreatedAt:      inv.CreatedAt,
	}
}

func entReportToDomain(rpt *ent.OrgReport) *domain.OrganizationReport {
	return &domain.OrganizationReport{
		ID:             rpt.ID,
		OrganizationID: rpt.OrganizationID,
		ReporterID:     rpt.ReporterID,
		Reason:         rpt.Reason,
		Description:    rpt.Description,
		OrgResponse:    rpt.OrgResponse,
		OrgRespondedBy: rpt.OrgRespondedBy,
		OrgRespondedAt: rpt.OrgRespondedAt,
		Status:         string(rpt.Status),
		ResolvedBy:     rpt.ResolvedBy,
		ResolvedAt:     rpt.ResolvedAt,
		ResolutionNote: rpt.ResolutionNote,
		CreatedAt:      rpt.CreatedAt,
		UpdatedAt:      rpt.UpdatedAt,
	}
}
