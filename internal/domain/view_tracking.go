package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// ViewEvent represents a single view tracking event for analytics.
// This struct is used to store view events in ClickHouse for detailed analytics.
type ViewEvent struct {
	EventTime  time.Time
	EntityType string // "novel" or "chapter"
	EntityID   uuid.UUID
	NovelID    uuid.UUID
	ChapterID  *uuid.UUID // Nullable - only set when viewing a chapter
	UserID     *uuid.UUID // Nullable - only set for authenticated users
	IPAddress  string
	ViewCount  uint32
}

// ViewBuffer represents buffered view counts awaiting sync to PostgreSQL.
// This struct is used to store aggregated counts from Redis before bulk updating PostgreSQL.
type ViewBuffer struct {
	EntityType string    // "novel" or "chapter"
	EntityID   uuid.UUID
	Count      int64
}

// ViewTrackingRepository định nghĩa interface cho Redis-based view tracking operations.
// Repository này xử lý deduplication, buffering, và event queuing cho view tracking system.
type ViewTrackingRepository interface {
	// CheckDeduplication kiểm tra xem view từ identifier này đã được ghi nhận gần đây chưa.
	// Returns true nếu là duplicate (nên skip), false nếu là unique (nên count).
	//
	// Parameters:
	//   - ctx: Context
	//   - entityType: Loại entity ("novel" hoặc "chapter")
	//   - entityID: ID của entity
	//   - identifier: Unique identifier (user_id hoặc IP address)
	//
	// Returns:
	//   - bool: true nếu duplicate, false nếu unique
	//   - error: Lỗi nếu có
	CheckDeduplication(ctx context.Context, entityType string, entityID uuid.UUID, identifier string) (bool, error)

	// RecordDeduplication đánh dấu rằng view đã được count (set Redis key với TTL).
	// Key sẽ tự động expire sau ttlSeconds.
	//
	// Parameters:
	//   - ctx: Context
	//   - entityType: Loại entity ("novel" hoặc "chapter")
	//   - entityID: ID của entity
	//   - identifier: Unique identifier (user_id hoặc IP address)
	//   - ttlSeconds: Time-to-live in seconds (deduplication window)
	//
	// Returns:
	//   - error: Lỗi nếu có
	RecordDeduplication(ctx context.Context, entityType string, entityID uuid.UUID, identifier string, ttlSeconds int) error

	// IncrementBuffer atomically increments view counter buffer trong Redis.
	// Sử dụng HINCRBY cho atomic increment.
	//
	// Parameters:
	//   - ctx: Context
	//   - entityType: Loại entity ("novel" hoặc "chapter")
	//   - entityID: ID của entity
	//
	// Returns:
	//   - error: Lỗi nếu có
	IncrementBuffer(ctx context.Context, entityType string, entityID uuid.UUID) error

	// GetAllBuffers retrieves tất cả buffered counts và clears them atomically.
	// Method này được gọi bởi background worker để sync sang PostgreSQL.
	// Sử dụng Redis pipeline để ensure atomic read-and-clear.
	//
	// Parameters:
	//   - ctx: Context
	//
	// Returns:
	//   - []ViewBuffer: List các buffered counts
	//   - error: Lỗi nếu có
	GetAllBuffers(ctx context.Context) ([]ViewBuffer, error)

	// EnqueueEvent adds một view event vào ClickHouse queue (Redis List).
	// Events sẽ được consumed bởi background worker và inserted vào ClickHouse.
	//
	// Parameters:
	//   - ctx: Context
	//   - event: ViewEvent cần enqueue
	//
	// Returns:
	//   - error: Lỗi nếu có
	EnqueueEvent(ctx context.Context, event *ViewEvent) error

	// DequeueEvents retrieves và removes một batch events từ queue.
	// Returns nil nếu queue empty.
	//
	// Parameters:
	//   - ctx: Context
	//   - batchSize: Số lượng events tối đa cần dequeue
	//
	// Returns:
	//   - []*ViewEvent: List events, nil nếu queue empty
	//   - error: Lỗi nếu có
	DequeueEvents(ctx context.Context, batchSize int) ([]*ViewEvent, error)
}

// ViewAnalyticsRepository định nghĩa interface cho ClickHouse analytics operations.
// Repository này xử lý batch insert events và analytics queries.
type ViewAnalyticsRepository interface {
	// BatchInsertEvents inserts multiple view events vào ClickHouse.
	// Sử dụng ClickHouse batch API để tối ưu performance.
	//
	// Parameters:
	//   - ctx: Context
	//   - events: List events cần insert
	//
	// Returns:
	//   - error: Lỗi nếu có
	BatchInsertEvents(ctx context.Context, events []*ViewEvent) error

	// GetViewStats retrieves aggregated view statistics cho một entity trong time range.
	// Query này aggregate data từ ClickHouse để provide analytics.
	//
	// Parameters:
	//   - ctx: Context
	//   - entityType: Loại entity ("novel" hoặc "chapter")
	//   - entityID: ID của entity
	//   - from: Start time của range
	//   - to: End time của range
	//
	// Returns:
	//   - map[string]any: Map chứa statistics (total_views, unique_users, etc.)
	//   - error: Lỗi nếu có
	GetViewStats(ctx context.Context, entityType string, entityID uuid.UUID, from, to time.Time) (map[string]any, error)
}
