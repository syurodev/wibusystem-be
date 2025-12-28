// ============================================================================
// Embedding Repository (Ent Implementation)
// ============================================================================
//
// Repository này quản lý vector embeddings cho novels sử dụng Ent + pgvector.
// FindSimilar sử dụng raw SQL vì Ent không hỗ trợ pgvector operators (<=>).
//
// ============================================================================

package embedding

import (
	"context"
	"database/sql"
	"system/internal/platform/database"

	"github.com/gofrs/uuid/v5"
	"github.com/pgvector/pgvector-go"

	ent "system/internal/ent/generated"
	"system/internal/ent/generated/novelembedding"
)

// entRepository implements Repository using Ent
type entRepository struct {
	client *ent.Client
	db     *sql.DB // For raw SQL queries (FindSimilar)
}

// NewEntRepository creates a new embedding repository using Ent.
func NewEntRepository(client *ent.Client, db *sql.DB) Repository {
	return &entRepository{client: client, db: db}
}

// Upsert inserts or updates an embedding for a novel.
func (r *entRepository) Upsert(ctx context.Context, novelID uuid.UUID, embedding pgvector.Vector, modelVersion, sourceHash string) error {
	// Check if exists
	existing, err := database.GetClientFromContext(ctx, r.client).NovelEmbedding.Query().
		Where(novelembedding.NovelIDEQ(novelID)).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if existing != nil {
		// Update existing
		updateBuilder := database.GetClientFromContext(ctx, r.client).NovelEmbedding.UpdateOne(existing).
			SetEmbedding(embedding).
			SetModelVersion(modelVersion)
		if sourceHash != "" {
			updateBuilder.SetSourceHash(sourceHash)
		} else {
			updateBuilder.ClearSourceHash()
		}
		_, err = updateBuilder.Save(ctx)
		return err
	}

	// Create new
	createBuilder := database.GetClientFromContext(ctx, r.client).NovelEmbedding.Create().
		SetNovelID(novelID).
		SetEmbedding(embedding).
		SetModelVersion(modelVersion)
	if sourceHash != "" {
		createBuilder.SetSourceHash(sourceHash)
	}
	_, err = createBuilder.Save(ctx)
	return err
}

// GetByNovelID retrieves an embedding by novel ID.
func (r *entRepository) GetByNovelID(ctx context.Context, novelID uuid.UUID) (*NovelEmbedding, error) {
	e, err := database.GetClientFromContext(ctx, r.client).NovelEmbedding.Query().
		Where(novelembedding.NovelIDEQ(novelID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entEmbeddingToDomain(e), nil
}

// FindSimilar finds novels similar to the given embedding.
// Note: Uses raw SQL because Ent doesn't support pgvector operators (<=>).
func (r *entRepository) FindSimilar(ctx context.Context, embedding pgvector.Vector, excludeNovelID uuid.UUID, limit int) ([]SimilarNovel, error) {
	query := `
		SELECT 
			ne.novel_id,
			ne.embedding <=> $1 AS distance
		FROM catalog.novel_embeddings ne
		JOIN catalog.novels n ON n.id = ne.novel_id
		WHERE ne.novel_id != $2 
		  AND n.deleted_at IS NULL
		ORDER BY distance
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, embedding, excludeNovelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SimilarNovel
	for rows.Next() {
		var sn SimilarNovel
		if err := rows.Scan(&sn.NovelID, &sn.Distance); err != nil {
			return nil, err
		}
		results = append(results, sn)
	}
	return results, rows.Err()
}

// Helper function to convert Ent entity to domain model
func entEmbeddingToDomain(e *ent.NovelEmbedding) *NovelEmbedding {
	return &NovelEmbedding{
		ID:           e.ID,
		NovelID:      e.NovelID,
		Embedding:    e.Embedding,
		ModelVersion: e.ModelVersion,
		SourceHash:   e.SourceHash,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}
