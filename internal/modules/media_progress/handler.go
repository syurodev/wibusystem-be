/*
Media Progress Handler - HTTP Request Handling Layer
=====================================================

ARCHITECTURE:
─────────────
Handler layer nhận HTTP requests, validate input, gọi Service, và format response.
Sử dụng Gin framework theo pattern của dự án.

ENDPOINTS:
──────────
GET    /api/v1/history           → GetProgressList()   - Lấy danh sách progress (paginated)
GET    /api/v1/history/recent    → GetRecentProgress() - Lấy N mục gần nhất
POST   /api/v1/history           → UpdateProgress()    - Cập nhật progress
DELETE /api/v1/history/:id       → DeleteProgress()    - Xóa progress
POST   /api/v1/history/clear     → ClearAllProgress()  - Xóa toàn bộ

GET    /api/v1/progress/:type/:id/units           → GetUnitProgress()    - Lấy chapter status
POST   /api/v1/progress/:type/:id/units/:uid/complete → MarkUnitComplete() - Đánh dấu đã đọc
*/

package media_progress

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"

	"system/internal/app/middleware"
	"system/internal/domain"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

// Handler handles media progress HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new media progress Handler instance
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// =============================================================================
// REQUEST/RESPONSE TYPES
// =============================================================================

// UpdateProgressRequest - request body for POST /api/v1/history
type UpdateProgressRequest struct {
	ContentID         string            `json:"content_id" binding:"required"`
	MediaType         string            `json:"media_type"` // "novel", "manga", "anime" - nếu không có sẽ tự detect
	LatestUnitID      string            `json:"latest_unit_id" binding:"required"`
	NovelLastReadInfo *NovelPositionDTO `json:"novel_last_read_info,omitempty"`
	MangaLastPageRead *int              `json:"manga_last_page_read,omitempty"`
	AnimeLastTime     *string           `json:"anime_last_episode_time_viewed,omitempty"`
}

// NovelPositionDTO - position info cho novel
type NovelPositionDTO struct {
	NodeID  string `json:"node_id"`
	Preview string `json:"preview"`
}

// MediaProgressResponse - response for a single progress item
type MediaProgressResponse struct {
	ID string `json:"id"`

	// Media info
	Media *MediaInfoResponse `json:"media"`

	// Current unit (chapter/episode)
	LatestUnit *UnitInfoResponse `json:"latest_unit"`

	// Position (based on media type)
	NovelLastReadInfo *NovelPositionDTO `json:"novel_last_read_info,omitempty"`
	MangaLastPageRead *int              `json:"manga_last_page_read,omitempty"`
	AnimeLastTime     *string           `json:"anime_last_episode_time_viewed,omitempty"`

	// Progress stats
	UserProgressPercentage float64 `json:"user_progress_percentage"`
	TotalUnits             int     `json:"total_units,omitempty"`
	CompletedUnits         int     `json:"completed_units,omitempty"`

	// Timestamps
	LastViewedAt     string `json:"last_viewed_at"`
	ContentUpdatedAt string `json:"content_updated_at,omitempty"`
}

// MediaInfoResponse - media info in progress response
type MediaInfoResponse struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Slug      string     `json:"slug"`
	CoverURL  *string    `json:"cover_url"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Genres    []GenreDTO `json:"genres,omitempty"`
	Author    *OwnerDTO  `json:"author,omitempty"`
	Rating    float64    `json:"rating,omitempty"`
	Views     int64      `json:"views,omitempty"`
	Favorites int        `json:"favorites,omitempty"`
}

// UnitInfoResponse - chapter/episode info
type UnitInfoResponse struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Slug   string `json:"slug,omitempty"`
}

// GenreDTO - genre info
type GenreDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// OwnerDTO - owner/author info
type OwnerDTO struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	AvatarURL   *string `json:"avatar_url"`
}

// UnitProgressResponse - response for chapter read status
type UnitProgressResponse struct {
	UnitID      string  `json:"unit_id"`
	Status      string  `json:"status"`
	IsRead      bool    `json:"is_read"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

// =============================================================================
// HANDLER METHODS
// =============================================================================

// GetProgressList lấy danh sách progress với pagination
// GET /api/v1/history
//
// FLOW:
// 1. Get userID từ JWT token
// 2. Parse query parameters (page, limit, type, sort)
// 3. Gọi Service.GetProgressList()
// 4. Map domain entities → response DTOs
// 5. Return với pagination meta
func (h *Handler) GetProgressList(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	// Parse query params
	page := parseIntParam(c, "page", 1)
	limit := parseIntParam(c, "limit", 15)
	mediaType := c.Query("type")
	sortBy := c.DefaultQuery("sort", "recent")

	filter := domain.MediaProgressFilter{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
	if mediaType != "" && mediaType != "all" {
		filter.MediaType = &mediaType
	}
	if sortBy == "recent" {
		filter.SortBy = "last_accessed_at"
		filter.SortOrder = "desc"
	}

	items, total, err := h.service.GetProgressList(c.Request.Context(), userID, filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	// Map to response
	responses := make([]MediaProgressResponse, len(items))
	for i, item := range items {
		responses[i] = mapMediaProgressToResponse(item)
	}

	meta := &response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalItems: int(total),
		TotalPages: (int(total) + limit - 1) / limit,
	}

	response.Success(c, http.StatusOK, I18nGetSuccess, responses, meta)
}

// GetRecentProgress lấy N mục gần nhất cho "Continue" section
// GET /api/v1/history/recent
//
// FLOW:
// 1. Get userID từ JWT token
// 2. Parse limit từ query (default 12)
// 3. Gọi Service.GetRecentProgress()
// 4. Map → response
func (h *Handler) GetRecentProgress(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	limit := parseIntParam(c, "limit", 12)

	items, err := h.service.GetRecentProgress(c.Request.Context(), userID, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	// Map to response
	responses := make([]MediaProgressResponse, len(items))
	for i, item := range items {
		responses[i] = mapMediaProgressToResponse(item)
	}

	response.Success(c, http.StatusOK, I18nGetSuccess, responses, nil)
}

// UpdateProgress cập nhật progress khi user đọc/xem
// POST /api/v1/history
//
// FLOW:
// 1. Get userID từ JWT token
// 2. Parse request body
// 3. Xác định media_type (nếu không có sẽ mặc định là novel)
// 4. Parse position data theo media_type
// 5. Gọi Service.UpdateProgress()
// 6. Return updated progress
func (h *Handler) UpdateProgress(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nValidationFailed, err.Error())
		return
	}

	// Parse IDs
	mediaID, err := uuid.FromString(req.ContentID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CONTENT_ID", I18nInvalidID, nil)
		return
	}

	unitID, err := uuid.FromString(req.LatestUnitID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_UNIT_ID", I18nInvalidID, nil)
		return
	}

	// Determine media type
	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = domain.MediaTypeNovel // default
	}

	// Parse position based on media type
	var position any
	switch mediaType {
	case domain.MediaTypeNovel:
		if req.NovelLastReadInfo != nil {
			position = domain.NovelPosition{
				NodeID:  req.NovelLastReadInfo.NodeID,
				Preview: req.NovelLastReadInfo.Preview,
			}
		}
	case domain.MediaTypeManga:
		if req.MangaLastPageRead != nil {
			position = domain.MangaPosition{Page: *req.MangaLastPageRead}
		}
	case domain.MediaTypeAnime:
		if req.AnimeLastTime != nil {
			position = domain.AnimePosition{Time: *req.AnimeLastTime}
		}
	}

	input := UpdateProgressInput{
		UserID:    userID,
		MediaType: mediaType,
		MediaID:   mediaID,
		UnitID:    unitID,
		Position:  position,
	}

	progress, err := h.service.UpdateProgress(c.Request.Context(), input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", I18nUpdateFailed, nil)
		return
	}

	resp := mapMediaProgressToResponse(progress)
	response.Success(c, http.StatusOK, I18nUpdateSuccess, resp, nil)
}

// DeleteProgress xóa progress
// DELETE /api/v1/history/:id
func (h *Handler) DeleteProgress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	err = h.service.DeleteProgress(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", I18nDeleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nDeleteSuccess, nil, nil)
}

// ClearAllProgress xóa toàn bộ progress
// POST /api/v1/history/clear
func (h *Handler) ClearAllProgress(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	err = h.service.ClearAllProgress(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "CLEAR_FAILED", I18nClearFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nClearSuccess, nil, nil)
}

// GetUnitProgress lấy trạng thái đọc của chapters
// GET /api/v1/progress/:media_type/:media_id/units
//
// FLOW:
// 1. Get userID
// 2. Parse media_type và media_id từ path
// 3. Gọi Service.GetUnitProgress()
// 4. Return list với status mỗi chapter
func (h *Handler) GetUnitProgress(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	mediaType := c.Param("media_type")
	mediaIDStr := c.Param("media_id")
	mediaID, err := uuid.FromString(mediaIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MEDIA_ID", I18nInvalidID, nil)
		return
	}

	items, err := h.service.GetUnitProgress(c.Request.Context(), userID, mediaType, mediaID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	// Map to response
	responses := make([]UnitProgressResponse, len(items))
	for i, item := range items {
		responses[i] = UnitProgressResponse{
			UnitID: item.UnitID.String(),
			Status: string(item.Status),
			IsRead: item.Status == domain.UnitStatusCompleted,
		}
		if item.CompletedAt != nil {
			completedAt := item.CompletedAt.Format(timeutil.ISO8601Layout)
			responses[i].CompletedAt = &completedAt
		}
	}

	response.Success(c, http.StatusOK, I18nGetSuccess, responses, nil)
}

// MarkUnitComplete đánh dấu chapter đã đọc xong
// POST /api/v1/progress/:media_type/:media_id/units/:unit_id/complete
func (h *Handler) MarkUnitComplete(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	mediaType := c.Param("media_type")
	mediaIDStr := c.Param("media_id")
	unitIDStr := c.Param("unit_id")

	mediaID, err := uuid.FromString(mediaIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_MEDIA_ID", I18nInvalidID, nil)
		return
	}

	unitID, err := uuid.FromString(unitIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_UNIT_ID", I18nInvalidID, nil)
		return
	}

	err = h.service.MarkUnitComplete(c.Request.Context(), userID, mediaType, mediaID, unitID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "MARK_COMPLETE_FAILED", I18nMarkCompleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nMarkCompleteSuccess, nil, nil)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func parseIntParam(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	var result int
	if _, err := c.GetQuery(key); err {
		return defaultVal
	}
	// Simple parsing
	for _, ch := range val {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		}
	}
	if result == 0 {
		return defaultVal
	}
	return result
}

func mapMediaProgressToResponse(mp *domain.MediaProgress) MediaProgressResponse {
	resp := MediaProgressResponse{
		ID:                     mp.ID.String(),
		UserProgressPercentage: mp.ProgressPercentage,
		TotalUnits:             mp.TotalUnits,
		CompletedUnits:         mp.CompletedUnits,
		LastViewedAt:           mp.LastAccessedAt.Format(timeutil.ISO8601Layout),
	}

	// Map Media info
	if mp.Media != nil {
		resp.Media = &MediaInfoResponse{
			ID:       mp.Media.ID.String(),
			Title:    mp.Media.Title,
			Slug:     mp.Media.Slug,
			CoverURL: mp.Media.CoverURL,
			Type:     mp.MediaType,
			Status:   mp.Media.Status,
		}
		if mp.Media.OwnerDisplayName != nil {
			resp.Media.Author = &OwnerDTO{
				ID:          mp.Media.OwnerID.String(),
				DisplayName: *mp.Media.OwnerDisplayName,
			}
			if mp.Media.OwnerUsername != nil {
				resp.Media.Author.Username = *mp.Media.OwnerUsername
			}
			resp.Media.Author.AvatarURL = mp.Media.OwnerAvatarURL
		}
	}

	// Map CurrentUnit info
	if mp.CurrentUnit != nil {
		resp.LatestUnit = &UnitInfoResponse{
			ID:     mp.CurrentUnit.ID.String(),
			Number: mp.CurrentUnit.Number,
			Title:  mp.CurrentUnit.Title,
			Slug:   mp.CurrentUnit.Slug,
		}
	}

	// Map Position based on media type
	// TODO: Parse mp.Position JSON and populate appropriate field

	return resp
}
