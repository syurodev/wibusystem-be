// Package http contains HTTP handlers for the Identity module.
package http

import (
	"log"
	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/repository"

	"wibusystem/internal/modules/identity/dto"
	"wibusystem/internal/modules/identity/handler/middleware"
	"wibusystem/internal/modules/identity/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantHandler handles tenant-related HTTP requests.
type TenantHandler struct {
	tenantService service.TenantService
}

// NewTenantHandler creates a new TenantHandler instance.
func NewTenantHandler(tenantService service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// CreateTenant creates a new tenant.
// POST /api/v1/tenants
func (h *TenantHandler) CreateTenant(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	var req dto.CreateTenantRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Context()
	tenant, err := h.tenantService.CreateTenant(ctx, userID, req.Name, req.Slug, req.Description)
	if err != nil {
		log.Printf("[TenantHandler.CreateTenant] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to create tenant",
			fiber.StatusBadRequest,
		).WithCode("CREATE_TENANT_FAILED")
	}

	tenantResponse := mapTenantToResponse(tenant)
	return c.Status(fiber.StatusCreated).JSON(dto.CreateTenantResponse{
		Tenant:  tenantResponse,
		Message: "Tenant created successfully",
	})
}

// GetTenant returns a tenant by ID.
// GET /api/v1/tenants/:tenantId
func (h *TenantHandler) GetTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Invalid tenant ID format",
			fiber.StatusBadRequest,
		)
	}

	ctx := c.Context()
	tenant, err := h.tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		log.Printf("[TenantHandler.GetTenant] Error: %v", err)
		return middleware.NewAppError(
			middleware.ErrNotFound,
			"Tenant not found",
			fiber.StatusNotFound,
		).WithCode("TENANT_NOT_FOUND")
	}

	tenantResponse := mapTenantToResponse(tenant)
	return c.Status(fiber.StatusOK).JSON(dto.GetTenantResponse{
		Tenant: tenantResponse,
	})
}

// UpdateTenant updates a tenant.
// PUT /api/v1/tenants/:tenantId
func (h *TenantHandler) UpdateTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Invalid tenant ID format",
			fiber.StatusBadRequest,
		)
	}

	var req dto.UpdateTenantRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Context()
	tenant, err := h.tenantService.UpdateTenant(ctx, tenantID, req.Name, req.Description)
	if err != nil {
		log.Printf("[TenantHandler.UpdateTenant] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to update tenant",
			fiber.StatusBadRequest,
		).WithCode("UPDATE_TENANT_FAILED")
	}

	tenantResponse := mapTenantToResponse(tenant)
	return c.Status(fiber.StatusOK).JSON(dto.UpdateTenantResponse{
		Tenant:  tenantResponse,
		Message: "Tenant updated successfully",
	})
}

// DeleteTenant deletes a tenant (soft delete).
// DELETE /api/v1/tenants/:tenantId
func (h *TenantHandler) DeleteTenant(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Invalid tenant ID format",
			fiber.StatusBadRequest,
		)
	}

	var req dto.DeleteTenantRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	if !req.ConfirmDeletion {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Tenant deletion must be confirmed",
			fiber.StatusBadRequest,
		).WithCode("DELETION_NOT_CONFIRMED")
	}

	ctx := c.Context()
	if err := h.tenantService.DeleteTenant(ctx, tenantID); err != nil {
		log.Printf("[TenantHandler.DeleteTenant] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to delete tenant",
			fiber.StatusBadRequest,
		).WithCode("DELETE_TENANT_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(dto.DeleteTenantResponse{
		Message: "Tenant deleted successfully",
	})
}

// ListTenants returns a paginated list of tenants.
// GET /api/v1/tenants
func (h *TenantHandler) ListTenants(c *fiber.Ctx) error {
	var req dto.ListTenantsRequest
	if err := middleware.ValidateQuery(c, &req); err != nil {
		return err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := repository.TenantListFilter{
		Limit:        pageSize,
		Offset:       (page - 1) * pageSize,
		NameContains: req.Search,
	}

	if req.Status != "" {
		status := domain.TenantStatus(req.Status)
		filter.Status = &status
	}

	ctx := c.Context()
	tenants, total, err := h.tenantService.ListTenants(ctx, filter)
	if err != nil {
		log.Printf("[TenantHandler.ListTenants] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to list tenants",
			fiber.StatusInternalServerError,
		).WithCode("LIST_TENANTS_FAILED")
	}

	tenantResponses := make([]dto.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResponses[i] = mapTenantToResponse(tenant)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.Status(fiber.StatusOK).JSON(dto.ListTenantsResponse{
		Tenants:    tenantResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetUserTenants returns all tenants the authenticated user is a member of.
// GET /api/v1/tenants/my-tenants
func (h *TenantHandler) GetUserTenants(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	ctx := c.Context()
	tenants, err := h.tenantService.GetUserTenants(ctx, userID)
	if err != nil {
		log.Printf("[TenantHandler.GetUserTenants] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to get user tenants",
			fiber.StatusInternalServerError,
		).WithCode("GET_USER_TENANTS_FAILED")
	}

	tenantResponses := make([]dto.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResponses[i] = mapTenantToResponse(tenant)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"tenants": tenantResponses,
		"total":   len(tenantResponses),
	})
}

// GetUserOwnedTenants returns all tenants owned by the authenticated user.
// GET /api/v1/tenants/my-owned-tenants
func (h *TenantHandler) GetUserOwnedTenants(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	ctx := c.Context()
	tenants, err := h.tenantService.GetUserOwnedTenants(ctx, userID)
	if err != nil {
		log.Printf("[TenantHandler.GetUserOwnedTenants] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to get owned tenants",
			fiber.StatusInternalServerError,
		).WithCode("GET_OWNED_TENANTS_FAILED")
	}

	tenantResponses := make([]dto.TenantResponse, len(tenants))
	for i, tenant := range tenants {
		tenantResponses[i] = mapTenantToResponse(tenant)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"tenants": tenantResponses,
		"total":   len(tenantResponses),
	})
}

// AddMember adds a member to a tenant.
// POST /api/v1/tenants/:tenantId/members
func (h *TenantHandler) AddMember(c *fiber.Ctx) error {
	inviterID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Invalid tenant ID format",
			fiber.StatusBadRequest,
		)
	}

	var req dto.AddTenantMemberRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Context()
	member, err := h.tenantService.AddMember(ctx, tenantID, req.UserID, domain.MemberRole(req.Role), inviterID)
	if err != nil {
		log.Printf("[TenantHandler.AddMember] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to add member",
			fiber.StatusBadRequest,
		).WithCode("ADD_MEMBER_FAILED")
	}

	memberResponse := mapMemberToResponse(member)
	return c.Status(fiber.StatusCreated).JSON(dto.AddTenantMemberResponse{
		Member:  memberResponse,
		Message: "Member added successfully",
	})
}

// UpdateMember updates a member's role in a tenant.
// PUT /api/v1/tenants/:tenantId/members/:userId
func (h *TenantHandler) UpdateMember(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid tenant ID format", fiber.StatusBadRequest)
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid user ID format", fiber.StatusBadRequest)
	}

	var req dto.UpdateTenantMemberRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Context()
	err = h.tenantService.UpdateMemberRole(ctx, tenantID, userID, domain.MemberRole(req.Role))
	if err != nil {
		log.Printf("[TenantHandler.UpdateMember] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to update member",
			fiber.StatusBadRequest,
		).WithCode("UPDATE_MEMBER_FAILED")
	}

	// Refetch member to return updated data
	member, err := h.tenantService.GetMember(ctx, tenantID, userID)
	if err != nil {
		return middleware.NewAppError(err, "Failed to retrieve updated member", fiber.StatusInternalServerError)
	}

	memberResponse := mapMemberToResponse(member)
	return c.Status(fiber.StatusOK).JSON(dto.UpdateTenantMemberResponse{
		Member:  memberResponse,
		Message: "Member updated successfully",
	})
}

// RemoveMember removes a member from a tenant.
// DELETE /api/v1/tenants/:tenantId/members/:userId
func (h *TenantHandler) RemoveMember(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid tenant ID format", fiber.StatusBadRequest)
	}

	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid user ID format", fiber.StatusBadRequest)
	}

	ctx := c.Context()
	if err := h.tenantService.RemoveMember(ctx, tenantID, userID); err != nil {
		log.Printf("[TenantHandler.RemoveMember] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to remove member",
			fiber.StatusBadRequest,
		).WithCode("REMOVE_MEMBER_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(dto.RemoveTenantMemberResponse{
		Message: "Member removed successfully",
	})
}

// ListMembers returns all members of a tenant.
// GET /api/v1/tenants/:tenantId/members
func (h *TenantHandler) ListMembers(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid tenant ID format", fiber.StatusBadRequest)
	}

	var req dto.ListTenantMembersRequest
	if err := middleware.ValidateQuery(c, &req); err != nil {
		return err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	ctx := c.Context()
	members, total, err := h.tenantService.ListMembersPaginated(ctx, tenantID, pageSize, offset)
	if err != nil {
		log.Printf("[TenantHandler.ListMembers] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to list members",
			fiber.StatusInternalServerError,
		).WithCode("LIST_MEMBERS_FAILED")
	}

	memberResponses := make([]dto.TenantMemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = mapMemberToResponse(member)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.Status(fiber.StatusOK).JSON(dto.ListTenantMembersResponse{
		Members:    memberResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// TransferOwnership transfers ownership of a tenant to another member.
// POST /api/v1/tenants/:tenantId/transfer-ownership
func (h *TenantHandler) TransferOwnership(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid tenant ID format", fiber.StatusBadRequest)
	}

	var req dto.TransferOwnershipRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Context()
	err = h.tenantService.TransferOwnership(ctx, tenantID, req.NewOwnerID)
	if err != nil {
		log.Printf("[TenantHandler.TransferOwnership] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to transfer ownership",
			fiber.StatusBadRequest,
		).WithCode("TRANSFER_OWNERSHIP_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Ownership transferred successfully",
	})
}

// GetTenantStats returns statistics about a tenant.
// GET /api/v1/tenants/:tenantId/stats
func (h *TenantHandler) GetTenantStats(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid tenant ID format", fiber.StatusBadRequest)
	}

	ctx := c.Context()
	stats, err := h.tenantService.GetTenantStats(ctx, tenantID)
	if err != nil {
		log.Printf("[TenantHandler.GetTenantStats] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to get tenant stats",
			fiber.StatusInternalServerError,
		).WithCode("GET_TENANT_STATS_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}

// Helper functions

func mapTenantToResponse(tenant *domain.Tenant) dto.TenantResponse {
	return dto.TenantResponse{
		ID:          tenant.ID,
		Name:        tenant.Name,
		Slug:        tenant.Slug,
		Description: tenant.Description,
		OwnerID:     tenant.OwnerID,
		Status:      string(tenant.Status),
		CreatedAt:   tenant.CreatedAt,
		UpdatedAt:   tenant.UpdatedAt,
		DeletedAt:   tenant.DeletedAt,
	}
}

func mapMemberToResponse(member *domain.TenantMember) dto.TenantMemberResponse {
	return dto.TenantMemberResponse{
		ID:       member.ID,
		TenantID: member.TenantID,
		UserID:   member.UserID,
		Role:     string(member.Role),
		User: dto.UserResponse{
			ID: member.UserID,
		},
		JoinedAt:  member.JoinedAt,
		UpdatedAt: member.UpdatedAt,
	}
}
