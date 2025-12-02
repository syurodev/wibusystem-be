package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	"system/internal/platform/database"
)

// viewAnalyticsClickHouseRepo implements ViewAnalyticsRepository interface.
// Xử lý tất cả ClickHouse operations cho analytics system.
type viewAnalyticsClickHouseRepo struct {
	ch *database.ClickHouseClient
}

// NewViewAnalyticsClickHouseRepository tạo instance mới của analytics ClickHouse repository.
//
// Parameters:
//   - ch: ClickHouse client instance
//
// Returns:
//   - domain.ViewAnalyticsRepository: Repository implementation
func NewViewAnalyticsClickHouseRepository(ch *database.ClickHouseClient) domain.ViewAnalyticsRepository {
	return &viewAnalyticsClickHouseRepo{ch: ch}
}

// BatchInsertEvents inserts multiple events vào ClickHouse sử dụng batch API.
// Batch API tối ưu performance bằng cách gom nhiều inserts thành một request.
//
// ClickHouse operation:
//
//	INSERT INTO view_events (event_time, entity_type, entity_id, novel_id, chapter_id, user_id, ip_address, view_count)
//	VALUES (...), (...), (...)
//
// Parameters:
//   - ctx: Context
//   - events: List events cần insert
//
// Returns:
//   - error: Lỗi nếu có
func (r *viewAnalyticsClickHouseRepo) BatchInsertEvents(ctx context.Context, events []*domain.ViewEvent) error {
	if len(events) == 0 {
		return nil
	}

	// Prepare batch insert
	batch, err := r.ch.Conn.PrepareBatch(ctx, "INSERT INTO view_events")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	// Add rows to batch
	for _, event := range events {
		// Map entity type string to enum value
		entityTypeEnum := "chapter"
		if event.EntityType == "novel" {
			entityTypeEnum = "novel"
		}

		err := batch.Append(
			event.EventTime,
			entityTypeEnum,
			event.EntityID,
			event.NovelID,
			event.ChapterID, // Nullable UUID
			event.UserID,    // Nullable UUID
			event.IPAddress,
			event.ViewCount,
		)
		if err != nil {
			return fmt.Errorf("failed to append row to batch: %w", err)
		}
	}

	// Send batch to ClickHouse
	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// GetViewStats retrieves aggregated view statistics cho một entity trong time range.
// Query aggregate data từ view_events table để provide analytics.
//
// ClickHouse query:
//
//	SELECT count() AS total_views, uniqExact(user_id) AS unique_users, ...
//	FROM view_events
//	WHERE entity_type = ? AND entity_id = ? AND event_time >= ? AND event_time < ?
//
// Parameters:
//   - ctx: Context
//   - entityType: "novel" hoặc "chapter"
//   - entityID: UUID của entity
//   - from: Start time của range
//   - to: End time của range
//
// Returns:
//   - map[string]any: Map chứa statistics
//   - error: Lỗi nếu có
func (r *viewAnalyticsClickHouseRepo) GetViewStats(ctx context.Context, entityType string, entityID uuid.UUID, from, to time.Time) (map[string]any, error) {
	query := `
		SELECT
			count() AS total_views,
			uniqExact(user_id) AS unique_users,
			uniqExact(ip_address) AS unique_ips,
			toDate(min(event_time)) AS first_view_date,
			toDate(max(event_time)) AS last_view_date
		FROM view_events
		WHERE entity_type = ? AND entity_id = ? AND event_time >= ? AND event_time < ?
	`

	row := r.ch.Conn.QueryRow(ctx, query, entityType, entityID, from, to)

	var stats struct {
		TotalViews    uint64
		UniqueUsers   uint64
		UniqueIPs     uint64
		FirstViewDate time.Time
		LastViewDate  time.Time
	}

	if err := row.Scan(&stats.TotalViews, &stats.UniqueUsers, &stats.UniqueIPs, &stats.FirstViewDate, &stats.LastViewDate); err != nil {
		return nil, fmt.Errorf("failed to scan stats: %w", err)
	}

	return map[string]any{
		"total_views":     stats.TotalViews,
		"unique_users":    stats.UniqueUsers,
		"unique_ips":      stats.UniqueIPs,
		"first_view_date": stats.FirstViewDate,
		"last_view_date":  stats.LastViewDate,
	}, nil
}
