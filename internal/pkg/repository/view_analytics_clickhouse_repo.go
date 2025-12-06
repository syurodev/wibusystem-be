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
	batch, err := r.ch.Conn.PrepareBatch(ctx, `
		INSERT INTO view_events (
			event_time, media_type, media_id, unit_id, user_id, ip_address, view_count,
			author_ids, genre_ids, artist_ids, tag_ids, group_id, owner_id, studio_id, original_language,
			platform, os, browser, country_code, city, referrer,
			is_premium, user_role
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	// Add rows to batch
	for _, event := range events {
		err := batch.Append(
			event.EventTime,
			event.MediaType,
			event.MediaID,
			event.UnitID,
			event.UserID,
			event.IPAddress,
			event.ViewCount,
			event.AuthorIDs,
			event.GenreIDs,
			event.ArtistIDs,
			event.TagIDs,
			event.GroupID,
			event.OwnerID,
			event.StudioID,
			event.OriginalLanguage,
			event.Platform,
			event.OS,
			event.Browser,
			event.CountryCode,
			event.City,
			event.Referrer,
			event.IsPremium,
			event.UserRole,
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
		WHERE media_type = ? AND media_id = ? AND event_time >= ? AND event_time < ?
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

// GetTopTrending retrieves top trending entities based on view counts.
// Supports filtering by media type and time range.
func (r *viewAnalyticsClickHouseRepo) GetTopTrending(ctx context.Context, mediaType string, timeRange string, limit int) ([]map[string]any, error) {
	// Base query using the materialized view for performance
	query := `
		SELECT
			media_type,
			media_id,
			sum(total_views) AS views
		FROM view_events_daily
		WHERE event_date >= now() - INTERVAL ?
	`
	args := []interface{}{}

	// Parse time range (simple implementation)
	// In production, might want better parsing or pass duration
	interval := "7 DAY" // default
	switch timeRange {
	case "day":
		interval = "1 DAY"
	case "week":
		interval = "7 DAY"
	case "month":
		interval = "30 DAY"
	}
	
	// Note: ClickHouse parameter binding for INTERVAL is tricky, usually needs direct string injection or specific syntax.
	// For safety, we'll use the validated interval string directly in the query since it's from a controlled switch.
	query = fmt.Sprintf(`
		SELECT
			media_type,
			media_id,
			sum(total_views) AS views
		FROM view_events_daily
		WHERE event_date >= now() - INTERVAL %s
	`, interval)

	// Add media type filter if provided
	if mediaType != "" {
		query += " AND media_type = ?"
		args = append(args, mediaType)
	}

	query += " GROUP BY media_type, media_id ORDER BY views DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.ch.Conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending: %w", err)
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var (
			mType string
			mID   uuid.UUID
			views uint64
		)
		if err := rows.Scan(&mType, &mID, &views); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, map[string]any{
			"media_type": mType,
			"media_id":   mID,
			"views":      views,
		})
	}

	return results, nil
}

// GetGenreActiveReaders retrieves the count of active readers per genre for the last N days.
// Active reader = distinct user (or IP if anonymous) who viewed any content in that genre.
func (r *viewAnalyticsClickHouseRepo) GetGenreActiveReaders(ctx context.Context, days int) (map[uuid.UUID]int64, error) {
	// Query to count distinct users/IPs per genre
	// We use array join because genre_ids is an array column
	query := fmt.Sprintf(`
		SELECT
			genre_id,
			uniqExact(if(user_id != '00000000-0000-0000-0000-000000000000', toString(user_id), ip_address)) as active_readers
		FROM view_events
		ARRAY JOIN genre_ids AS genre_id
		WHERE event_time >= now() - INTERVAL %d DAY
		GROUP BY genre_id
	`, days)

	rows, err := r.ch.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active readers: %w", err)
	}
	defer rows.Close()

	results := make(map[uuid.UUID]int64)
	for rows.Next() {
		var (
			genreID uuid.UUID
			count   uint64
		)
		if err := rows.Scan(&genreID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results[genreID] = int64(count)
	}

	return results, nil
}
