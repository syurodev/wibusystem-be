/*
View Tracking System - Architecture Flow
=========================================

Hệ thống view tracking được thiết kế theo kiến trúc event-driven với 3 layers chính:

┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API LAYER                                          │
│  User Request → TrackChapterView() → Dedup Check → Buffer Increment → Enqueue  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           REDIS LAYER (Buffer)                                  │
│                                                                                 │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐              │
│  │ Dedup Keys       │  │ View Buffers     │  │ Event Queue      │              │
│  │ (TTL: 1 hour)    │  │ (Hash: count)    │  │ (List: JSON)     │              │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘              │
└─────────────────────────────────────────────────────────────────────────────────┘
                                         │
                          ViewSyncWorker (Background, định kỳ)
                                         │
                        ┌────────────────┴────────────────┐
                        ▼                                 ▼
┌────────────────────────────────┐     ┌────────────────────────────────────────┐
│       POSTGRESQL               │     │           CLICKHOUSE                   │
│  (view_count updates)          │     │  (Analytics events)                    │
│                                │     │                                        │
│  novels.view_count += N        │     │  view_events table                     │
│  chapters.view_count += N      │     │  → trending queries                    │
│                                │     │  → genre stats                         │
│                                │     │  → creator stats                       │
└────────────────────────────────┘     └────────────────────────────────────────┘

FLOW CHI TIẾT:
==============

1. TRACKING FLOW (Real-time):
   a. API nhận request view chapter
   b. CheckDeduplication() - kiểm tra user/IP đã view gần đây chưa (Redis EXISTS)
   c. Nếu duplicate → skip, return false
   d. Nếu unique:
      - RecordDeduplication() - đánh dấu đã view (Redis SETNX với TTL)
      - IncrementBuffer() - tăng counter trong Redis (HINCRBY)
      - EnqueueEvent() - đẩy ViewEvent vào queue (RPUSH)
   e. Return true

2. SYNC FLOW (Background Worker - mỗi N phút):
   a. GetAllBuffers() - lấy tất cả buffered counts từ Redis (atomic read + clear)
   b. Batch update PostgreSQL (novels.view_count, chapters.view_count)
   c. DequeueEvents() - lấy batch events từ queue (LPOP)
   d. BatchInsertEvents() - insert vào ClickHouse

3. ANALYTICS FLOW (On-demand queries):
   - GetTopTrending() - top trending content
   - GetGenreActiveReaders() - active readers per genre
   - GetCreatorViewStats() - creator statistics
   - GetTopGenresByViews() - top genres by view count
   - GetTopCreatorsByViews() / GetTopOrgsByViews() - leaderboards

DESIGN DECISIONS:
=================
- Redis buffer: Tránh write amplification trực tiếp vào PostgreSQL
- Deduplication: Tránh spam views từ cùng user/IP
- Event queue: Decouple tracking với analytics storage
- ClickHouse: Optimized cho time-series analytics
- Dual storage: PostgreSQL cho display count, ClickHouse cho analytics
*/

package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// MediaType constants for analytics
const (
	MediaTypeNovel = "novel"
	MediaTypeManga = "manga"
	MediaTypeAnime = "anime"
)

// OwnerType constants for distinguishing content ownership
const (
	OwnerTypeUser = "user" // Content owned by a creator (individual user)
	OwnerTypeOrg  = "org"  // Content owned by an organization
)

// ============================================================================
// Analytics Enums
// ============================================================================

// AnalyticsPeriod định nghĩa các khoảng thời gian cho analytics queries
type AnalyticsPeriod string

const (
	PeriodDay   AnalyticsPeriod = "day"
	PeriodWeek  AnalyticsPeriod = "week"
	PeriodMonth AnalyticsPeriod = "month"
	PeriodYear  AnalyticsPeriod = "year"
)

// IsValid kiểm tra xem period có hợp lệ không
func (p AnalyticsPeriod) IsValid() bool {
	switch p {
	case PeriodDay, PeriodWeek, PeriodMonth, PeriodYear:
		return true
	default:
		return false
	}
}

// ToClickHouseInterval chuyển đổi period sang ClickHouse INTERVAL format
func (p AnalyticsPeriod) ToClickHouseInterval() string {
	switch p {
	case PeriodDay:
		return "1 DAY"
	case PeriodWeek:
		return "7 DAY"
	case PeriodMonth:
		return "30 DAY"
	case PeriodYear:
		return "365 DAY"
	default:
		return "7 DAY"
	}
}

// AnalyticsEntityType định nghĩa các loại entity cho analytics ranking
type AnalyticsEntityType string

const (
	EntityTypeGenre   AnalyticsEntityType = "genre"
	EntityTypeCreator AnalyticsEntityType = "creator"
	EntityTypeOrg     AnalyticsEntityType = "org"
	EntityTypeNovel   AnalyticsEntityType = "novel"
	EntityTypeManga   AnalyticsEntityType = "manga"
	EntityTypeAnime   AnalyticsEntityType = "anime"
)

// IsValid kiểm tra xem entity type có hợp lệ không
func (e AnalyticsEntityType) IsValid() bool {
	switch e {
	case EntityTypeGenre, EntityTypeCreator, EntityTypeOrg, EntityTypeNovel, EntityTypeManga, EntityTypeAnime:
		return true
	default:
		return false
	}
}

// ViewEvent represents a single view tracking event for analytics.
// This struct is used to store view events in ClickHouse for detailed analytics.
type ViewEvent struct {
	EventTime time.Time
	MediaType string // "novel", "manga", "anime"
	MediaID   uuid.UUID
	UnitID    uuid.UUID // ChapterID or EpisodeID
	UserID    *uuid.UUID
	IPAddress string
	ViewCount uint32

	// Content Metadata
	AuthorIDs        []uuid.UUID
	GenreIDs         []uuid.UUID
	ArtistIDs        []uuid.UUID
	TagIDs           []uuid.UUID
	GroupID          *uuid.UUID
	OwnerID          *uuid.UUID
	OwnerType        string // "user" or "org"
	StudioID         *uuid.UUID
	OriginalLanguage string

	// User Context
	Platform    string
	OS          string
	Browser     string
	CountryCode string
	City        string
	Referrer    string

	// User Status
	IsPremium bool
	UserRole  string
}

// ViewBuffer represents buffered view counts awaiting sync to PostgreSQL.
// This struct is used to store aggregated counts from Redis before bulk updating PostgreSQL.
type ViewBuffer struct {
	EntityType string // "novel" or "chapter"
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
	// DEPRECATED: Use PeekBuffers + ClearBuffers for safer processing.
	//
	// Parameters:
	//   - ctx: Context
	//
	// Returns:
	//   - []ViewBuffer: List các buffered counts
	//   - error: Lỗi nếu có
	GetAllBuffers(ctx context.Context) ([]ViewBuffer, error)

	// PeekBuffers retrieves tất cả buffered counts WITHOUT clearing them.
	// Use this followed by ClearBuffers after successful processing.
	//
	// Parameters:
	//   - ctx: Context
	//
	// Returns:
	//   - []ViewBuffer: List các buffered counts
	//   - error: Lỗi nếu có
	PeekBuffers(ctx context.Context) ([]ViewBuffer, error)

	// ClearBuffers removes all buffered counts from Redis.
	// Call this ONLY after successful processing to acknowledge.
	//
	// Parameters:
	//   - ctx: Context
	//
	// Returns:
	//   - error: Lỗi nếu có
	ClearBuffers(ctx context.Context) error

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
	// DEPRECATED: Use PeekEvents + AcknowledgeEvents for safer processing.
	//
	// Parameters:
	//   - ctx: Context
	//   - batchSize: Số lượng events tối đa cần dequeue
	//
	// Returns:
	//   - []*ViewEvent: List events, nil nếu queue empty
	//   - error: Lỗi nếu có
	DequeueEvents(ctx context.Context, batchSize int) ([]*ViewEvent, error)

	// PeekEvents reads events from queue WITHOUT removing them.
	// Use this followed by AcknowledgeEvents after successful processing.
	//
	// Parameters:
	//   - ctx: Context
	//   - batchSize: Số lượng events tối đa cần peek
	//
	// Returns:
	//   - []*ViewEvent: List events, nil nếu queue empty
	//   - error: Lỗi nếu có
	PeekEvents(ctx context.Context, batchSize int) ([]*ViewEvent, error)

	// AcknowledgeEvents removes N events from the head of queue after successful processing.
	//
	// Parameters:
	//   - ctx: Context
	//   - count: Number of events to remove (should match PeekEvents count)
	//
	// Returns:
	//   - error: Lỗi nếu có
	AcknowledgeEvents(ctx context.Context, count int) error

	// EnqueueActivity adds a content activity to the queue.
	EnqueueActivity(ctx context.Context, activity *ContentActivity) error

	// DequeueActivities retrieves and removes a batch of content activities from the queue.
	// DEPRECATED: Use PeekActivities + AcknowledgeActivities for safer processing.
	DequeueActivities(ctx context.Context, batchSize int) ([]*ContentActivity, error)

	// PeekActivities reads activities from queue WITHOUT removing them.
	PeekActivities(ctx context.Context, batchSize int) ([]*ContentActivity, error)

	// AcknowledgeActivities removes N activities from the head of queue after successful processing.
	AcknowledgeActivities(ctx context.Context, count int) error
}

// RankStat đại diện cho ranking với so sánh với kỳ trước
type RankStat struct {
	EntityID     uuid.UUID
	EntityType   string // "genre", "creator", "org", "novel", "manga", "anime"
	TotalViews   uint64
	UniqueUsers  uint64
	CurrentRank  int
	PreviousRank *int // nil = mục mới
	RankChange   *int // dương = tăng, âm = giảm, 0 = không đổi, nil = mục mới
}

// ViewAnalyticsRepository định nghĩa interface cho ClickHouse analytics operations.
// Repository này xử lý batch insert events và analytics queries.
type ViewAnalyticsRepository interface {
	// BatchInsertEvents inserts nhiều view events vào ClickHouse.
	// Sử dụng ClickHouse batch API để tối ưu hiệu năng.
	//
	// Parameters:
	//   - ctx: Context
	//   - events: Danh sách events cần insert
	//
	// Returns:
	//   - error: Lỗi nếu có
	BatchInsertEvents(ctx context.Context, events []*ViewEvent) error

	// CreateRankSnapshot tạo snapshot xếp hạng cho một loại entity và kỳ cụ thể.
	// Method này tổng hợp view counts và lưu vào bảng rank_snapshots.
	//
	// Parameters:
	//   - ctx: Context
	//   - snapshotDate: Ngày snapshot
	//   - period: "week", "month", "year"
	//   - entityType: "genre", "creator", "org", "novel", "manga", "anime"
	//   - limit: Số lượng top items cần lưu
	//
	// Returns:
	//   - error: Lỗi nếu có
	CreateRankSnapshot(ctx context.Context, snapshotDate time.Time, period string, entityType string, limit int) error

	// GetRankWithComparison lấy danh sách top entities kèm so sánh thứ hạng với kỳ trước.
	//
	// Parameters:
	//   - ctx: Context
	//   - period: "week", "month", "year"
	//   - entityType: "genre", "creator", "org", "novel", "manga", "anime"
	//   - limit: Số lượng items cần trả về
	//
	// Returns:
	//   - []RankStat: Danh sách xếp hạng kèm so sánh
	//   - error: Lỗi nếu có
	// GetRankWithComparison lấy danh sách top entities kèm so sánh thứ hạng với kỳ trước.
	// offset: 0 = current period, 1 = previous period
	GetRankWithComparison(ctx context.Context, period string, entityType string, offset int, limit int) ([]RankStat, error)

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

	// GetTopTrending retrieves top trending entities based on view counts.
	//
	// Parameters:
	//   - ctx: Context
	//   - mediaType: Optional filter by media type (novel, manga, anime). Empty string for all.
	//   - timeRange: Time range for trending (e.g. "1 day", "7 days", "30 days")
	//   - limit: Number of items to return
	//
	// Returns:
	//   - []map[string]any: List of trending items
	//   - error: Error if any
	// GetTopTrending retrieves top trending entities based on view counts.
	//
	// Parameters:
	//   - ctx: Context
	//   - mediaType: Optional filter by media type (novel, manga, anime). Empty string for all.
	//   - timeRange: Time range for trending (e.g. "1 day", "7 days", "30 days")
	//   - limit: Number of items to return
	//
	// Returns:
	//   - []map[string]any: List of trending items
	//   - error: Error if any
	GetTopTrending(ctx context.Context, mediaType string, timeRange string, limit int) ([]map[string]any, error)

	// GetGenreActiveReaders retrieves the count of active readers per genre for the last N days.
	//
	// Parameters:
	//   - ctx: Context
	//   - days: Number of days to look back
	//
	// Returns:
	//   - map[uuid.UUID]int64: Map of genre ID to active readers count
	//   - error: Error if any
	GetGenreActiveReaders(ctx context.Context, days int) (map[uuid.UUID]int64, error)

	// GetCreatorViewStats retrieves view statistics for multiple creators (by owner_id).
	//
	// Parameters:
	//   - ctx: Context
	//   - ownerIDs: List of owner/creator user IDs
	//   - timeRange: Time range (day, week, month, year, all)
	//
	// Returns:
	//   - map[uuid.UUID]CreatorViewStats: Map of owner ID to their view stats
	//   - error: Error if any
	GetCreatorViewStats(ctx context.Context, ownerIDs []uuid.UUID, timeRange string) (map[uuid.UUID]CreatorViewStats, error)

	// BatchInsertActivities inserts multiple content activities into ClickHouse.
	BatchInsertActivities(ctx context.Context, activities []*ContentActivity) error

	// GetTopActiveCreators retrieves top active creators based on content creation.
	GetTopActiveCreators(ctx context.Context, timeRange string, limit int) ([]CreatorActivityStat, error)

	// GetTopActiveOrgs retrieves top active organizations based on content creation.
	GetTopActiveOrgs(ctx context.Context, timeRange string, limit int) ([]OrgActivityStat, error)

	// GetTopGenresByViews retrieves genres with most views in a calendar-based time range.
	// Uses calendar periods (actual week starting Monday, actual month starting 1st, etc.)
	//
	// Parameters:
	//   - ctx: Context
	//   - period: Time period (day, week, month, year)
	//   - offset: 0 for current period, 1 for previous period
	//   - limit: Number of genres to return
	//
	// Returns:
	//   - []GenreViewStat: List of genre view stats ordered by views desc
	//   - error: Error if any
	GetTopGenresByViews(ctx context.Context, period string, offset int, limit int) ([]GenreViewStat, error)

	// GetTopCreatorsByViews retrieves creators with most views in a calendar-based time range.
	GetTopCreatorsByViews(ctx context.Context, period string, offset int, limit int) ([]CreatorViewStat, error)

	// GetTopOrgsByViews retrieves organizations with most views in a calendar-based time range.
	GetTopOrgsByViews(ctx context.Context, period string, offset int, limit int) ([]OrgViewStat, error)
}

// Action types for content activities
const (
	ActionTypeCreate  = "create"
	ActionTypePublish = "publish"
	ActionTypeDelete  = "delete"
)

// Target types for content activities
const (
	TargetTypeMedia   = "media"
	TargetTypeVolume  = "volume"
	TargetTypeChapter = "chapter"
	TargetTypeEpisode = "episode"
)

// ContentActivity represents a content creation/publishing activity for analytics.
type ContentActivity struct {
	EventTime  time.Time
	ActionType string // "create", "publish", "delete"
	MediaType  string // "novel", "manga", "anime"
	MediaID    uuid.UUID
	TargetType string // "media", "chapter", "volume", "episode"
	TargetID   uuid.UUID
	UserID     uuid.UUID
	OrgID      *uuid.UUID
	Weight     int64 // e.g., word count
}

// CreatorActivityStat represents aggregated activity stats for a creator
type CreatorActivityStat struct {
	UserID        uuid.UUID
	TotalActivity int64
	TotalWeight   int64
	User          *User // Optional, filled by service
}

// OrgActivityStat represents aggregated activity stats for an organization
type OrgActivityStat struct {
	OrgID         uuid.UUID
	TotalActivity int64
	Organization  *Organization // Optional, filled by service
}

// CreatorViewStats represents aggregated view stats for a creator
type CreatorViewStats struct {
	TotalViews      int64
	PopularWorkID   *uuid.UUID
	PopularWorkType string // novel, manga, anime
}

// GenreViewStat represents aggregated view stats for a genre
type GenreViewStat struct {
	GenreID     uuid.UUID
	TotalViews  uint64
	UniqueUsers uint64
}

// CreatorViewStat represents aggregated view stats for a creator (owner_type = 'user')
type CreatorViewStat struct {
	CreatorID   uuid.UUID
	TotalViews  uint64
	UniqueUsers uint64
}

// OrgViewStat represents aggregated view stats for an organization (owner_type = 'org')
type OrgViewStat struct {
	OrgID       uuid.UUID
	TotalViews  uint64
	UniqueUsers uint64
}
