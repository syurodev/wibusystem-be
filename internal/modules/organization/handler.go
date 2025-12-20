package organization

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	orgdto "system/internal/dto/organization"
	analytics_module "system/internal/modules/analytics"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

type Handler struct {
	orgService   OrganizationService
	analyticsSvc analytics_module.AnalyticsService
	logger       *zap.Logger
}

func NewHandler(orgService OrganizationService, analyticsSvc analytics_module.AnalyticsService, logger *zap.Logger) *Handler {
	return &Handler{
		orgService:   orgService,
		analyticsSvc: analyticsSvc,
		logger:       logger,
	}
}

// CreateOrganization tạo organization mới
func (h *Handler) CreateOrganization(c *gin.Context) {
	var req orgdto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	org, err := h.orgService.CreateOrganization(c.Request.Context(), req.Name, req.Slug, req.Description, userID)
	if err != nil {
		h.handleError(c, err, I18nCreateFailed)
		return
	}

	resp := mapToOrganizationDetailResponse(org, nil, true, false)
	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateOrganization cập nhật organization
func (h *Handler) UpdateOrganization(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	var req orgdto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	updatedOrg, err := h.orgService.UpdateOrganization(c.Request.Context(), org.ID, req.Name, req.Description, req.AvatarURL, req.IsRecruiting, userID)
	if err != nil {
		h.handleError(c, err, I18nUpdateFailed)
		return
	}

	resp := mapToOrganizationDetailResponse(updatedOrg, nil, true, false)
	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteOrganization xóa organization
func (h *Handler) DeleteOrganization(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.DeleteOrganization(c.Request.Context(), org.ID, userID); err != nil {
		h.handleError(c, err, I18nDeleteFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nDeletedSuccess, nil, nil)
}

// GetOrganization lấy thông tin organization (public - dùng slug)
func (h *Handler) GetOrganization(c *gin.Context) {
	org, err := h.getOrgBySlug(c)
	if err != nil {
		return
	}

	// Check if current user is member
	var myRole *string
	var hasReported bool
	userIDStr, exists := middleware.GetUserID(c)
	if exists {
		if userID, err := uuid.FromString(userIDStr); err == nil {
			hasReported, _ = h.orgService.HasUserReported(c.Request.Context(), org.ID, userID)
		}
	}

	resp := mapToOrganizationDetailResponse(org, myRole, false, hasReported)
	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListOrganizations lấy danh sách organizations
func (h *Handler) ListOrganizations(c *gin.Context) {
	var req orgdto.ListOrganizationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	var status *domain.OrganizationStatus
	if req.Status != nil {
		s := domain.OrganizationStatus(*req.Status)
		status = &s
	}

	filter := domain.OrganizationFilter{
		SearchQuery:  req.Search,
		Status:       status,
		IsRecruiting: req.IsRecruiting,
		SortBy:       req.SortBy,
		SortOrder:    req.SortOrder,
		Limit:        req.Limit,
		Offset:       req.Offset,
	}

	orgs, total, err := h.orgService.ListOrganizations(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err, I18nListFailed)
		return
	}

	orgResponses := make([]orgdto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		orgResponses[i] = mapToOrganizationResponse(org)
	}

	resp := orgdto.ListOrganizationsResponse{
		Organizations: orgResponses,
		Total:         total,
		Limit:         req.Limit,
		Offset:        req.Offset,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, resp, nil)
}

// UpdateSettings cập nhật settings
func (h *Handler) UpdateSettings(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	var req orgdto.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.UpdateSettings(c.Request.Context(), org.ID, req.BypassInviteApproval, userID); err != nil {
		h.handleError(c, err, I18nUpdateFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nSettingsUpdatedSuccess, nil, nil)
}

// ListMembers lấy danh sách members (public - dùng slug)
func (h *Handler) ListMembers(c *gin.Context) {
	org, err := h.getOrgBySlug(c)
	if err != nil {
		return
	}

	var req orgdto.ListMembersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	members, total, err := h.orgService.GetMembers(c.Request.Context(), org.ID, req.Limit, req.Offset)
	if err != nil {
		h.handleError(c, err, I18nListFailed)
		return
	}

	memberResponses := make([]orgdto.MemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = mapToMemberResponse(m)
	}

	resp := orgdto.ListMembersResponse{
		Members: memberResponses,
		Total:   total,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, resp, nil)
}

// InviteMember mời member vào org
func (h *Handler) InviteMember(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	var req orgdto.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.InviteMember(c.Request.Context(), org.ID, userID, req.UserID); err != nil {
		h.handleError(c, err, I18nInviteFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nMemberInvitedSuccess, nil, nil)
}

// RemoveMember xóa member khỏi org
func (h *Handler) RemoveMember(c *gin.Context) {
	userIDStr := c.Param("user_id")

	targetUserID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nInvalidID, nil)
		return
	}

	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	currentUserID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.RemoveMember(c.Request.Context(), org.ID, currentUserID, targetUserID); err != nil {
		h.handleError(c, err, I18nDeleteFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nMemberRemovedSuccess, nil, nil)
}

// UpdateMemberRole thay đổi role của member
func (h *Handler) UpdateMemberRole(c *gin.Context) {
	userIDStr := c.Param("user_id")

	targetUserID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nInvalidID, nil)
		return
	}

	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	var req orgdto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	currentUserID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.UpdateMemberRole(c.Request.Context(), org.ID, currentUserID, targetUserID, req.Role); err != nil {
		h.handleError(c, err, I18nUpdateFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nMemberRoleUpdatedSuccess, nil, nil)
}

// LeaveOrganization rời khỏi org
func (h *Handler) LeaveOrganization(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.LeaveOrganization(c.Request.Context(), org.ID, userID); err != nil {
		h.handleError(c, err, I18nDeleteFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nMemberRemovedSuccess, nil, nil)
}

// ListPendingInvites lấy danh sách pending invites
func (h *Handler) ListPendingInvites(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	invites, err := h.orgService.ListPendingInvites(c.Request.Context(), org.ID)
	if err != nil {
		h.handleError(c, err, I18nListFailed)
		return
	}

	inviteResponses := make([]orgdto.PendingInviteResponse, len(invites))
	for i, inv := range invites {
		inviteResponses[i] = mapToPendingInviteResponse(inv)
	}

	resp := orgdto.ListPendingInvitesResponse{
		Invites: inviteResponses,
		Total:   int64(len(invites)),
	}

	response.Success(c, http.StatusOK, I18nListSuccess, resp, nil)
}

// ProcessPendingInvite xử lý pending invite
func (h *Handler) ProcessPendingInvite(c *gin.Context) {
	inviteIDStr := c.Param("id")

	inviteID, err := uuid.FromString(inviteIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidID", I18nInvalidID, nil)
		return
	}

	var req orgdto.ProcessInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.ProcessPendingInvite(c.Request.Context(), inviteID, userID, req.Action); err != nil {
		h.handleError(c, err, I18nUpdateFailed)
		return
	}

	if req.Action == "approve" {
		response.Success(c, http.StatusOK, I18nInviteApprovedSuccess, nil, nil)
	} else {
		response.Success(c, http.StatusOK, I18nInviteRejectedSuccess, nil, nil)
	}
}

// ReportOrganization report organization
func (h *Handler) ReportOrganization(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	var req orgdto.ReportOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.ReportOrganization(c.Request.Context(), org.ID, userID, req.Reason, req.Description); err != nil {
		h.handleError(c, err, I18nReportFailed)
		return
	}

	response.Success(c, http.StatusCreated, I18nReportCreatedSuccess, nil, nil)
}

// ListReports lấy danh sách reports
func (h *Handler) ListReports(c *gin.Context) {
	org, err := h.getOrgByID(c)
	if err != nil {
		return
	}

	reports, err := h.orgService.ListReports(c.Request.Context(), org.ID)
	if err != nil {
		h.handleError(c, err, I18nListFailed)
		return
	}

	reportResponses := make([]orgdto.ReportResponse, len(reports))
	for i, r := range reports {
		reportResponses[i] = mapToReportResponse(r)
	}

	resp := orgdto.ListReportsResponse{
		Reports: reportResponses,
		Total:   int64(len(reports)),
	}

	response.Success(c, http.StatusOK, I18nListSuccess, resp, nil)
}

// RespondToReport phản hồi report
func (h *Handler) RespondToReport(c *gin.Context) {
	reportIDStr := c.Param("id")

	reportID, err := uuid.FromString(reportIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidID", I18nInvalidID, nil)
		return
	}

	var req orgdto.RespondReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	if err := h.orgService.RespondToReport(c.Request.Context(), reportID, userID, req.Response); err != nil {
		h.handleError(c, err, I18nUpdateFailed)
		return
	}

	response.Success(c, http.StatusOK, I18nReportRespondedSuccess, nil, nil)
}

// GetMyOrganizations lấy organizations của user
func (h *Handler) GetMyOrganizations(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	owned, member, err := h.orgService.GetUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err, I18nListFailed)
		return
	}

	resp := orgdto.MyOrganizationsResponse{
		Member: make([]orgdto.OrganizationResponse, len(member)),
	}

	if owned != nil && owned.Organization != nil {
		ownedOrg := mapToOrganizationResponse(owned.Organization)
		resp.Owned = &ownedOrg
	}

	for i, m := range member {
		if m.Organization != nil {
			resp.Member[i] = mapToOrganizationResponse(m.Organization)
		}
	}

	response.Success(c, http.StatusOK, I18nListSuccess, resp, nil)
}

// Helper functions

func (h *Handler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", I18nAuthUnauthorized, nil)
		return uuid.Nil, pkgerrors.Unauthorized(I18nAuthUnauthorized, "unauthorized")
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nAuthInvalidUserID, nil)
		return uuid.Nil, err
	}

	return userID, nil
}

// getOrgByID lấy org bằng ID (cho protected routes)
func (h *Handler) getOrgByID(c *gin.Context) (*domain.Organization, error) {
	identifier := c.Param("identifier")
	orgID, err := uuid.FromString(identifier)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidID", I18nInvalidID, nil)
		return nil, err
	}

	org, err := h.orgService.GetOrganizationByID(c.Request.Context(), orgID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "NotFound", I18nNotFound, nil)
		return nil, err
	}

	return org, nil
}

// getOrgBySlug lấy org bằng slug (cho public routes)
func (h *Handler) getOrgBySlug(c *gin.Context) (*domain.Organization, error) {
	slug := c.Param("identifier")

	org, err := h.orgService.GetOrganizationBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			response.Error(c, http.StatusNotFound, "NotFound", I18nNotFound, nil)
			return nil, err
		}
		h.handleError(c, err, I18nGetFailed)
		return nil, err
	}

	return org, nil
}

func (h *Handler) handleError(c *gin.Context, err error, fallbackI18n string) {
	if appErr, ok := pkgerrors.AsAppError(err); ok {
		response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
		return
	}
	if err == pgx.ErrNoRows {
		response.Error(c, http.StatusNotFound, "NotFound", I18nNotFound, nil)
		return
	}
	h.logger.Error("Handler error", zap.Error(err))
	response.Error(c, http.StatusInternalServerError, "InternalError", fallbackI18n, nil)
}

// Mapping functions

func mapToOrganizationResponse(org *domain.Organization) orgdto.OrganizationResponse {
	resp := orgdto.OrganizationResponse{
		ID:                    org.ID.String(),
		Name:                  org.Name,
		Slug:                  org.Slug,
		Status:                string(org.Status),
		AvatarURL:             org.AvatarURL,
		IsRecruiting:          org.IsRecruiting,
		MemberCount:           org.MemberCount,
		CompletedTranslations: org.CompletedTranslations,
		CreatedAt:             org.CreatedAt.Format(timeutil.ISO8601Layout),
	}

	if len(org.Description) > 0 {
		resp.Description = (*json.RawMessage)(&org.Description)
	}

	return resp
}

func mapToOrganizationDetailResponse(org *domain.Organization, myRole *string, canInvite bool, hasReported bool) orgdto.OrganizationDetailResponse {
	resp := orgdto.OrganizationDetailResponse{
		OrganizationResponse: mapToOrganizationResponse(org),
		MyRole:               myRole,
		CanInvite:            canInvite,
		HasReported:          hasReported,
	}

	return resp
}

func mapToMemberResponse(m *domain.OrganizationMembership) orgdto.MemberResponse {
	resp := orgdto.MemberResponse{
		UserID: m.UserID.String(),
		Role:   string(m.Role),
		Status: m.Status,
	}

	if m.User != nil {
		if m.User.DisplayName != nil {
			resp.DisplayName = *m.User.DisplayName
		}
		resp.Username = m.User.Username
		resp.AvatarURL = m.User.AvatarURL
	}

	if m.JoinedAt != nil {
		joinedAt := m.JoinedAt.Format(timeutil.ISO8601Layout)
		resp.JoinedAt = &joinedAt
	}

	return resp
}

func mapToPendingInviteResponse(inv *domain.OrganizationPendingInvite) orgdto.PendingInviteResponse {
	resp := orgdto.PendingInviteResponse{
		ID:        inv.ID.String(),
		UserID:    inv.UserID.String(),
		InvitedBy: inv.InvitedBy.String(),
		ExpiresAt: inv.ExpiresAt.Format(timeutil.ISO8601Layout),
		CreatedAt: inv.CreatedAt.Format(timeutil.ISO8601Layout),
	}

	if inv.User != nil {
		if inv.User.DisplayName != nil {
			resp.DisplayName = *inv.User.DisplayName
		}
		resp.Username = inv.User.Username
		resp.AvatarURL = inv.User.AvatarURL
	}

	if inv.Inviter != nil && inv.Inviter.DisplayName != nil {
		resp.InviterName = *inv.Inviter.DisplayName
	}

	return resp
}

func mapToReportResponse(r *domain.OrganizationReport) orgdto.ReportResponse {
	resp := orgdto.ReportResponse{
		ID:          r.ID.String(),
		ReporterID:  r.ReporterID.String(),
		Reason:      r.Reason,
		Description: r.Description,
		OrgResponse: r.OrgResponse,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Format(timeutil.ISO8601Layout),
	}

	if r.Reporter != nil && r.Reporter.DisplayName != nil {
		resp.ReporterName = *r.Reporter.DisplayName
	}

	if r.OrgRespondedBy != nil {
		respondedBy := r.OrgRespondedBy.String()
		resp.OrgRespondedBy = &respondedBy
	}

	if r.OrgRespondedAt != nil {
		respondedAt := r.OrgRespondedAt.Format(timeutil.ISO8601Layout)
		resp.OrgRespondedAt = &respondedAt
	}

	return resp
}

// GetTopOrgsByViews returns top organizations by view count
// @Summary Get top organizations by views
// @Description Get organizations with highest view counts for a calendar-based time period
// @Tags Organizations
// @Produce json
// @Param period query string false "Time period (day, week, month, year)" default(week)
// @Param offset query int false "0 = current period, 1 = previous period" default(0)
// @Param limit query int false "Limit (default 10)" default(10)
// @Success 200 {object} response.StandardResponse{data=[]orgdto.OrganizationResponse}
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/organizations/top [get]
func (h *Handler) GetTopOrgsByViews(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	// Validate period
	validPeriods := map[string]bool{"day": true, "week": true, "month": true, "year": true}
	if !validPeriods[period] {
		period = "week"
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 && o <= 52 {
		offset = o
	}

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	orgs, err := h.analyticsSvc.GetTopOrgsByViews(c.Request.Context(), period, offset, limit)
	if err != nil {
		h.logger.Error("Failed to get top orgs by views", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "GET_TOP_ORGS_ERROR", I18nListFailed, nil)
		return
	}

	// Map to response
	orgResponses := make([]orgdto.OrganizationResponse, len(orgs))
	for i, org := range orgs {
		orgResponses[i] = mapToOrganizationResponse(org)
	}

	response.Success(c, http.StatusOK, I18nListSuccess, orgResponses, nil)
}

