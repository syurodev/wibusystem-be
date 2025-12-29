// ============================================================================
// Embedding Service
// ============================================================================
//
// Service này cung cấp business logic cho Embedding module.
// Quản lý vector embeddings để tìm similar novels.
//
// Queue Operations:
//   - QueueNovelForEmbedding: Đẩy novel ID vào Redis queue
//   - PopPendingNovelIDs: Lấy novel IDs từ queue để xử lý
//
// Embedding Operations:
//   - GenerateAndStoreEmbedding: Tạo embedding từ novel metadata và lưu vào DB
//   - FindSimilarNovels: Tìm novels tương tự dựa trên vector similarity
//   - GetEmbedding: Lấy embedding của novel
//
// Architecture:
//   - Uses Hugot (local ONNX model) hoặc NoOp embedder
//   - Redis queue cho background processing
//   - pgvector cho vector storage và similarity search
//
// ============================================================================

package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gofrs/uuid/v5"
	"github.com/pgvector/pgvector-go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"system/configs"
)

// Embedder interface for generating embeddings locally.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Close() error
}

// Service provides embedding operations for novels.
type Service struct {
	repo     Repository
	embedder Embedder
	redis    *redis.Client
	logger   *zap.Logger
	config   *configs.EmbeddingConfig
}

// NewService creates a new embedding service.
func NewService(
	repo Repository,
	embedder Embedder,
	redisClient *redis.Client,
	logger *zap.Logger,
	config *configs.EmbeddingConfig,
) *Service {
	return &Service{
		repo:     repo,
		embedder: embedder,
		redis:    redisClient,
		logger:   logger,
		config:   config,
	}
}

// QueueNovelForEmbedding adds a novel ID to the Redis queue for background processing.
func (s *Service) QueueNovelForEmbedding(ctx context.Context, novelID uuid.UUID) error {
	return s.redis.LPush(ctx, s.config.RedisQueueKey, novelID.String()).Err()
}

// PopPendingNovelIDs pops novel IDs from the Redis queue.
// DEPRECATED: Use PeekPendingNovelIDs + AcknowledgePendingNovelIDs for safer processing.
func (s *Service) PopPendingNovelIDs(ctx context.Context, count int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for i := 0; i < count; i++ {
		result, err := s.redis.RPop(ctx, s.config.RedisQueueKey).Result()
		if err == redis.Nil {
			break // Queue is empty
		}
		if err != nil {
			return ids, err
		}

		id, err := uuid.FromString(result)
		if err != nil {
			s.logger.Warn("Invalid UUID in embedding queue", zap.String("value", result))
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// PeekPendingNovelIDs reads novel IDs from the queue WITHOUT removing them.
// Use with AcknowledgePendingNovelIDs after successful processing.
func (s *Service) PeekPendingNovelIDs(ctx context.Context, count int) ([]uuid.UUID, error) {
	// LRANGE reads from tail (since we LPUSH to add items)
	// Queue works as: LPUSH (add) -> [...] <- RPOP (consume)
	// So we need to read from end: LRANGE key -count -1
	results, err := s.redis.LRange(ctx, s.config.RedisQueueKey, int64(-count), -1).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var ids []uuid.UUID
	for _, result := range results {
		id, err := uuid.FromString(result)
		if err != nil {
			s.logger.Warn("Invalid UUID in embedding queue", zap.String("value", result))
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// AcknowledgePendingNovelIDs removes N novel IDs from the queue after successful processing.
// Call this after successfully processing IDs returned by PeekPendingNovelIDs.
func (s *Service) AcknowledgePendingNovelIDs(ctx context.Context, count int) error {
	if count <= 0 {
		return nil
	}
	// LTRIM keeps elements from 0 to (len-count-1), removing last 'count' elements
	// Since we peeked from tail, we trim from tail
	_, err := s.redis.LTrim(ctx, s.config.RedisQueueKey, 0, int64(-count-1)).Result()
	return err
}

// GenerateAndStoreEmbedding generates and stores an embedding for a novel.
func (s *Service) GenerateAndStoreEmbedding(ctx context.Context, input *NovelEmbeddingInput) error {
	text := input.BuildEmbeddingText()
	sourceHash := hashText(text)

	// Generate embedding using local model
	vector, err := s.embedder.Embed(ctx, text)
	if err != nil {
		return err
	}

	// Store in database
	embedding := pgvector.NewVector(vector)
	return s.repo.Upsert(ctx, input.NovelID, embedding, s.config.ModelName, sourceHash)
}

// FindSimilarNovels finds novels similar to the given novel.
func (s *Service) FindSimilarNovels(ctx context.Context, novelID uuid.UUID, limit int) ([]SimilarNovel, error) {
	// Get the embedding for the source novel
	embedding, err := s.repo.GetByNovelID(ctx, novelID)
	if err != nil {
		return nil, err
	}

	// Find similar novels
	return s.repo.FindSimilar(ctx, embedding.Embedding, novelID, limit)
}

// GetEmbedding retrieves an embedding by novel ID.
func (s *Service) GetEmbedding(ctx context.Context, novelID uuid.UUID) (*NovelEmbedding, error) {
	return s.repo.GetByNovelID(ctx, novelID)
}

// hashText computes SHA256 hash of the input text.
func hashText(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
