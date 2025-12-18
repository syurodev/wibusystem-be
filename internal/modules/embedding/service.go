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
