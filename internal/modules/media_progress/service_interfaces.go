/*
Media Progress Service Interfaces
=================================

Định nghĩa các interface cho service layer.
Tách interfaces ra file riêng để:
- Handler có thể depend on interface thay vì concrete implementation
- Dễ dàng mock trong unit tests
- Clean architecture boundaries
*/

package media_progress

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// Service định nghĩa interface cho media progress business logic
type Service interface {
	// UpdateProgress cập nhật tiến độ đọc/xem của user
	//
	// FLOW:
	// 1. Validate input
	// 2. Get current total_units from novel (để tính %)
	// 3. Upsert unit_progress (track chapter này)
	// 4. Upsert media_progress (update current position)
	//
	// Input: UpdateProgressInput
	// Output: MediaProgressResponse
	UpdateProgress(ctx context.Context, input UpdateProgressInput) (*domain.MediaProgress, error)

	// GetRecentProgress lấy N mục gần nhất cho "Continue" section
	//
	// FLOW:
	// 1. Validate user is authenticated
	// 2. Query media_progress ORDER BY last_accessed_at DESC LIMIT N
	// 3. Return list với đầy đủ info để hiển thị
	GetRecentProgress(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.MediaProgress, error)

	// GetProgressByMedia lấy progress của user cho 1 media cụ thể
	GetProgressByMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) (*domain.MediaProgress, error)

	// GetProgressList lấy danh sách progress với pagination
	GetProgressList(ctx context.Context, userID uuid.UUID, filter domain.MediaProgressFilter) ([]*domain.MediaProgress, int64, error)

	// GetUnitProgress lấy trạng thái đọc của tất cả chapters trong 1 media
	//
	// FLOW:
	// 1. Query unit_progress WHERE media_id = ?
	// 2. Return map[unit_id] → {status, completed_at}
	//
	// Use case: Chapter list hiển thị icon "đã đọc"
	GetUnitProgress(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) ([]*domain.UnitProgress, error)

	// MarkUnitComplete đánh dấu chapter/episode là đã hoàn thành
	//
	// FLOW:
	// 1. Update unit_progress.status = 'completed'
	// 2. Increment media_progress.completed_units
	// 3. Recalculate progress_percentage
	MarkUnitComplete(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID, unitID uuid.UUID) error

	// DeleteProgress xóa progress của 1 media
	DeleteProgress(ctx context.Context, id uuid.UUID) error

	// ClearAllProgress xóa toàn bộ progress của user
	ClearAllProgress(ctx context.Context, userID uuid.UUID) error
}

// UpdateProgressInput chứa dữ liệu để update progress
type UpdateProgressInput struct {
	UserID    uuid.UUID
	MediaType string    // "novel", "manga", "anime"
	MediaID   uuid.UUID // novel_id, manga_id, or anime_id
	UnitID    uuid.UUID // chapter_id or episode_id

	// Position (optional, depends on media type)
	// Novel: NovelPosition{node_id, preview}
	// Manga: MangaPosition{page}
	// Anime: AnimePosition{time, seconds}
	Position any
}
