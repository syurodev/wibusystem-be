/*
View Tracking Repository - Redis Operations
============================================

Repository này quản lý tất cả Redis operations cho view tracking system.
Xem thêm: internal/domain/view_tracking.go cho architecture overview.

REDIS DATA STRUCTURES:
----------------------

┌─────────────────────────────────────────────────────────────────────────────────┐
│ 1. DEDUPLICATION KEYS (String với TTL)                                         │
│    Key pattern: view:dedup:{entity_type}:{entity_id}:{identifier}              │
│    Value: "1"                                                                   │
│    TTL: DedupWindowSeconds (default 300s = 5 phút)                             │
│                                                                                 │
│    Example: view:dedup:chapter:550e8400-...:user123 = "1" (TTL 300s)           │
│                                                                                 │
│    Operations:                                                                  │
│    - CheckDeduplication: EXISTS key → true if exists (duplicate)              │
│    - RecordDeduplication: SETNX key "1" EX ttlSeconds                          │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│ 2. VIEW BUFFERS (Hash)                                                          │
│    Keys: view:buffer:chapter, view:buffer:novel                                │
│    Fields: {entity_id} → count                                                 │
│                                                                                 │
│    Example:                                                                     │
│    view:buffer:chapter {                                                        │
│      "550e8400-...": "15",   // chapter A có 15 views pending                  │
│      "6ba7b810-...": "7"     // chapter B có 7 views pending                   │
│    }                                                                            │
│                                                                                 │
│    Operations:                                                                  │
│    - IncrementBuffer: HINCRBY view:buffer:chapter {entity_id} 1                │
│    - GetAllBuffers:   HGETALL + DEL (atomic via pipeline)                      │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│ 3. EVENT QUEUE (List - FIFO)                                                    │
│    Key: view:events                                                             │
│    Values: [event_json, event_json, ...]                                       │
│                                                                                 │
│    Operations:                                                                  │
│    - EnqueueEvent:  RPUSH view:events {event_json}                             │
│    - DequeueEvents: LPOP view:events (repeated N times)                        │
│                                                                                 │
│    Flow: API → RPUSH → [queue] → LPOP → ClickHouse                            │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│ 4. ACTIVITY QUEUE (List - FIFO)                                                 │
│    Key: activity:queue                                                          │
│    Values: [activity_json, activity_json, ...]                                 │
│                                                                                 │
│    Similar to event queue but for content creation activities                   │
└─────────────────────────────────────────────────────────────────────────────────┘
*/

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

	// Activity queue: LIST structure
	// activity:queue -> [activity_json, ...]
	keyActivityQueue = "activity:queue"
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

// PeekBuffers retrieves all buffered counts WITHOUT clearing them.
// Use ClearBuffers after successful processing to remove the data.
//
// Redis operation: HGETALL (read only, no delete)
func (r *viewTrackingRedisRepo) PeekBuffers(ctx context.Context) ([]domain.ViewBuffer, error) {
	var buffers []domain.ViewBuffer

	// Get chapter buffers (read only)
	chapterCounts, err := r.rdb.Client.HGetAll(ctx, keyBufferChapter).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to peek chapter buffers: %w", err)
	}

	for entityIDStr, countStr := range chapterCounts {
		entityID, err := uuid.FromString(entityIDStr)
		if err != nil {
			continue
		}

		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil {
			continue
		}

		buffers = append(buffers, domain.ViewBuffer{
			EntityType: "chapter",
			EntityID:   entityID,
			Count:      count,
		})
	}

	// Get novel buffers (read only)
	novelCounts, err := r.rdb.Client.HGetAll(ctx, keyBufferNovel).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to peek novel buffers: %w", err)
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

// ClearBuffers removes all buffer data from Redis.
// Call this ONLY after successful processing (PostgreSQL updates complete).
//
// Redis operation: DEL view:buffer:chapter view:buffer:novel
func (r *viewTrackingRedisRepo) ClearBuffers(ctx context.Context) error {
	// Delete both buffer keys
	_, err := r.rdb.Client.Del(ctx, keyBufferChapter, keyBufferNovel).Result()
	if err != nil {
		return fmt.Errorf("failed to clear buffers: %w", err)
	}
	return nil
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

// EnqueueActivity adds a content activity to the Activity queue.
// Activity được serialize thành JSON và pushed vào Redis list.
//
// Redis operation: RPUSH activity:queue {activity_json}
//
// Parameters:
//   - ctx: Context
//   - activity: ContentActivity cần enqueue
//
// Returns:
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) EnqueueActivity(ctx context.Context, activity *domain.ContentActivity) error {
	// Marshal activity to JSON
	activityJSON, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("failed to marshal activity: %w", err)
	}

	// RPUSH adds to the tail of the list
	_, err = r.rdb.Client.RPush(ctx, keyActivityQueue, activityJSON).Result()
	if err != nil {
		return fmt.Errorf("failed to enqueue activity: %w", err)
	}

	return nil
}

// DequeueActivities retrieves và removes một batch content activities từ queue.
// Sử dụng LPOP để dequeue từ head của list.
//
// Redis operation: LPOP activity:queue (repeated batchSize times)
//
// Parameters:
//   - ctx: Context
//   - batchSize: Số lượng activities tối đa cần dequeue
//
// Returns:
//   - []*domain.ContentActivity: List activities, nil nếu queue empty
//   - error: Lỗi nếu có
func (r *viewTrackingRedisRepo) DequeueActivities(ctx context.Context, batchSize int) ([]*domain.ContentActivity, error) {
	// LPOP removes from the head of the list
	// Note: LPop with count is available in Redis 6.2+
	var activityJSONs []string

	for i := 0; i < batchSize; i++ {
		result, err := r.rdb.Client.LPop(ctx, keyActivityQueue).Result()
		if err != nil {
			if err == redis.Nil {
				// Queue is empty
				break
			}
			return nil, fmt.Errorf("failed to dequeue activity: %w", err)
		}
		activityJSONs = append(activityJSONs, result)
	}

	// Nếu không có activities, return nil
	if len(activityJSONs) == 0 {
		return nil, nil
	}

	// Unmarshal activities
	var activities []*domain.ContentActivity
	for _, activityJSON := range activityJSONs {
		var activity domain.ContentActivity
		if err := json.Unmarshal([]byte(activityJSON), &activity); err != nil {
			// Log error but continue processing other activities
			continue
		}
		activities = append(activities, &activity)
	}

	return activities, nil
}

// PeekEvents reads N events from queue WITHOUT removing them.
// Use LRANGE to read from head of list without modifying.
//
// Redis operation: LRANGE view:events 0 (batchSize-1)
func (r *viewTrackingRedisRepo) PeekEvents(ctx context.Context, batchSize int) ([]*domain.ViewEvent, error) {
	// LRANGE reads elements without removing (0-indexed, inclusive end)
	eventJSONs, err := r.rdb.Client.LRange(ctx, keyEventQueue, 0, int64(batchSize-1)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to peek events: %w", err)
	}

	if len(eventJSONs) == 0 {
		return nil, nil
	}

	// Unmarshal events
	var events []*domain.ViewEvent
	for _, eventJSON := range eventJSONs {
		var event domain.ViewEvent
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// AcknowledgeEvents removes N events from the head of queue after successful processing.
// Use LTRIM to remove processed elements from head.
//
// Redis operation: LTRIM view:events N -1 (keeps elements from index N to end)
func (r *viewTrackingRedisRepo) AcknowledgeEvents(ctx context.Context, count int) error {
	if count <= 0 {
		return nil
	}

	// LTRIM keeps elements from 'count' to end, effectively removing first 'count' elements
	_, err := r.rdb.Client.LTrim(ctx, keyEventQueue, int64(count), -1).Result()
	if err != nil {
		return fmt.Errorf("failed to acknowledge events: %w", err)
	}
	return nil
}

// PeekActivities reads N activities from queue WITHOUT removing them.
//
// Redis operation: LRANGE activity:queue 0 (batchSize-1)
func (r *viewTrackingRedisRepo) PeekActivities(ctx context.Context, batchSize int) ([]*domain.ContentActivity, error) {
	activityJSONs, err := r.rdb.Client.LRange(ctx, keyActivityQueue, 0, int64(batchSize-1)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to peek activities: %w", err)
	}

	if len(activityJSONs) == 0 {
		return nil, nil
	}

	var activities []*domain.ContentActivity
	for _, activityJSON := range activityJSONs {
		var activity domain.ContentActivity
		if err := json.Unmarshal([]byte(activityJSON), &activity); err != nil {
			continue
		}
		activities = append(activities, &activity)
	}

	return activities, nil
}

// AcknowledgeActivities removes N activities from the head of queue after successful processing.
//
// Redis operation: LTRIM activity:queue N -1
func (r *viewTrackingRedisRepo) AcknowledgeActivities(ctx context.Context, count int) error {
	if count <= 0 {
		return nil
	}

	_, err := r.rdb.Client.LTrim(ctx, keyActivityQueue, int64(count), -1).Result()
	if err != nil {
		return fmt.Errorf("failed to acknowledge activities: %w", err)
	}
	return nil
}
