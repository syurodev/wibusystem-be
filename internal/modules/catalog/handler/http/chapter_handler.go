package http

import (
	"wibusystem/internal/modules/catalog/dto"
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ChapterHandler handles HTTP requests for chapters.
type ChapterHandler struct {
	chapterService service.ChapterService
}

// NewChapterHandler creates a new instance of ChapterHandler.
func NewChapterHandler(chapterService service.ChapterService) *ChapterHandler {
	return &ChapterHandler{chapterService: chapterService}
}

// CreateChapter handles the creation of a new chapter.
func (h *ChapterHandler) CreateChapter(c *fiber.Ctx) (any, *response.Pagination, error) {
	req := new(dto.CreateChapterRequest)
	if err := c.BodyParser(req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "cannot parse request")
	}

	chapter, err := h.chapterService.CreateChapter(c.Context(), req.VolumeID, req.Title, req.Content, req.ChapterNumber)
	if err != nil {
		return nil, nil, err
	}

	return chapter, nil, nil
}

// GetChapter handles retrieving a single chapter by its ID.
func (h *ChapterHandler) GetChapter(c *fiber.Ctx) (any, *response.Pagination, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid chapter ID format")
	}

	chapter, err := h.chapterService.GetChapter(c.Context(), id)
	if err != nil {
		return nil, nil, err
	}

	return chapter, nil, nil
}

// ListChaptersByVolume handles listing all chapters for a volume.
func (h *ChapterHandler) ListChaptersByVolume(c *fiber.Ctx) (any, *response.Pagination, error) {
	volumeID, err := uuid.Parse(c.Params("volumeId"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid volume ID format")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	chapters, total, err := h.chapterService.ListChaptersByVolume(c.Context(), volumeID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}

	return chapters, pagination, nil
}
