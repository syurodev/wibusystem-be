package http

import (
	"wibusystem/internal/modules/catalog/dto"
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// VolumeHandler handles HTTP requests for volumes.
type VolumeHandler struct {
	volumeService service.VolumeService
}

// NewVolumeHandler creates a new instance of VolumeHandler.
func NewVolumeHandler(volumeService service.VolumeService) *VolumeHandler {
	return &VolumeHandler{volumeService: volumeService}
}

// CreateVolume handles the creation of a new volume.
func (h *VolumeHandler) CreateVolume(c *fiber.Ctx) (any, *response.Pagination, error) {
	req := new(dto.CreateVolumeRequest)
	if err := c.BodyParser(req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "cannot parse request")
	}

	volume, err := h.volumeService.CreateVolume(c.Context(), req.NovelID, req.Title, req.VolumeNumber)
	if err != nil {
		return nil, nil, err
	}

	return volume, nil, nil
}

// GetVolume handles retrieving a single volume by its ID.
func (h *VolumeHandler) GetVolume(c *fiber.Ctx) (any, *response.Pagination, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid volume ID format")
	}

	volume, err := h.volumeService.GetVolume(c.Context(), id)
	if err != nil {
		return nil, nil, err
	}

	return volume, nil, nil
}

// ListVolumesByNovel handles listing all volumes for a novel.
func (h *VolumeHandler) ListVolumesByNovel(c *fiber.Ctx) (any, *response.Pagination, error) {
	novelID, err := uuid.Parse(c.Params("novelId"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid novel ID format")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)

	volumes, total, err := h.volumeService.ListVolumesByNovel(c.Context(), novelID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}

	return volumes, pagination, nil
}
