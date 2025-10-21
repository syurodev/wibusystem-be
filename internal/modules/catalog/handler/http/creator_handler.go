package http

import (
	"wibusystem/internal/modules/catalog/dto"
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreatorHandler handles HTTP requests for creators.
type CreatorHandler struct {
	creatorService service.CreatorService
}

// NewCreatorHandler creates a new instance of CreatorHandler.
func NewCreatorHandler(creatorService service.CreatorService) *CreatorHandler {
	return &CreatorHandler{creatorService: creatorService}
}

// CreateCreator handles the creation of a new creator.
func (h *CreatorHandler) CreateCreator(c *fiber.Ctx) (any, *response.Pagination, error) {
	req := new(dto.CreateCreatorRequest)
	if err := c.BodyParser(req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "cannot parse request")
	}

	creator, err := h.creatorService.CreateCreator(c.Context(), req.Name, req.Bio)
	if err != nil {
		return nil, nil, err
	}

	return creator, nil, nil
}

// GetCreator handles retrieving a single creator by its ID.
func (h *CreatorHandler) GetCreator(c *fiber.Ctx) (any, *response.Pagination, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid creator ID format")
	}

	creator, err := h.creatorService.GetCreator(c.Context(), id)
	if err != nil {
		return nil, nil, err
	}

	return creator, nil, nil
}

// ListCreators handles listing all creators with pagination.
func (h *CreatorHandler) ListCreators(c *fiber.Ctx) (any, *response.Pagination, error) {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	creators, total, err := h.creatorService.ListCreators(c.Context(), page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}

	return creators, pagination, nil
}
