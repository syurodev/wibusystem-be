package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// UserStatistics chứa các chỉ số thống kê của user
type UserStatistics struct {
	UserID uuid.UUID `json:"user_id"`

	// Social stats
	FollowerCount  int `json:"follower_count"`
	FollowingCount int `json:"following_count"`

	// Content counts (by type)
	NovelCount int `json:"novel_count"`
	MangaCount int `json:"manga_count"`
	AnimeCount int `json:"anime_count"`

	// Chapter/Episode counts
	NovelChapterCount int `json:"novel_chapter_count"`
	MangaChapterCount int `json:"manga_chapter_count"`
	AnimeEpisodeCount int `json:"anime_episode_count"`

	// Engagement
	TotalViews int64 `json:"total_views"`

	// Timestamps
	LastContentUpdatedAt *time.Time `json:"last_content_updated_at"`
	LastActivityAt       *time.Time `json:"last_activity_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// WorksCount returns computed total works (novel + manga + anime)
func (s *UserStatistics) WorksCount() int {
	return s.NovelCount + s.MangaCount + s.AnimeCount
}

// UserStatisticsRepository định nghĩa interface cho việc truy cập dữ liệu user statistics
type UserStatisticsRepository interface {
	// GetByUserID lấy statistics theo user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserStatistics, error)

	// Create tạo statistics mới cho user
	Create(ctx context.Context, stats *UserStatistics) error

	// Update cập nhật statistics
	Update(ctx context.Context, stats *UserStatistics) error

	// IncrementNovelCount tăng số novel đã đăng
	IncrementNovelCount(ctx context.Context, userID uuid.UUID) error

	// DecrementNovelCount giảm số novel đã đăng
	DecrementNovelCount(ctx context.Context, userID uuid.UUID) error

	// IncrementMangaCount tăng số manga đã đăng
	IncrementMangaCount(ctx context.Context, userID uuid.UUID) error

	// DecrementMangaCount giảm số manga đã đăng
	DecrementMangaCount(ctx context.Context, userID uuid.UUID) error

	// IncrementAnimeCount tăng số anime đã đăng
	IncrementAnimeCount(ctx context.Context, userID uuid.UUID) error

	// DecrementAnimeCount giảm số anime đã đăng
	DecrementAnimeCount(ctx context.Context, userID uuid.UUID) error

	// IncrementNovelChapterCount tăng số chapter novel
	IncrementNovelChapterCount(ctx context.Context, userID uuid.UUID) error

	// IncrementMangaChapterCount tăng số chapter manga
	IncrementMangaChapterCount(ctx context.Context, userID uuid.UUID) error

	// IncrementAnimeEpisodeCount tăng số episode anime
	IncrementAnimeEpisodeCount(ctx context.Context, userID uuid.UUID) error

	// UpdateLastContentUpdatedAt cập nhật timestamp nội dung mới nhất
	UpdateLastContentUpdatedAt(ctx context.Context, userID uuid.UUID) error

	// EnsureExists đảm bảo record tồn tại cho user (upsert)
	EnsureExists(ctx context.Context, userID uuid.UUID) error
}
