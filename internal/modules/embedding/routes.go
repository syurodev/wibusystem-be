package embedding

import "github.com/gin-gonic/gin"

// RegisterRoutes registers embedding API routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// GET /api/v1/embeddings/similar?type=novel&id=xxx&limit=10
	router.GET("/similar", h.GetSimilarContent)

	// GET /api/v1/embeddings/status?type=novel&id=xxx
	router.GET("/status", h.GetEmbeddingStatus)
}
