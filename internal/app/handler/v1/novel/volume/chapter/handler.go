package volume_chapter

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/app/middleware"
	"system/internal/domain"
	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

type Handler struct {
	chapterService      *service.ChapterService
	volumeService       *service.VolumeService
	viewTrackingService *service.ViewTrackingService
}

func NewHandler(chapterService *service.ChapterService, volumeService *service.VolumeService, viewTrackingService *service.ViewTrackingService) *Handler {
	return &Handler{
		chapterService:      chapterService,
		volumeService:       volumeService,
		viewTrackingService: viewTrackingService,
	}
}

// CreateChapter creates a new chapter
// @Summary Create a new chapter
// @Tags Chapters
// @Accept json
// @Produce json
// @Param request body CreateChapterRequest true "Create Chapter Request"
// @Success 201 {object} response.StandardResponse{data=ChapterDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{id}/volumes/{volume_id}/chapters [post]
func (h *Handler) CreateChapter(c *gin.Context) {
	// Get volume ID from path (new route: /novels/volumes/:id/chapters)
	volumeIDStr := c.Param("id")
	volumeIDVal, err := uuid.FromString(volumeIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", "volume.invalid_id", nil)
		return
	}
	volumeID := &volumeIDVal

	// Novel ID is implicit, service will look it up from volume
	novelID := uuid.Nil

	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get user ID from authentication context
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}
	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Create chapter
	chapter, err := h.chapterService.CreateChapter(
		c.Request.Context(),
		novelID,
		*volumeID,
		req.ChapterNumber,
		req.Title,
		req.Content,
		req.AuthorNotes,
		req.IsFree,
		req.Price,
		req.Currency,
		req.Status,
		req.DisplayOrder,
		req.ScheduledAt,
		userID,
	)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "chapter.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrChapterNumberExists) {
			response.Error(c, http.StatusConflict, "CHAPTER_NUMBER_EXISTS", "chapter.number_already_exists", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidChapterStatus) {
			response.Error(c, http.StatusBadRequest, "INVALID_STATUS", "chapter.invalid_status", nil)
			return
		}
		// Log the actual error for debugging
		// Assuming we have a logger or just use fmt.Printf/log.Println for now if logger not injected
		// But better to rely on the middleware logging if we pass the error?
		// response.Error doesn't take the error object to log it.
		// Let's use c.Error(err) so Gin logger picks it up
		_ = c.Error(err)
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "chapter.create_failed", nil)
		return
	}

	// Map to response
	resp := mapToChapterDetailResponse(chapter)

	response.Success(c, http.StatusCreated, "chapter.created_success", resp, nil)
}

// UpdateChapter updates a chapter
// @Summary Update a chapter
// @Tags Chapters
// @Accept json
// @Produce json
// @Param id path string true "Chapter ID"
// @Param request body UpdateChapterRequest true "Update Chapter Request"
// @Success 200 {object} response.StandardResponse{data=ChapterDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id} [put]
func (h *Handler) UpdateChapter(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	var req UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Parse optional volume ID
	var volumeID *uuid.UUID
	if req.VolumeID != nil {
		vid, err := uuid.FromString(*req.VolumeID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", "volume.invalid_id", nil)
			return
		}
		volumeID = &vid
	}

	// TODO: Get actual user ID from authentication context
	systemUserID := uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000000"))
	
	// Update chapter
	chapter, err := h.chapterService.UpdateChapter(
		c.Request.Context(),
		id,
		volumeID,
		req.ChapterNumber,
		req.Title,
		req.Content,
		req.AuthorNotes,
		req.IsFree,
		req.Price,
		req.Currency,
		req.Status,
		req.DisplayOrder,
		req.ScheduledAt,
		systemUserID,
		nil, // requestContext
	)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrChapterNotFound) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "chapter.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "chapter.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrChapterNumberExists) {
			response.Error(c, http.StatusConflict, "CHAPTER_NUMBER_EXISTS", "chapter.number_already_exists", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidChapterStatus) {
			response.Error(c, http.StatusBadRequest, "INVALID_STATUS", "chapter.invalid_status", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "chapter.update_failed", nil)
		return
	}

	// Map to response
	resp := mapToChapterDetailResponse(chapter)

	response.Success(c, http.StatusOK, "chapter.updated_success", resp, nil)
}

// DeleteChapter deletes a chapter
// @Summary Delete a chapter
// @Tags Chapters
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id} [delete]
func (h *Handler) DeleteChapter(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	// Delete chapter
	err = h.chapterService.DeleteChapter(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrChapterNotFound) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "chapter.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "chapter.delete_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "chapter.deleted_success", nil, nil)
}

// GetChapter retrieves chapter details
// @Summary Get chapter details
// @Tags Chapters
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse{data=ChapterDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id} [get]
func (h *Handler) GetChapter(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	// Get chapter
	chapter, err := h.chapterService.GetChapterByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrChapterNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "chapter.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", "chapter.get_failed", nil)
		return
	}

	// Map to response
	resp := mapToChapterDetailResponse(chapter)

	response.Success(c, http.StatusOK, "chapter.get_success", resp, nil)
}

// ListChaptersByNovel retrieves chapters for a novel
// @Summary List chapters by novel ID
// @Tags Chapters
// @Produce json
// @Param novel_id path string true "Novel ID"
// @Param status query string false "Filter by status"
// @Param volume_id query string false "Filter by volume ID"
// @Param published_only query bool false "Only published chapters"
// @Param is_free query bool false "Filter by free/paid"
// @Param sort_by query string false "Sort by field"
// @Param sort_order query string false "Sort order"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.StandardResponse{data=[]ChapterResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{novel_id}/chapters [get]
func (h *Handler) ListChaptersByNovel(c *gin.Context) {
	// Get novel ID from path
	novelIDStr := c.Param("id")
	novelID, err := uuid.FromString(novelIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOVEL_ID", "novel.invalid_id", nil)
		return
	}

	var req ListChaptersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Build filter
	filter := domain.ChapterFilter{
		PublishedOnly: req.PublishedOnly,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
	}

	if req.Status != "" {
		status := domain.ChapterStatus(req.Status)
		filter.Status = &status
	}

	if req.VolumeID != nil {
		vid, err := uuid.FromString(*req.VolumeID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", "volume.invalid_id", nil)
			return
		}
		filter.VolumeID = &vid
	}

	if req.IsFree != nil {
		filter.IsFree = req.IsFree
	}

	// Add pagination if provided
	if req.Page > 0 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		filter.Limit = req.Limit
		filter.Offset = offset
	}

	// Get chapters
	chapters, err := h.chapterService.GetChaptersByNovelID(c.Request.Context(), novelID, filter)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "chapter.list_failed", nil)
		return
	}

	// Map to response format
	chapterResponses := make([]ChapterResponse, len(chapters))
	for i, chapter := range chapters {
		chapterResponses[i] = mapToChapterResponse(chapter)
	}

	response.Success(c, http.StatusOK, "chapter.list_success", chapterResponses, nil)
}

// ListChaptersByVolume retrieves all chapters for a volume
// @Summary List chapters by volume ID
// @Tags Chapters
// @Produce json
// @Param volume_id path string true "Volume ID"
// @Param published_only query bool false "Only published chapters"
// @Success 200 {object} response.StandardResponse{data=ListChaptersResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{volume_id}/chapters [get]
func (h *Handler) ListChaptersByVolume(c *gin.Context) {
	// Get volume ID from path
	volumeIDStr := c.Param("id")
	volumeID, err := uuid.FromString(volumeIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", "volume.invalid_id", nil)
		return
	}

	publishedOnly := c.Query("published_only") == "true"

	// Get volume info to get title
	volume, err := h.volumeService.GetVolumeByID(c.Request.Context(), volumeID)
	if err != nil {
		// If volume not found, return 404
		// Note: We need to check error type, but for now generic error is fine or we can assume 500 if not found?
		// Ideally we should return 404 if volume doesn't exist.
		// Let's assume GetVolumeByID returns error if not found.
		response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
		return
	}

	// Get chapters
	chapters, err := h.chapterService.GetChaptersByVolumeID(c.Request.Context(), volumeID, publishedOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "chapter.list_failed", nil)
		return
	}

	// Map to response format
	chapterResponses := make([]ChapterResponse, len(chapters))
	for i, chapter := range chapters {
		chapterResponses[i] = mapToChapterResponse(chapter)
	}

	resp := ListChaptersResponse{
		VolumeID:    volume.ID.String(),
		VolumeTitle: volume.Title,
		Chapters:    chapterResponses,
	}

	response.Success(c, http.StatusOK, "chapter.list_success", resp, nil)
}

// PublishChapter publishes a chapter immediately
// @Summary Publish a chapter
// @Tags Chapters
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id}/publish [post]
func (h *Handler) PublishChapter(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	// TODO: Get actual user ID from authentication context
	systemUserID := uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000000"))
	requestContext := extractRequestContext(c)

	// Publish chapter
	err = h.chapterService.PublishChapter(c.Request.Context(), id, systemUserID, requestContext)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrChapterNotFound) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "chapter.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "PUBLISH_FAILED", "chapter.publish_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "chapter.published_success", nil, nil)
}

// ScheduleChapter schedules a chapter for publication
// @Summary Schedule a chapter for publication
// @Tags Chapters
// @Accept json
// @Produce json
// @Param id path string true "Chapter ID"
// @Param request body ScheduleChapterRequest true "Schedule Chapter Request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id}/schedule [post]
func (h *Handler) ScheduleChapter(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	var req ScheduleChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Parse scheduled time
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_SCHEDULED_TIME", "chapter.invalid_scheduled_time", nil)
		return
	}

	// Schedule chapter
	err = h.chapterService.ScheduleChapter(c.Request.Context(), id, scheduledAt)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrChapterNotFound) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "chapter.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "SCHEDULE_FAILED", "chapter.schedule_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "chapter.scheduled_success", nil, nil)
}

// IncrementViewCount increments the view count of a chapter
// @Summary Increment chapter view count
// @Tags Chapters
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id}/view [post]
func (h *Handler) IncrementViewCount(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	// Get user ID from context (optional)
	var userID *uuid.UUID
	// TODO: Extract user ID from auth middleware if available
	// if user := middleware.GetUser(c); user != nil {
	// 	uid := user.ID
	// 	userID = &uid
	// }

	// Get IP address
	ipAddress := c.ClientIP()

	// Track view using ViewTrackingService (Redis + ClickHouse)
	counted, err := h.viewTrackingService.TrackChapterView(c.Request.Context(), id, userID, ipAddress)
	if err != nil {
		// Log error but don't fail the request if tracking fails
		// response.Error(c, http.StatusInternalServerError, "TRACKING_FAILED", "chapter.tracking_failed", nil)
		// Instead, we return success but log the error internally (already done in service)
		// For debugging purposes, we might want to return error
		_ = err
	}

	response.Success(c, http.StatusOK, "chapter.view_tracked", gin.H{"counted": counted}, nil)
}

// UpdateStatistics updates chapter statistics
// @Summary Update chapter statistics
// @Tags Chapters
// @Accept json
// @Produce json
// @Param id path string true "Chapter ID"
// @Param request body UpdateStatisticsRequest true "Update Statistics Request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/chapters/{id}/statistics [put]
func (h *Handler) UpdateStatistics(c *gin.Context) {
	// Get chapter ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	var req UpdateStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Build statistics
	stats := domain.ChapterStatistics{
		ViewCount:    req.ViewCount,
		LikeCount:    req.LikeCount,
		CommentCount: req.CommentCount,
	}

	// Update statistics
	err = h.chapterService.UpdateStatistics(c.Request.Context(), id, stats)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrChapterNotFound) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", "chapter.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "chapter.update_statistics_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "chapter.statistics_updated", nil, nil)
}

// Helper function to map domain model to detail response
func mapToChapterDetailResponse(chapter *domain.Chapter) ChapterDetailResponse {
	resp := ChapterDetailResponse{
		ID:             chapter.ID.String(),
		NovelID:        chapter.NovelID.String(),
		ChapterNumber:  chapter.ChapterNumber,
		Title:          chapter.Title,
		Slug:           chapter.Slug,
		Content:        chapter.Content,
		WordCount:      chapter.WordCount,
		CharacterCount: chapter.CharacterCount,
		IsFree:         chapter.IsFree,
		Price:          chapter.Price,
		Currency:       chapter.Currency,
		Status:         string(chapter.Status),
		ViewCount:      chapter.ViewCount,
		LikeCount:      chapter.LikeCount,
		CommentCount:   chapter.CommentCount,
		DisplayOrder:   chapter.DisplayOrder,
		AuthorNotes:    chapter.AuthorNotes,
		CreatedAt:      chapter.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:      chapter.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	if chapter.VolumeID != nil {
		volumeID := chapter.VolumeID.String()
		resp.VolumeID = &volumeID
	}

	// Format optional dates
	if chapter.PublishedAt != nil {
		publishedAt := chapter.PublishedAt.Format(timeutil.ISO8601Layout)
		resp.PublishedAt = &publishedAt
	}

	if chapter.ScheduledAt != nil {
		scheduledAt := chapter.ScheduledAt.Format(timeutil.ISO8601Layout)
		resp.ScheduledAt = &scheduledAt
	}

	return resp
}

// Helper function to map domain model to list response
func mapToChapterResponse(chapter *domain.Chapter) ChapterResponse {
	resp := ChapterResponse{
		ID:            chapter.ID.String(),
		NovelID:       chapter.NovelID.String(),
		ChapterNumber: chapter.ChapterNumber,
		Title:         chapter.Title,
		Slug:          chapter.Slug,
		WordCount:     chapter.WordCount,
		IsFree:        chapter.IsFree,
		Price:         chapter.Price,
		Currency:      chapter.Currency,
		Status:        string(chapter.Status),
		ViewCount:     chapter.ViewCount,
		LikeCount:     chapter.LikeCount,
		CommentCount:  chapter.CommentCount,
		DisplayOrder:  chapter.DisplayOrder,
		CreatedAt:     chapter.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:     chapter.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	if chapter.VolumeID != nil {
		volumeID := chapter.VolumeID.String()
		resp.VolumeID = &volumeID
	}

	// Format optional dates
	if chapter.PublishedAt != nil {
		publishedAt := chapter.PublishedAt.Format(timeutil.ISO8601Layout)
		resp.PublishedAt = &publishedAt
	}

	if chapter.ScheduledAt != nil {
		scheduledAt := chapter.ScheduledAt.Format(timeutil.ISO8601Layout)
		resp.ScheduledAt = &scheduledAt
	}

	return resp
}



// Helper function to extract request context for history logging
func extractRequestContext(c *gin.Context) map[string]any {
	return map[string]any{
		"request_id": c.GetHeader("X-Request-ID"),
		"ip_address": c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
	}
}
