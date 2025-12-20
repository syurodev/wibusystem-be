package analytics

import (
	"context"
	"fmt"
	"strings"
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

	// Add media type filter if provided (skip "all" as it means no filter)
	if mediaType != "" && mediaType != "all" {
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

// GetCreatorViewStats retrieves view statistics for multiple creators (by owner_id).
func (r *viewAnalyticsClickHouseRepo) GetCreatorViewStats(ctx context.Context, ownerIDs []uuid.UUID, timeRange string) (map[uuid.UUID]domain.CreatorViewStats, error) {
	if len(ownerIDs) == 0 {
		return make(map[uuid.UUID]domain.CreatorViewStats), nil
	}

	// Determine date filter based on timeRange
	var dateFilter string
	switch timeRange {
	case "day":
		dateFilter = "event_time >= now() - INTERVAL 1 DAY"
	case "week":
		dateFilter = "event_time >= now() - INTERVAL 7 DAY"
	case "month":
		dateFilter = "event_time >= now() - INTERVAL 30 DAY"
	case "year":
		dateFilter = "event_time >= now() - INTERVAL 365 DAY"
	default: // "all" or empty
		dateFilter = "1=1"
	}

	// Build owner IDs placeholder
	ownerIDStrs := make([]string, len(ownerIDs))
	for i, id := range ownerIDs {
		ownerIDStrs[i] = fmt.Sprintf("'%s'", id.String())
	}
	ownerIDList := strings.Join(ownerIDStrs, ",")

	// Query for total views per owner
	totalViewsQuery := fmt.Sprintf(`
		SELECT 
			owner_id,
			sum(view_count) as total_views
		FROM view_events
		WHERE owner_id IN (%s) AND %s
		GROUP BY owner_id
	`, ownerIDList, dateFilter)

	rows, err := r.ch.Conn.Query(ctx, totalViewsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query total views: %w", err)
	}

	results := make(map[uuid.UUID]domain.CreatorViewStats)
	for rows.Next() {
		var ownerID uuid.UUID
		var totalViews uint64
		if err := rows.Scan(&ownerID, &totalViews); err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan total views: %w", err)
		}
		results[ownerID] = domain.CreatorViewStats{
			TotalViews: int64(totalViews),
		}
	}
	rows.Close()

	// Query for popular work per owner (work with most views in the period)
	popularWorkQuery := fmt.Sprintf(`
		SELECT 
			owner_id,
			media_id,
			media_type,
			sum(view_count) as views
		FROM view_events
		WHERE owner_id IN (%s) AND %s
		GROUP BY owner_id, media_id, media_type
		ORDER BY owner_id, views DESC
	`, ownerIDList, dateFilter)

	rows2, err := r.ch.Conn.Query(ctx, popularWorkQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query popular works: %w", err)
	}
	defer rows2.Close()

	// Track which owners we've already set popular work for (take first = highest views)
	seenOwners := make(map[uuid.UUID]bool)
	for rows2.Next() {
		var ownerID, mediaID uuid.UUID
		var mediaType string
		var views uint64
		if err := rows2.Scan(&ownerID, &mediaID, &mediaType, &views); err != nil {
			return nil, fmt.Errorf("failed to scan popular work: %w", err)
		}

		if !seenOwners[ownerID] {
			seenOwners[ownerID] = true
			stats := results[ownerID]
			stats.PopularWorkID = &mediaID
			stats.PopularWorkType = mediaType
			results[ownerID] = stats
		}
	}

	return results, nil
}

// BatchInsertActivities inserts multiple content activities into ClickHouse.
func (r *viewAnalyticsClickHouseRepo) BatchInsertActivities(ctx context.Context, activities []*domain.ContentActivity) error {
	if len(activities) == 0 {
		return nil
	}

	batch, err := r.ch.Conn.PrepareBatch(ctx, `
		INSERT INTO content_activities (
			event_time, action_type, media_type, media_id, target_type, target_id, user_id, org_id, weight
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare activity batch: %w", err)
	}

	for _, activity := range activities {
		err := batch.Append(
			activity.EventTime,
			activity.ActionType,
			activity.MediaType,
			activity.MediaID,
			activity.TargetType,
			activity.TargetID,
			activity.UserID,
			activity.OrgID,
			activity.Weight,
		)
		if err != nil {
			return fmt.Errorf("failed to append activity to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send activity batch: %w", err)
	}

	return nil
}

// GetTopActiveCreators retrieves top active creators based on content creation.
func (r *viewAnalyticsClickHouseRepo) GetTopActiveCreators(ctx context.Context, timeRange string, limit int) ([]domain.CreatorActivityStat, error) {
	// Determines date filter based on timeRange
	interval := "30 DAY" // default
	switch timeRange {
	case "day":
		interval = "1 DAY"
	case "week":
		interval = "7 DAY"
	case "month":
		interval = "30 DAY"
	case "year":
		interval = "365 DAY"
	}

	// Calculate weighted score: count * 1 + weight (e.g. word count / 1000) ??
	// For now, let's just count activities where action_type = 'publish'
	// Simplest metric: Total Activities.
	// Users want "Most Active" usually means "Most Updates".

	query := fmt.Sprintf(`
		SELECT
			user_id,
			count() as total_activity,
			sum(weight) as total_weight
		FROM content_activities
		WHERE action_type = 'publish' AND event_time >= now() - INTERVAL %s
		GROUP BY user_id
		ORDER BY total_activity DESC
		LIMIT %d
	`, interval, limit)

	rows, err := r.ch.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query top active creators: %w", err)
	}
	defer rows.Close()

	var stats []domain.CreatorActivityStat
	for rows.Next() {
		var s domain.CreatorActivityStat
		if err := rows.Scan(&s.UserID, &s.TotalActivity, &s.TotalWeight); err != nil {
			return nil, fmt.Errorf("failed to scan creator stat: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetTopActiveOrgs retrieves top active organizations based on content creation.
func (r *viewAnalyticsClickHouseRepo) GetTopActiveOrgs(ctx context.Context, timeRange string, limit int) ([]domain.OrgActivityStat, error) {
	// Determines date filter
	interval := "30 DAY"
	switch timeRange {
	case "day":
		interval = "1 DAY"
	case "week":
		interval = "7 DAY"
	case "month":
		interval = "30 DAY"
	case "year":
		interval = "365 DAY"
	}

	query := fmt.Sprintf(`
		SELECT
			org_id,
			count() as total_activity
		FROM content_activities
		WHERE action_type = 'publish' AND org_id IS NOT NULL AND event_time >= now() - INTERVAL %s
		GROUP BY org_id
		ORDER BY total_activity DESC
		LIMIT %d
	`, interval, limit)

	rows, err := r.ch.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query top active orgs: %w", err)
	}
	defer rows.Close()

	var stats []domain.OrgActivityStat
	for rows.Next() {
		var s domain.OrgActivityStat
		// Note: org_id in DB is Nullable(UUID), but we filtered IS NOT NULL so it returns UUID.
		// ClickHouse driver handles this scan if destination is UUID.
		if err := rows.Scan(&s.OrgID, &s.TotalActivity); err != nil {
			return nil, fmt.Errorf("failed to scan org stat: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetTopGenresByViews retrieves genres with most views in a calendar-based time range.
// Uses calendar periods (week starting Monday, month starting 1st).
//
// Parameters:
//   - period: "day", "week", "month", "year"
//   - offset: 0 = current period, 1 = previous period
func (r *viewAnalyticsClickHouseRepo) GetTopGenresByViews(ctx context.Context, period string, offset int, limit int) ([]domain.GenreViewStat, error) {
	now := time.Now()

	var startDate, endDate time.Time

	switch period {
	case "day":
		// Current day or previous day
		targetDay := now.AddDate(0, 0, -offset)
		startDate = time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), 0, 0, 0, 0, targetDay.Location())
		endDate = startDate.AddDate(0, 0, 1)

	case "week":
		// Week starts on Monday
		// Find Monday of current week
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		mondayOfThisWeek := now.AddDate(0, 0, -(weekday - 1))

		if offset == 0 {
			// Current week: Monday to now
			startDate = time.Date(mondayOfThisWeek.Year(), mondayOfThisWeek.Month(), mondayOfThisWeek.Day(), 0, 0, 0, 0, now.Location())
			endDate = now.AddDate(0, 0, 1) // Include today
		} else {
			// Previous week: Monday to Sunday of week (offset) weeks ago
			targetMonday := mondayOfThisWeek.AddDate(0, 0, -7*offset)
			startDate = time.Date(targetMonday.Year(), targetMonday.Month(), targetMonday.Day(), 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(0, 0, 7) // Full week
		}

	case "month":
		if offset == 0 {
			// Current month: 1st of month to now
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = now.AddDate(0, 0, 1) // Include today
		} else {
			// Previous month(s)
			targetMonth := now.AddDate(0, -offset, 0)
			startDate = time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, now.Location())
			// End of that month
			endDate = startDate.AddDate(0, 1, 0)
		}

	case "year":
		if offset == 0 {
			// Current year: Jan 1 to now
			startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			endDate = now.AddDate(0, 0, 1) // Include today
		} else {
			// Previous year(s)
			targetYear := now.Year() - offset
			startDate = time.Date(targetYear, 1, 1, 0, 0, 0, 0, now.Location())
			endDate = time.Date(targetYear+1, 1, 1, 0, 0, 0, 0, now.Location())
		}

	default:
		// Default to current week
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		mondayOfThisWeek := now.AddDate(0, 0, -(weekday - 1))
		startDate = time.Date(mondayOfThisWeek.Year(), mondayOfThisWeek.Month(), mondayOfThisWeek.Day(), 0, 0, 0, 0, now.Location())
		endDate = now.AddDate(0, 0, 1)
	}

	// Query from the raw view_events table using ARRAY JOIN
	// (Using raw table instead of materialized view for compatibility)
	query := `
		SELECT
			genre_id,
			sum(view_count) AS total_views,
			uniqExact(if(user_id IS NOT NULL, toString(user_id), ip_address)) AS unique_users
		FROM view_events
		ARRAY JOIN genre_ids AS genre_id
		WHERE toDate(event_time) >= ? AND toDate(event_time) < ?
		GROUP BY genre_id
		ORDER BY total_views DESC
		LIMIT ?
	`

	rows, err := r.ch.Conn.Query(ctx, query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top genres by views: %w", err)
	}
	defer rows.Close()

	var stats []domain.GenreViewStat
	for rows.Next() {
		var s domain.GenreViewStat
		if err := rows.Scan(&s.GenreID, &s.TotalViews, &s.UniqueUsers); err != nil {
			return nil, fmt.Errorf("failed to scan genre view stat: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

