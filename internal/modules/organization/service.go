// ============================================================================
// Organization Service
// ============================================================================
//
// Service này cung cấp business logic cho Organization module.
// Quản lý các nhóm dịch (translation teams) trong hệ thống.
//
// Organization Management:
//   - CreateOrganization: Tạo org mới (kiểm tra user chưa là owner)
//   - UpdateOrganization: Cập nhật thông tin org
//   - DeleteOrganization: Soft delete org
//   - GetOrganizationByID/GetOrganizationBySlug: Lấy org
//   - ListOrganizations: Lấy danh sách với filter
//   - UpdateSettings: Cập nhật settings (bypass invite, etc.)
//
// Membership Management:
//   - GetMembers: Lấy danh sách members
//   - GetUserOrganizations: Lấy orgs của user (owned + member)
//   - InviteMember: Mời user vào org (tạo pending invite nếu cần approval)
//   - RemoveMember: Xóa member (không thể xóa owner)
//   - UpdateMemberRole: Thay đổi role (admin/member)
//   - LeaveOrganization: Rời org (owner không thể rời)
//
// Invite Management:
//   - ListPendingInvites: Lấy danh sách invites
//   - ProcessPendingInvite: Approve/reject invite
//
// Report System:
//   - ReportOrganization: Báo cáo org vi phạm
//   - ListReports: Lấy danh sách reports
//   - RespondToReport: Org response to report
//   - HasUserReported: Kiểm tra user đã report chưa
//
// Business Rules:
//   - MaxMembershipCount: Giới hạn số orgs user có thể tham gia (5)
//   - User chỉ có thể own 1 org
//   - Owner không thể rời org, phải transfer ownership trước
//
// ============================================================================

package organization

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

const MaxMembershipCount = 5

// organizationServiceImpl triển khai OrganizationService
type organizationServiceImpl struct {
	orgRepo domain.OrganizationRepository
}

// NewService tạo một instance mới của OrganizationService
func NewService(orgRepo domain.OrganizationRepository) OrganizationService {
	return &organizationServiceImpl{
		orgRepo: orgRepo,
	}
}

// CreateOrganization tạo organization mới
func (s *organizationServiceImpl) CreateOrganization(ctx context.Context, name, slug string, description json.RawMessage, createdBy uuid.UUID) (*domain.Organization, error) {
	// Check if user already owns an org
	isOwner, err := s.orgRepo.IsUserOwnerOfAnyOrg(ctx, createdBy)
	if err != nil {
		return nil, err
	}
	if isOwner {
		return nil, pkgerrors.Conflict(I18nAlreadyOwner, "user is already owner of an organization")
	}

	// Check slug exists
	exists, err := s.orgRepo.SlugExists(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
	}

	id, _ := uuid.NewV4()

	var descJSON json.RawMessage
	if len(description) > 0 && string(description) != "null" {
		if !json.Valid(description) {
			return nil, pkgerrors.BadRequest(I18nValidationFailed, "invalid description JSON")
		}
		descJSON = description
	} else {
		descJSON = json.RawMessage("{}")
	}

	org := &domain.Organization{
		ID:           id,
		Name:         name,
		Slug:         slug,
		Status:       domain.OrganizationStatusActive,
		Description:  descJSON,
		Settings:     json.RawMessage(`{"bypass_invite_approval": false}`),
		IsRecruiting: false,
		CreatedBy:    &createdBy,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Add user as owner
	now := time.Now()
	membership := &domain.OrganizationMembership{
		UserID:         createdBy,
		OrganizationID: id,
		Status:         "active",
		Role:           domain.OrganizationMemberRoleOwner,
		IsActive:       true,
		InvitedBy:      &createdBy,
		InvitedAt:      &now,
		JoinedAt:       &now,
		CreatedBy:      &createdBy,
	}

	if err := s.orgRepo.AddMember(ctx, membership); err != nil {
		return nil, err
	}

	return s.orgRepo.GetByID(ctx, id)
}

// UpdateOrganization cập nhật organization
func (s *organizationServiceImpl) UpdateOrganization(ctx context.Context, orgID uuid.UUID, name *string, description json.RawMessage, avatarURL *string, isRecruiting *bool, updatedBy uuid.UUID) (*domain.Organization, error) {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "organization not found")
	}

	if name != nil {
		org.Name = *name
	}
	if len(description) > 0 && string(description) != "null" {
		if !json.Valid(description) {
			return nil, pkgerrors.BadRequest(I18nValidationFailed, "invalid description JSON")
		}
		org.Description = description
	}
	if avatarURL != nil {
		org.AvatarURL = avatarURL
	}
	if isRecruiting != nil {
		org.IsRecruiting = *isRecruiting
	}
	org.UpdatedBy = &updatedBy

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

	return s.orgRepo.GetByID(ctx, orgID)
}

// DeleteOrganization xóa organization
func (s *organizationServiceImpl) DeleteOrganization(ctx context.Context, orgID uuid.UUID, deletedBy uuid.UUID) error {
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "organization not found")
	}

	return s.orgRepo.Delete(ctx, orgID)
}

// GetOrganizationBySlug lấy organization theo slug
func (s *organizationServiceImpl) GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	return s.orgRepo.GetBySlug(ctx, slug)
}

// GetOrganizationByID lấy organization theo ID
func (s *organizationServiceImpl) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	return s.orgRepo.GetByID(ctx, id)
}

// ListOrganizations lấy danh sách organizations
func (s *organizationServiceImpl) ListOrganizations(ctx context.Context, filter domain.OrganizationFilter) ([]*domain.Organization, int64, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	return s.orgRepo.List(ctx, filter)
}

// UpdateSettings cập nhật settings
func (s *organizationServiceImpl) UpdateSettings(ctx context.Context, orgID uuid.UUID, bypassInviteApproval *bool, updatedBy uuid.UUID) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "organization not found")
	}

	settings := org.GetSettings()
	if bypassInviteApproval != nil {
		settings.BypassInviteApproval = *bypassInviteApproval
	}

	settingsJSON, _ := json.Marshal(settings)
	org.Settings = settingsJSON
	org.UpdatedBy = &updatedBy

	return s.orgRepo.Update(ctx, org)
}

// GetMembers lấy danh sách members
func (s *organizationServiceImpl) GetMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*domain.OrganizationMembership, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.orgRepo.GetMembers(ctx, orgID, limit, offset)
}

// GetUserOrganizations lấy organizations của user
func (s *organizationServiceImpl) GetUserOrganizations(ctx context.Context, userID uuid.UUID) (*domain.OrganizationMembership, []*domain.OrganizationMembership, error) {
	memberships, err := s.orgRepo.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var owned *domain.OrganizationMembership
	var member []*domain.OrganizationMembership

	for _, m := range memberships {
		if m.Role == domain.OrganizationMemberRoleOwner {
			owned = m
		} else {
			member = append(member, m)
		}
	}

	return owned, member, nil
}

// InviteMember mời member vào org
func (s *organizationServiceImpl) InviteMember(ctx context.Context, orgID, inviterID, userID uuid.UUID) error {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "organization not found")
	}

	// Check if user is already a member
	_, err = s.orgRepo.GetMembership(ctx, userID, orgID)
	if err == nil {
		return pkgerrors.Conflict(I18nUserAlreadyMember, "user is already a member")
	}

	// Check membership limit for the invited user
	count, err := s.orgRepo.CountUserMemberships(ctx, userID)
	if err != nil {
		return err
	}
	if count >= MaxMembershipCount {
		return pkgerrors.Conflict(I18nMaxMemberships, "user has reached maximum memberships")
	}

	settings := org.GetSettings()

	if settings.BypassInviteApproval {
		// Add as member directly
		now := time.Now()
		membership := &domain.OrganizationMembership{
			UserID:         userID,
			OrganizationID: orgID,
			Status:         "active",
			Role:           domain.OrganizationMemberRoleMember,
			IsActive:       true,
			InvitedBy:      &inviterID,
			InvitedAt:      &now,
			JoinedAt:       &now,
			CreatedBy:      &inviterID,
		}
		return s.orgRepo.AddMember(ctx, membership)
	}

	// Check if already invited
	_, err = s.orgRepo.GetPendingInviteByUserAndOrg(ctx, userID, orgID)
	if err == nil {
		return pkgerrors.Conflict(I18nUserAlreadyInvited, "user already has a pending invite")
	}

	// Create pending invite
	id, _ := uuid.NewV4()
	invite := &domain.OrganizationPendingInvite{
		ID:             id,
		OrganizationID: orgID,
		UserID:         userID,
		InvitedBy:      inviterID,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	return s.orgRepo.CreatePendingInvite(ctx, invite)
}

// RemoveMember xóa member khỏi org
func (s *organizationServiceImpl) RemoveMember(ctx context.Context, orgID, removedByID, userID uuid.UUID) error {
	membership, err := s.orgRepo.GetMembership(ctx, userID, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotMember, "user is not a member")
	}

	if membership.Role == domain.OrganizationMemberRoleOwner {
		return pkgerrors.Forbidden(I18nCannotKickOwner, "cannot remove owner")
	}

	return s.orgRepo.RemoveMember(ctx, userID, orgID)
}

// UpdateMemberRole thay đổi role của member
func (s *organizationServiceImpl) UpdateMemberRole(ctx context.Context, orgID, updatedByID, userID uuid.UUID, role string) error {
	membership, err := s.orgRepo.GetMembership(ctx, userID, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotMember, "user is not a member")
	}

	if membership.Role == domain.OrganizationMemberRoleOwner {
		return pkgerrors.Forbidden(I18nCannotRemoveOwner, "cannot change owner role")
	}

	membership.Role = domain.OrganizationMemberRole(role)
	membership.UpdatedBy = &updatedByID

	return s.orgRepo.UpdateMember(ctx, membership)
}

// LeaveOrganization rời khỏi org
func (s *organizationServiceImpl) LeaveOrganization(ctx context.Context, orgID, userID uuid.UUID) error {
	membership, err := s.orgRepo.GetMembership(ctx, userID, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotMember, "you are not a member")
	}

	if membership.Role == domain.OrganizationMemberRoleOwner {
		return pkgerrors.Forbidden(I18nOwnerCannotLeave, "owner cannot leave, transfer ownership first")
	}

	return s.orgRepo.RemoveMember(ctx, userID, orgID)
}

// ListPendingInvites lấy danh sách pending invites
func (s *organizationServiceImpl) ListPendingInvites(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationPendingInvite, error) {
	return s.orgRepo.ListPendingInvites(ctx, orgID)
}

// ProcessPendingInvite xử lý pending invite
func (s *organizationServiceImpl) ProcessPendingInvite(ctx context.Context, inviteID uuid.UUID, approvedBy uuid.UUID, action string) error {
	invite, err := s.orgRepo.GetPendingInvite(ctx, inviteID)
	if err != nil {
		return pkgerrors.NotFound(I18nInviteNotFound, "invite not found")
	}

	if invite.Status != "pending" {
		return pkgerrors.Conflict(I18nInviteNotFound, "invite already processed")
	}

	if time.Now().After(invite.ExpiresAt) {
		return pkgerrors.BadRequest(I18nInviteExpired, "invite has expired")
	}

	now := time.Now()
	invite.ProcessedAt = &now
	invite.ApprovedBy = &approvedBy

	if action == "approve" {
		invite.Status = "approved"

		// Add member
		membership := &domain.OrganizationMembership{
			UserID:         invite.UserID,
			OrganizationID: invite.OrganizationID,
			Status:         "active",
			Role:           domain.OrganizationMemberRoleMember,
			IsActive:       true,
			InvitedBy:      &invite.InvitedBy,
			InvitedAt:      &invite.CreatedAt,
			JoinedAt:       &now,
			CreatedBy:      &approvedBy,
		}

		if err := s.orgRepo.AddMember(ctx, membership); err != nil {
			return err
		}
	} else {
		invite.Status = "rejected"
	}

	return s.orgRepo.UpdatePendingInvite(ctx, invite)
}

// ReportOrganization report organization
func (s *organizationServiceImpl) ReportOrganization(ctx context.Context, orgID, reporterID uuid.UUID, reason string, description *string) error {
	// Check if org exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "organization not found")
	}

	// Check if user is a member (cannot report own org)
	_, err = s.orgRepo.GetMembership(ctx, reporterID, orgID)
	if err == nil {
		return pkgerrors.Forbidden(I18nCannotReportOwnOrg, "cannot report your own organization")
	}

	// Check if already reported
	_, err = s.orgRepo.GetReportByUserAndOrg(ctx, reporterID, orgID)
	if err == nil {
		return pkgerrors.Conflict(I18nAlreadyReported, "you have already reported this organization")
	}

	id, _ := uuid.NewV4()
	report := &domain.OrganizationReport{
		ID:             id,
		OrganizationID: orgID,
		ReporterID:     reporterID,
		Reason:         reason,
		Description:    description,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	return s.orgRepo.CreateReport(ctx, report)
}

// ListReports lấy danh sách reports
func (s *organizationServiceImpl) ListReports(ctx context.Context, orgID uuid.UUID) ([]*domain.OrganizationReport, error) {
	return s.orgRepo.ListReportsByOrg(ctx, orgID)
}

// RespondToReport phản hồi report
func (s *organizationServiceImpl) RespondToReport(ctx context.Context, reportID uuid.UUID, responderID uuid.UUID, response string) error {
	report, err := s.orgRepo.GetReport(ctx, reportID)
	if err != nil {
		return pkgerrors.NotFound(I18nReportNotFound, "report not found")
	}

	if report.Status == "org_responded" {
		return pkgerrors.Conflict(I18nReportAlreadyResponded, "report already responded")
	}

	now := time.Now()
	report.OrgResponse = &response
	report.OrgRespondedBy = &responderID
	report.OrgRespondedAt = &now
	report.Status = "org_responded"

	return s.orgRepo.UpdateReport(ctx, report)
}

// HasUserReported kiểm tra user đã report chưa
func (s *organizationServiceImpl) HasUserReported(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	_, err := s.orgRepo.GetReportByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return false, nil
	}
	return true, nil
}
