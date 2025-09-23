package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/services/interfaces"
)

// CharacterHandler handles character management endpoints
type CharacterHandler struct {
	characterService interfaces.CharacterServiceInterface
	loc              *i18n.Translator
}

// NewCharacterHandler creates a new character handler
func NewCharacterHandler(characterService interfaces.CharacterServiceInterface, translator *i18n.Translator) *CharacterHandler {
	return &CharacterHandler{
		characterService: characterService,
		loc:              translator,
	}
}

// ListCharacters handles GET /characters
func (h *CharacterHandler) ListCharacters(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind query parameters
	var req d.ListCharactersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid query parameters",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get characters through service
	characters, total, err := h.characterService.ListCharacters(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "list")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Characters fetched successfully",
		Data:    characters,
		Error:   nil,
		Meta: map[string]interface{}{
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
			"total_items": total,
		},
	})
}

// GetCharacter handles GET /characters/:id
func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse character ID
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid character ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Character ID must be a valid UUID"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get character through service
	character, err := h.characterService.GetCharacterByID(ctx, characterID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "get")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Character fetched successfully",
		Data:    character,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// CreateCharacter handles POST /characters
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind request body
	var req d.CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid request body",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Create character through service
	character, err := h.characterService.CreateCharacter(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "create")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: "Character created successfully",
		Data:    character,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateCharacter handles PUT /characters/:id
func (h *CharacterHandler) UpdateCharacter(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse character ID
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid character ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Character ID must be a valid UUID"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind request body
	var req d.UpdateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid request body",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update character through service
	character, err := h.characterService.UpdateCharacter(ctx, characterID, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "update")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Character updated successfully",
		Data:    character,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteCharacter handles DELETE /characters/:id
func (h *CharacterHandler) DeleteCharacter(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse character ID
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid character ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Character ID must be a valid UUID"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Delete character through service
	err = h.characterService.DeleteCharacter(ctx, characterID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "delete")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Character deleted successfully",
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

