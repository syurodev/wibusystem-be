// ============================================================================
// Embedding Repository
// ============================================================================
//
// Repository này quản lý vector embeddings cho novels sử dụng pgvector.
// Dùng để tìm các novels tương tự (similar recommendations).
//
// Operations:
//   - Upsert: Tạo hoặc cập nhật embedding cho novel
//   - GetByNovelID: Lấy embedding theo novel ID
//   - FindSimilar: Tìm các novels tương tự dựa trên cosine distance
//
// SQL queries được load từ thư mục queries/ sử dụng go:embed.
//
// ============================================================================

package embedding

import (
	"context"
	_ "embed"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

//go:embed queries/upsert.sql
var upsertQuery string

//go:embed queries/find_similar.sql
var findSimilarQuery string

//go:embed queries/get_by_novel_id.sql
var getByNovelIDQuery string

// Repository provides database operations for novel embeddings.
type Repository interface {
	Upsert(ctx context.Context, novelID uuid.UUID, embedding pgvector.Vector, modelVersion, sourceHash string) error
	GetByNovelID(ctx context.Context, novelID uuid.UUID) (*NovelEmbedding, error)
	FindSimilar(ctx context.Context, embedding pgvector.Vector, excludeNovelID uuid.UUID, limit int) ([]SimilarNovel, error)
}

type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new embedding repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// Upsert inserts or updates an embedding for a novel.
func (r *repository) Upsert(ctx context.Context, novelID uuid.UUID, embedding pgvector.Vector, modelVersion, sourceHash string) error {
	_, err := r.pool.Exec(ctx, upsertQuery, novelID, embedding, modelVersion, sourceHash)
	return err
}

// GetByNovelID retrieves an embedding by novel ID.
func (r *repository) GetByNovelID(ctx context.Context, novelID uuid.UUID) (*NovelEmbedding, error) {
	rows, err := r.pool.Query(ctx, getByNovelIDQuery, novelID)
	if err != nil {
		return nil, err
	}

	embedding, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[NovelEmbedding])
	if err != nil {
		return nil, err
	}

	return &embedding, nil
}

// FindSimilar finds novels similar to the given embedding.
func (r *repository) FindSimilar(ctx context.Context, embedding pgvector.Vector, excludeNovelID uuid.UUID, limit int) ([]SimilarNovel, error) {
	rows, err := r.pool.Query(ctx, findSimilarQuery, embedding, excludeNovelID, limit)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, pgx.RowToStructByName[SimilarNovel])
}
