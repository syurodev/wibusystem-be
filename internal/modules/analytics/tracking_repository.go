package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"

	"system/internal/domain"
	"system/internal/platform/database"
)

// Redis key patterns cho view tracking system
const (
	// Deduplication keys: view:dedup:chapter:{chapter_id}:{user_id_or_ip}
	// TTL = DedupWindowSeconds (default 300s = 5 phút)
	keyDedupPrefix = "view:dedup:"

	// Buffer keys: HASH structure
	// view:buffer:chapter -> {chapter_id: count}
	// view:buffer:novel -> {novel_id: count}
	keyBufferChapter = "view:buffer:chapter"
	keyBufferNovel   = "view:buffer:novel"

	// Event queue: LIST structure
	// view:events -> [event1_json, event2_json, ...]
	keyEventQueue = "view:events"
)

// viewTrackingRedisRepo implements ViewTrackingRepository interface.
// Xử lý tất cả Redis operations cho view tracking system.
type viewTrackingRedisRepo struct {
	rdb *database.RedisClient
}

// NewViewTrackingRedisRepository tạo instance mới của view tracking Redis repository.
//
// Parameters:
//   - rdb: Redis client instance
//
// Returns:
//   - domain.ViewTrackingRepository: Repository implementation
func NewViewTrackingRedisRepository(rdb *database.RedisClient) domain.ViewTrackingRepository {
	return &viewTrackingRedisRepo{rdb: rdb}
}

// CheckDeduplication kiểm tra xem view có phải duplicate không.
// Sử dụng Redis EXISTS command để check key existence.
//
// Redis key format: view:dedup:chapter:{chapter_id}:{user_id_or_ip}
//
// Parameters:
//   - ctx: Context
//   - entityType: "novel" hoặc "chapter"
//   - entityID: UUID của entity
//   - identifier: User ID hoặc IP address
//
// Returns:
//   - bool: true nếu duplicate (key exists), false nếu unique (key not exists)
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) CheckDeduplication(ctx context.Context, entityType string, entityID uuid.UUID, identifier string) (bool, error) {
	// Tạo Redis key
	key := fmt.Sprintf("%s%s:%s:%s", keyDedupPrefix, entityType, entityID.String(), identifier)

	// EXISTS returns 1 if key exists, 0 otherwise
	exists, err := r.rdb.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check deduplication: %w", err)
	}

	return exists > 0, nil
}

// RecordDeduplication đánh dấu view đã được counted bằng cách set Redis key với TTL.
// Sử dụng SetNX để ensure atomic operation (chỉ set nếu key chưa exists).
//
// Redis operation: SETNX key "1" EX ttlSeconds
//
// Parameters:
//   - ctx: Context
//   - entityType: "novel" hoặc "chapter"
//   - entityID: UUID của entity
//   - identifier: User ID hoặc IP address
//   - ttlSeconds: Time-to-live in seconds (deduplication window)
//
// Returns:
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) RecordDeduplication(ctx context.Context, entityType string, entityID uuid.UUID, identifier string, ttlSeconds int) error {
	// Tạo Redis key
	key := fmt.Sprintf("%s%s:%s:%s", keyDedupPrefix, entityType, entityID.String(), identifier)

	// SETNX with expiration - only sets if key doesn't exist
	// Returns true if set, false if already exists
	_, err := r.rdb.Client.SetNX(ctx, key, "1", time.Duration(ttlSeconds)*time.Second).Result()
	if err != nil {
		return fmt.Errorf("failed to record deduplication: %w", err)
	}

	return nil
}

// IncrementBuffer atomically increments view counter trong Redis hash.
// Sử dụng HINCRBY cho atomic increment operation.
//
// Redis operation: HINCRBY view:buffer:chapter {chapter_id} 1
//
// Parameters:
//   - ctx: Context
//   - entityType: "novel" hoặc "chapter"
//   - entityID: UUID của entity
//
// Returns:
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) IncrementBuffer(ctx context.Context, entityType string, entityID uuid.UUID) error {
	// Chọn hash key dựa trên entity type
	hashKey := keyBufferChapter
	if entityType == "novel" {
		hashKey = keyBufferNovel
	}

	field := entityID.String()

	// HINCRBY atomically increments the hash field
	// Nếu field không tồn tại, Redis sẽ tạo mới với giá trị 0 trước khi increment
	_, err := r.rdb.Client.HIncrBy(ctx, hashKey, field, 1).Result()
	if err != nil {
		return fmt.Errorf("failed to increment buffer for %s %s: %w", entityType, field, err)
	}

	return nil
}

// GetAllBuffers retrieves tất cả buffered counts và clears them atomically.
// Sử dụng Redis pipeline để ensure atomicity của read-and-clear operation.
//
// Redis operations (in pipeline):
//  1. HGETALL view:buffer:chapter
//  2. DEL view:buffer:chapter
//  3. HGETALL view:buffer:novel
//  4. DEL view:buffer:novel
//
// Parameters:
//   - ctx: Context
//
// Returns:
//   - []domain.ViewBuffer: List các buffered counts
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) GetAllBuffers(ctx context.Context) ([]domain.ViewBuffer, error) {
	var buffers []domain.ViewBuffer

	// Sử dụng pipeline để atomic read-and-clear
	pipe := r.rdb.Client.Pipeline()

	// Get chapter buffers
	chapterCmd := pipe.HGetAll(ctx, keyBufferChapter)
	pipe.Del(ctx, keyBufferChapter)

	// Get novel buffers
	novelCmd := pipe.HGetAll(ctx, keyBufferNovel)
	pipe.Del(ctx, keyBufferNovel)

	// Execute pipeline atomically
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get and clear buffers: %w", err)
	}

	// Process chapter counts
	chapterCounts, err := chapterCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get chapter buffers: %w", err)
	}

	for entityIDStr, countStr := range chapterCounts {
		entityID, err := uuid.FromString(entityIDStr)
		if err != nil {
			// Skip invalid UUIDs
			continue
		}

		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			// Skip invalid counts
			continue
		}

		buffers = append(buffers, domain.ViewBuffer{
			EntityType: "chapter",
			EntityID:   entityID,
			Count:      count,
		})
	}

	// Process novel counts
	novelCounts, err := novelCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get novel buffers: %w", err)
	}

	for entityIDStr, countStr := range novelCounts {
		entityID, err := uuid.FromString(entityIDStr)
		if err != nil {
			continue
		}

		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			continue
		}

		buffers = append(buffers, domain.ViewBuffer{
			EntityType: "novel",
			EntityID:   entityID,
			Count:      count,
		})
	}

	return buffers, nil
}

// EnqueueEvent adds một view event vào ClickHouse queue.
// Event được serialize thành JSON và pushed vào Redis list.
//
// Redis operation: RPUSH view:events {event_json}
//
// Parameters:
//   - ctx: Context
//   - event: ViewEvent cần enqueue
//
// Returns:
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) EnqueueEvent(ctx context.Context, event *domain.ViewEvent) error {
	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// RPUSH adds to the tail of the list
	_, err = r.rdb.Client.RPush(ctx, keyEventQueue, eventJSON).Result()
	if err != nil {
		return fmt.Errorf("failed to enqueue event: %w", err)
	}

	return nil
}

// DequeueEvents retrieves và removes một batch events từ queue.
// Sử dụng LPOP để dequeue từ head của list.
//
// Redis operation: LPOP view:events (repeated batchSize times)
//
// Parameters:
//   - ctx: Context
//   - batchSize: Số lượng events tối đa cần dequeue
//
// Returns:
//   - []*domain.ViewEvent: List events, nil nếu queue empty
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) DequeueEvents(ctx context.Context, batchSize int) ([]*domain.ViewEvent, error) {
	// LPOP removes from the head of the list
	// Note: LPop with count is available in Redis 6.2+
	var eventJSONs []string

	for i := 0; i < batchSize; i++ {
		result, err := r.rdb.Client.LPop(ctx, keyEventQueue).Result()
		if err != nil {
			if err == redis.Nil {
				// Queue is empty
				break
			}
			return nil, fmt.Errorf("failed to dequeue event: %w", err)
		}
		eventJSONs = append(eventJSONs, result)
	}

	// Nếu không có events, return nil
	if len(eventJSONs) == 0 {
		return nil, nil
	}

	// Unmarshal events
	var events []*domain.ViewEvent
	for _, eventJSON := range eventJSONs {
		var event domain.ViewEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			// Log error but continue processing other events
			// In production, you'd want to send malformed events to a DLQ
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}
