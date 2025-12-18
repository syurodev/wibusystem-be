package embedding

// SimilarContentResponse là response cho API similar content
type SimilarContentResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	Distance      float32 `json:"distance"` // Cosine distance (0 = identical, 2 = opposite)
	Type          string  `json:"type"`     // novel, anime, manga
}

// EmbeddingStatusResponse là response cho API embedding status
type EmbeddingStatusResponse struct {
	HasEmbedding bool    `json:"has_embedding"`
	ModelVersion *string `json:"model_version,omitempty"`
	CreatedAt    *string `json:"created_at,omitempty"`
}
