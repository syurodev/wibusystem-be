package http

import (
	"wibusystem/internal/modules/catalog/dto"
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GenreHandler handles HTTP requests for genres.
type GenreHandler struct {
	genreService service.GenreService
}

// NewGenreHandler creates a new instance of GenreHandler.
func NewGenreHandler(genreService service.GenreService) *GenreHandler {
	return &GenreHandler{genreService: genreService}
}

// CreateGenre handles the creation of a new genre.
func (h *GenreHandler) CreateGenre(c *fiber.Ctx) (any, *response.Pagination, error) {
	req := new(dto.CreateGenreRequest)
	if err := c.BodyParser(req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "cannot parse request")
	}

	genre, err := h.genreService.CreateGenre(c.Context(), req.Name, req.Description)
	if err != nil {
		return nil, nil, err
	}

	return genre, nil, nil
}

// GetGenre handles retrieving a single genre by its ID.
func (h *GenreHandler) GetGenre(c *fiber.Ctx) (any, *response.Pagination, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid genre ID format")
	}

	genre, err := h.genreService.GetGenre(c.Context(), id)
	if err != nil {
		return nil, nil, err
	}

	return genre, nil, nil
}

// ListGenres handles listing all genres with pagination.
func (h *GenreHandler) ListGenres(c *fiber.Ctx) (any, *response.Pagination, error) {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	genres, total, err := h.genreService.ListGenres(c.Context(), page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}

	return genres, pagination, nil
}
