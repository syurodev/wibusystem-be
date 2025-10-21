package http

import (
	"wibusystem/internal/modules/catalog/dto"
	"wibusystem/internal/modules/catalog/service"
	"wibusystem/internal/platform/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CharacterHandler handles HTTP requests for characters.
type CharacterHandler struct {
	characterService service.CharacterService
}

// NewCharacterHandler creates a new instance of CharacterHandler.
func NewCharacterHandler(characterService service.CharacterService) *CharacterHandler {
	return &CharacterHandler{characterService: characterService}
}

// CreateCharacter handles the creation of a new character.
func (h *CharacterHandler) CreateCharacter(c *fiber.Ctx) (any, *response.Pagination, error) {
	req := new(dto.CreateCharacterRequest)
	if err := c.BodyParser(req); err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "cannot parse request")
	}

	character, err := h.characterService.CreateCharacter(c.Context(), req.NovelID, req.Name, req.Description)
	if err != nil {
		return nil, nil, err
	}

	return character, nil, nil
}

// GetCharacter handles retrieving a single character by its ID.
func (h *CharacterHandler) GetCharacter(c *fiber.Ctx) (any, *response.Pagination, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid character ID format")
	}

	character, err := h.characterService.GetCharacter(c.Context(), id)
	if err != nil {
		return nil, nil, err
	}

	return character, nil, nil
}

// ListCharactersByNovel handles listing all characters for a novel.
func (h *CharacterHandler) ListCharactersByNovel(c *fiber.Ctx) (any, *response.Pagination, error) {
	novelID, err := uuid.Parse(c.Params("novelId"))
	if err != nil {
		return nil, nil, fiber.NewError(fiber.StatusBadRequest, "invalid novel ID format")
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)

	characters, total, err := h.characterService.ListCharactersByNovel(c.Context(), novelID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	pagination := &response.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: (int(total) + pageSize - 1) / pageSize,
	}

	return characters, pagination, nil
}
