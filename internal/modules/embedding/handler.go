package embedding

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	embeddingdto "system/internal/dto/embedding"
	"system/internal/platform/cache"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

const (
	similarCacheTTL    = 10 * time.Minute
	similarCachePrefix = "embedding:similar:"
)

// ContentFetcher interface to fetch content data by type
type ContentFetcher interface {
	GetNovelsByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.Novel, error)
	// Future: GetAnimesByIDs, GetMangasByIDs
}

// Handler handles embedding-related HTTP requests
type Handler struct {
	service        *Service
	contentFetcher ContentFetcher
	cache          *cache.CacheService
}

// NewHandler creates a new embedding Handler
func NewHandler(service *Service, contentFetcher ContentFetcher, cacheService *cache.CacheService) *Handler {
	return &Handler{
		service:        service,
		contentFetcher: contentFetcher,
		cache:          cacheService,
	}
}

// GetSimilarContent finds similar content based on embedding similarity
// @Summary Get similar content (novels, anime, manga)
// @Tags Embeddings
// @Produce json
// @Param type query string true "Content type: novel, anime, manga"
// @Param id query string true "Content ID"
// @Param limit query int false "Limit" default(10)
// @Success 200 {object} response.StandardResponse{data=[]embeddingdto.SimilarContentResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/embeddings/similar [get]
func (h *Handler) GetSimilarContent(c *gin.Context) {
	contentType := c.Query("type")
	idStr := c.Query("id")

	// Validate type
	if contentType == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_TYPE", "embedding.missing_type", nil)
		return
	}

	// Currently only novel is supported
	if contentType != "novel" {
		response.Error(c, http.StatusBadRequest, "UNSUPPORTED_TYPE", "embedding.unsupported_type", nil)
		return
	}

	// Validate ID
	if idStr == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_ID", "embedding.missing_id", nil)
		return
	}

	contentID, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "embedding.invalid_id", nil)
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Cache key: embedding:similar:novel:<id>:<limit>
	cacheKey := fmt.Sprintf("%s%s:%s:%d", similarCachePrefix, contentType, idStr, limit)

	// Try cache first
	results, err := cache.GetOrSet(c.Request.Context(), h.cache, cacheKey, similarCacheTTL, func() ([]embeddingdto.SimilarContentResponse, error) {
		return h.fetchSimilarNovels(c.Request.Context(), contentID, limit)
	})

	if err != nil {
		response.Error(c, http.StatusNotFound, "EMBEDDING_NOT_FOUND", "embedding.not_found", nil)
		return
	}

	response.Success(c, http.StatusOK, "embedding.similar_found", results, nil)
}

// fetchSimilarNovels fetches similar novels from database (used by cache)
func (h *Handler) fetchSimilarNovels(ctx context.Context, contentID uuid.UUID, limit int) ([]embeddingdto.SimilarContentResponse, error) {
	similarNovels, err := h.service.FindSimilarNovels(ctx, contentID, limit)
	if err != nil {
		return nil, err
	}

	if len(similarNovels) == 0 {
		return []embeddingdto.SimilarContentResponse{}, nil
	}

	// Get novel IDs from similar results
	novelIDs := make([]uuid.UUID, len(similarNovels))
	distanceMap := make(map[uuid.UUID]float32)
	for i, sn := range similarNovels {
		novelIDs[i] = sn.NovelID
		distanceMap[sn.NovelID] = sn.Distance
	}

	// Fetch novel details
	novels, err := h.contentFetcher.GetNovelsByIDs(ctx, novelIDs)
	if err != nil {
		return nil, err
	}

	// Map to response
	results := make([]embeddingdto.SimilarContentResponse, 0, len(novels))
	for _, novel := range novels {
		results = append(results, embeddingdto.SimilarContentResponse{
			ID:            novel.ID.String(),
			Title:         novel.Title,
			Slug:          novel.Slug,
			CoverImageURL: novel.CoverImageURL,
			Distance:      distanceMap[novel.ID],
			Type:          "novel",
		})
	}

	return results, nil
}

// GetEmbeddingStatus checks if content has embedding
// @Summary Check embedding status
// @Tags Embeddings
// @Produce json
// @Param type query string true "Content type: novel"
// @Param id query string true "Content ID"
// @Success 200 {object} response.StandardResponse{data=embeddingdto.EmbeddingStatusResponse}
// @Router /api/v1/embeddings/status [get]
func (h *Handler) GetEmbeddingStatus(c *gin.Context) {
	contentType := c.Query("type")
	idStr := c.Query("id")

	if contentType != "novel" || idStr == "" {
		response.Error(c, http.StatusBadRequest, "INVALID_PARAMS", "embedding.invalid_params", nil)
		return
	}

	contentID, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "embedding.invalid_id", nil)
		return
	}

	embedding, err := h.service.GetEmbedding(c.Request.Context(), contentID)
	if err != nil || embedding == nil {
		response.Success(c, http.StatusOK, "embedding.status_found", embeddingdto.EmbeddingStatusResponse{
			HasEmbedding: false,
		}, nil)
		return
	}

	createdAt := embedding.CreatedAt.Format(timeutil.ISO8601Layout)
	response.Success(c, http.StatusOK, "embedding.status_found", embeddingdto.EmbeddingStatusResponse{
		HasEmbedding: true,
		ModelVersion: &embedding.ModelVersion,
		CreatedAt:    &createdAt,
	}, nil)
}
