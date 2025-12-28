package embedding

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/pgvector/pgvector-go"
)

// Repository provides database operations for novel embeddings.
type Repository interface {
	Upsert(ctx context.Context, novelID uuid.UUID, embedding pgvector.Vector, modelVersion, sourceHash string) error
	GetByNovelID(ctx context.Context, novelID uuid.UUID) (*NovelEmbedding, error)
	FindSimilar(ctx context.Context, embedding pgvector.Vector, excludeNovelID uuid.UUID, limit int) ([]SimilarNovel, error)
}
