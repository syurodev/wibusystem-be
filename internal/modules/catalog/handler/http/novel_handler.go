package http

import (
	"wibusystem/internal/modules/catalog/dto"
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// NovelHandler handles HTTP requests for novels.
type NovelHandler struct {
	novelService service.NovelService
}

// NewNovelHandler creates a new instance of NovelHandler.
func NewNovelHandler(novelService service.NovelService) *NovelHandler {
	return &NovelHandler{novelService: novelService}
}

// CreateNovel handles the creation of a new novel.
func (h *NovelHandler) CreateNovel(c *fiber.Ctx) (any, *response.Pagination, error) {
	req := new(dto.CreateNovelRequest)
	if err := c.BodyParser(req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "cannot parse request")
	}

	// TODO: Add validation middleware

	novel, err := h.novelService.CreateNovel(c.Context(), req.Title, req.Summary, req.AuthorID)
	if err != nil {
		return nil, nil, err
	}

	return novel, nil, nil
}

// GetNovel handles retrieving a single novel by its ID.
func (h *NovelHandler) GetNovel(c *fiber.Ctx) (any, *response.Pagination, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid novel ID format")
	}

	novel, err := h.novelService.GetNovel(c.Context(), id)
	if err != nil {
		return nil, nil, err // Service should return a specific error, e.g., ErrNotFound
	}

	return novel, nil, nil
}

// ListNovels handles listing all novels with pagination.
func (h *NovelHandler) ListNovels(c *fiber.Ctx) (any, *response.Pagination, error) {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)

	novels, total, err := h.novelService.ListNovels(c.Context(), page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}

	return novels, pagination, nil
}
