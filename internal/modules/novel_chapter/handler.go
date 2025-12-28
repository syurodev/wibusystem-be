package novel_chapter

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"

	"system/internal/app/middleware"
	"system/internal/domain"
	chapterdto "system/internal/dto/novel_chapter"
	ent "system/internal/ent/generated"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

// Handler handles chapter-related HTTP requests
type Handler struct {
	chapterService         ChapterService
	createChapterUC        CreateChapterUseCase
	updateChapterUC        UpdateChapterUseCase
	deleteChapterUC        DeleteChapterUseCase
	getChapterUC           GetChapterUseCase
	listChaptersByNovelUC  ListChaptersByNovelUseCase
	listChaptersByVolumeUC ListChaptersByVolumeUseCase
	publishChapterUC       PublishChapterUseCase
	scheduleChapterUC      ScheduleChapterUseCase
	viewTrackingSvc        ViewTracker
	updateStatisticsUC     UpdateStatisticsUseCase
}

// NewHandler creates a new chapter Handler instance
func NewHandler(
	chapterService ChapterService,
	createChapterUC CreateChapterUseCase,
	updateChapterUC UpdateChapterUseCase,
	deleteChapterUC DeleteChapterUseCase,
	getChapterUC GetChapterUseCase,
	listChaptersByNovelUC ListChaptersByNovelUseCase,
	listChaptersByVolumeUC ListChaptersByVolumeUseCase,
	publishChapterUC PublishChapterUseCase,
	scheduleChapterUC ScheduleChapterUseCase,
	viewTrackingSvc ViewTracker,
	updateStatisticsUC UpdateStatisticsUseCase,
) *Handler {
	return &Handler{
		chapterService:         chapterService,
		createChapterUC:        createChapterUC,
		updateChapterUC:        updateChapterUC,
		deleteChapterUC:        deleteChapterUC,
		getChapterUC:           getChapterUC,
		listChaptersByNovelUC:  listChaptersByNovelUC,
		listChaptersByVolumeUC: listChaptersByVolumeUC,
		publishChapterUC:       publishChapterUC,
		scheduleChapterUC:      scheduleChapterUC,
		viewTrackingSvc:        viewTrackingSvc,
		updateStatisticsUC:     updateStatisticsUC,
	}
}

// CreateChapter creates a new chapter
func (h *Handler) CreateChapter(c *gin.Context) {
	volumeIDStr := c.Param("identifier")
	volumeIDVal, err := uuid.FromString(volumeIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", I18nVolumeNotFound, nil)
		return
	}

	novelID := uuid.Nil

	var req chapterdto.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

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

	chapter, err := h.createChapterUC.Execute(c.Request.Context(), CreateChapterInput{
		NovelID:        novelID,
		VolumeID:       volumeIDVal,
		ChapterNumber:  req.ChapterNumber,
		Title:          req.Title,
		Content:        req.Content,
		WordCount:      req.WordCount,
		CharacterCount: req.CharacterCount,
		AuthorNotes:    req.AuthorNotes,
		IsFree:         req.IsFree,
		Price:          req.Price,
		Currency:       req.Currency,
		Status:         req.Status,
		DisplayOrder:   req.DisplayOrder,
		ScheduledAt:    req.ScheduledAt,
		CreatedBy:      userID,
	})
	if err != nil {
		fmt.Printf("[DEBUG] CreateChapter failed: %v\n", err)
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		// Check for PostgreSQL duplicate key constraint violation
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {
			response.Error(c, http.StatusConflict, "CHAPTER_NUMBER_EXISTS", I18nChapterNumberExists, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", I18nCreateFailed, nil)
		return
	}

	resp := mapToChapterDetailResponse(chapter)
	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateChapter updates a chapter
// @Summary Update a chapter
// @Tags Chapters
// @Accept json
// @Produce json
// @Param identifier path string true "Chapter ID"
// @Param request body UpdateChapterRequest true "Update Chapter Request"
// @Success 200 {object} response.StandardResponse{data=ChapterDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier} [put]
func (h *Handler) UpdateChapter(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	var req chapterdto.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	var volumeID *uuid.UUID
	if req.VolumeID != nil {
		vid, err := uuid.FromString(*req.VolumeID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", "volume.invalid_id", nil)
			return
		}
		volumeID = &vid
	}

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

	chapter, err := h.updateChapterUC.Execute(c.Request.Context(), UpdateChapterInput{
		ID:             id,
		VolumeID:       volumeID,
		ChapterNumber:  req.ChapterNumber,
		Title:          req.Title,
		Content:        req.Content,
		WordCount:      req.WordCount,
		CharacterCount: req.CharacterCount,
		AuthorNotes:    req.AuthorNotes,
		IsFree:         req.IsFree,
		Price:          req.Price,
		Currency:       req.Currency,
		Status:         req.Status,
		DisplayOrder:   req.DisplayOrder,
		ScheduledAt:    req.ScheduledAt,
		ChangedBy:      userID,
		RequestContext: nil,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", I18nUpdateFailed, nil)
		return
	}

	resp := mapToChapterDetailResponse(chapter)
	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteChapter deletes a chapter
// @Summary Delete a chapter
// @Tags Chapters
// @Produce json
// @Param identifier path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier} [delete]
func (h *Handler) DeleteChapter(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	err = h.deleteChapterUC.Execute(c.Request.Context(), DeleteChapterInput{ID: id})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", I18nDeleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nDeletedSuccess, nil, nil)
}

// GetChapter retrieves chapter details
// @Summary Get chapter details
// @Tags Chapters
// @Produce json
// @Param identifier path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse{data=ChapterDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier} [get]
func (h *Handler) GetChapter(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	chapter, err := h.getChapterUC.Execute(c.Request.Context(), GetChapterInput{ID: id})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		if ent.IsNotFound(err) {
			response.Error(c, http.StatusNotFound, "CHAPTER_NOT_FOUND", I18nNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	resp := mapToChapterDetailResponse(chapter)
	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListChaptersByNovel retrieves chapters for a novel
// @Summary List chapters by novel ID
// @Tags Chapters
// @Produce json
// @Param identifier path string true "Novel ID"
// @Success 200 {object} response.StandardResponse{data=[]ChapterResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{identifier}/chapters [get]
func (h *Handler) ListChaptersByNovel(c *gin.Context) {
	novelIDStr := c.Param("identifier")
	novelID, err := uuid.FromString(novelIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOVEL_ID", "novel.invalid_id", nil)
		return
	}

	var req chapterdto.ListChaptersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	filter := domain.NovelChapterFilter{
		PublishedOnly: req.PublishedOnly,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
	}

	if req.Status != "" {
		status := domain.NovelChapterStatus(req.Status)
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

	if req.Page > 0 && req.Limit > 0 {
		offset := (req.Page - 1) * req.Limit
		filter.Limit = req.Limit
		filter.Offset = offset
	}

	chapters, err := h.listChaptersByNovelUC.Execute(c.Request.Context(), ListChaptersByNovelInput{
		NovelID: novelID,
		Filter:  filter,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "chapter.list_failed", nil)
		return
	}

	chapterResponses := make([]chapterdto.ChapterResponse, len(chapters))
	for i, chapter := range chapters {
		chapterResponses[i] = mapToChapterResponse(chapter)
	}

	response.Success(c, http.StatusOK, "chapter.list_success", chapterResponses, nil)
}

// ListChaptersByVolume retrieves all chapters for a volume
// @Summary List chapters by volume ID
// @Tags Chapters
// @Produce json
// @Param identifier path string true "Volume ID"
// @Success 200 {object} response.StandardResponse{data=[]ChapterResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/volumes/{identifier}/chapters [get]
func (h *Handler) ListChaptersByVolume(c *gin.Context) {
	volumeIDStr := c.Param("identifier")
	volumeID, err := uuid.FromString(volumeIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_VOLUME_ID", "volume.invalid_id", nil)
		return
	}

	publishedOnly := c.Query("published_only") == "true"

	chapters, err := h.listChaptersByVolumeUC.Execute(c.Request.Context(), ListChaptersByVolumeInput{
		VolumeID:      volumeID,
		PublishedOnly: publishedOnly,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "chapter.list_failed", nil)
		return
	}

	chapterResponses := make([]chapterdto.ChapterResponse, len(chapters))
	volumeTitle := ""
	for i, chapter := range chapters {
		chapterResponses[i] = mapToChapterResponse(chapter)
		// Extract volume title from chapter's volume if available
		if volumeTitle == "" && chapter.Volume != nil {
			volumeTitle = chapter.Volume.Title
		}
	}

	// Wrap response to match client expectations
	resp := chapterdto.ListChaptersResponse{
		VolumeID:    volumeIDStr,
		VolumeTitle: volumeTitle,
		Chapters:    chapterResponses,
	}

	response.Success(c, http.StatusOK, "chapter.list_success", resp, nil)
}

// PublishChapter publishes a chapter immediately
// @Summary Publish a chapter
// @Tags Chapters
// @Produce json
// @Param identifier path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier}/publish [post]
func (h *Handler) PublishChapter(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	changedBy := uuid.Nil
	if exists {
		changedBy, _ = uuid.FromString(userIDStr)
	}

	requestContext := extractRequestContext(c)

	err = h.publishChapterUC.Execute(c.Request.Context(), PublishChapterInput{
		ID:             id,
		ChangedBy:      changedBy,
		RequestContext: requestContext,
	})
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
// @Param identifier path string true "Chapter ID"
// @Param request body ScheduleChapterRequest true "Schedule Chapter Request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier}/schedule [post]
func (h *Handler) ScheduleChapter(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	var req chapterdto.ScheduleChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_SCHEDULED_TIME", "chapter.invalid_scheduled_time", nil)
		return
	}

	err = h.scheduleChapterUC.Execute(c.Request.Context(), ScheduleChapterInput{
		ID:          id,
		ScheduledAt: scheduledAt,
	})
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

// IncrementViewCount increments the view count of a chapter using analytics tracking
// @Summary Increment chapter view count
// @Tags Chapters
// @Produce json
// @Param identifier path string true "Chapter ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier}/view [post]
func (h *Handler) IncrementViewCount(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDStr, ok := middleware.GetUserID(c); ok {
		if uid, err := uuid.FromString(userIDStr); err == nil {
			userID = &uid
		}
	}

	// Get client IP address
	ipAddress := c.ClientIP()

	// Track view using ViewTrackingService (Redis buffered + ClickHouse analytics)
	_, err = h.viewTrackingSvc.TrackChapterView(c.Request.Context(), id, userID, ipAddress)
	if err != nil {
		// Log but don't fail - view tracking should not block user experience
		_ = err
	}

	response.Success(c, http.StatusOK, "chapter.view_tracked", nil, nil)
}

// UpdateStatistics updates chapter statistics
// @Summary Update chapter statistics
// @Tags Chapters
// @Accept json
// @Produce json
// @Param identifier path string true "Chapter ID"
// @Param request body UpdateStatisticsRequest true "Update Statistics Request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/chapters/{identifier}/statistics [put]
func (h *Handler) UpdateStatistics(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "chapter.invalid_id", nil)
		return
	}

	var req chapterdto.UpdateStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	stats := domain.NovelChapterStatistics{
		ViewCount:    req.ViewCount,
		LikeCount:    req.LikeCount,
		CommentCount: req.CommentCount,
	}

	err = h.updateStatisticsUC.Execute(c.Request.Context(), UpdateStatisticsInput{
		ID:    id,
		Stats: stats,
	})
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
func mapToChapterDetailResponse(chapter *domain.NovelChapter) chapterdto.ChapterDetailResponse {
	resp := chapterdto.ChapterDetailResponse{
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
func mapToChapterResponse(chapter *domain.NovelChapter) chapterdto.ChapterResponse {
	resp := chapterdto.ChapterResponse{
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
