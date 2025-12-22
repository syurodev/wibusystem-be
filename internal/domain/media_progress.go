/*
Media Progress Domain - Tiến độ đọc/xem của user
=================================================

ARCHITECTURE OVERVIEW:
━━━━━━━━━━━━━━━━━━━━━━
Hệ thống này theo dõi tiến độ đọc/xem của user trên tất cả các loại media.
Sử dụng 2-table design để tối ưu cả việc hiển thị "Continue Reading"
lẫn việc hiển thị trạng thái đã đọc trên chapter list.

┌─────────────────────────────────────────────────────────────────────────────┐
│                           DATA MODEL                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  MediaProgress (1 entry per user per media)                                 │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                                │
│  - Lưu vị trí HIỆN TẠI user đang đọc/xem                                   │
│  - Dùng cho "Continue Reading" section                                      │
│  - Query: ORDER BY last_accessed_at DESC LIMIT N                           │
│                                                                             │
│  UnitProgress (1 entry per user per chapter/episode)                        │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                         │
│  - Lưu trạng thái đã đọc/xem của từng chapter                              │
│  - Dùng cho hiển thị icon "đã đọc" trên chapter list                       │
│  - Query: WHERE media_id = X ORDER BY unit_id                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

FLOW: Khi User Đọc Chapter
━━━━━━━━━━━━━━━━━━━━━━━━━
1. User mở chapter để đọc
   │
   ▼
2. Frontend gọi POST /api/v1/history
   {
     "content_id": "novel-uuid",
     "media_type": "novel",
     "latest_unit_id": "chapter-uuid",
     "novel_last_read_info": { "node_id": "...", "preview": "..." }
   }
   │
   ▼
3. Service.UpdateProgress():
   a. UPSERT unit_progress
      - Tạo mới nếu chưa có (status = in_progress)
      - Update position nếu đã có
   │
   ▼
   b. UPSERT media_progress
      - Update current_unit_id = chapter-uuid
      - Update position = {...}
      - Update last_accessed_at = NOW()
   │
   ▼
4. Khi user đọc xong chapter (scroll tới cuối):
   - Frontend gọi POST /api/v1/progress/.../units/.../complete
   - Update unit_progress.status = 'completed'
   - Update unit_progress.completed_at = NOW()
   - Increment media_progress.completed_units

FLOW: Hiển thị "Continue Reading"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Frontend gọi GET /api/v1/history/recent?limit=12
   │
   ▼
2. Repository.GetRecent():
   SELECT mp.*, n.title, n.slug, n.cover_image_url, c.title, c.chapter_number
   FROM media_progress mp
   JOIN novels n ON mp.media_id = n.id
   JOIN chapters c ON mp.current_unit_id = c.id
   WHERE mp.user_id = $1
   ORDER BY mp.last_accessed_at DESC
   LIMIT $2
   │
   ▼
3. Return list với đầy đủ thông tin để hiển thị

FLOW: Hiển thị Chapter List với trạng thái đã đọc
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Frontend gọi GET /api/v1/progress/novel/{id}/units
   │
   ▼
2. Repository.GetUnitsByMedia():
   SELECT up.unit_id, up.status, up.completed_at
   FROM unit_progress up
   WHERE up.user_id = $1 AND up.media_type = $2 AND up.media_id = $3
   │
   ▼
3. Frontend merge với chapter list để hiển thị icon "đã đọc"
*/

package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// =============================================================================
// ENUMS & CONSTANTS
// =============================================================================

// UnitProgressStatus định nghĩa trạng thái đọc/xem của một unit (chapter/episode)
type UnitProgressStatus string

const (
	// UnitStatusInProgress - User đã bắt đầu đọc/xem nhưng chưa hoàn thành
	UnitStatusInProgress UnitProgressStatus = "in_progress"

	// UnitStatusCompleted - User đã đọc/xem xong
	UnitStatusCompleted UnitProgressStatus = "completed"
)

// NOTE: MediaType constants (MediaTypeNovel, MediaTypeManga, MediaTypeAnime)
// are defined in view_tracking.go and should be used from there.

// IsValid kiểm tra xem status có hợp lệ không
func (s UnitProgressStatus) IsValid() bool {
	switch s {
	case UnitStatusInProgress, UnitStatusCompleted:
		return true
	default:
		return false
	}
}

// =============================================================================
// MEDIA PROGRESS ENTITY
// Lưu trữ tiến độ tổng thể của user cho mỗi media
// =============================================================================

// MediaProgress represents a user's overall progress for a media item.
//
// Mỗi user có tối đa 1 entry cho mỗi media (novel/manga/anime).
// Entry này lưu trữ:
//   - Vị trí hiện tại (chapter/episode nào, vị trí trong chapter)
//   - Thống kê tiến độ (đã đọc bao nhiêu chapter, % hoàn thành)
//   - Thời gian truy cập gần nhất (để sort "Continue Reading")
type MediaProgress struct {
	ID     uuid.UUID
	UserID uuid.UUID

	// Media reference (polymorphic)
	MediaType string    // "novel", "manga", "anime"
	MediaID   uuid.UUID // References novels.id, mangas.id, or animes.id

	// Current position
	// CurrentUnitID là chapter/episode user đang đọc/xem
	CurrentUnitID uuid.UUID

	// Position chi tiết BÊN TRONG unit hiện tại
	// - Novel: { "node_id": "paragraph-uuid", "preview": "First 100 chars..." }
	// - Manga: { "page": 15 }
	// - Anime: { "time": "12:34", "seconds": 754 }
	Position json.RawMessage

	// Progress statistics (denormalized for fast reads)
	TotalUnits         int     // Tổng số chapter/episode của media
	CompletedUnits     int     // Số chapter/episode đã hoàn thành
	ProgressPercentage float64 // = (CompletedUnits / TotalUnits) * 100

	// Timestamps
	LastAccessedAt time.Time // Lần cuối user truy cập (để sort Continue Reading)
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// ==========================================================================
	// RELATIONS (loaded via JOINs, không lưu trong DB)
	// ==========================================================================

	// Media info để hiển thị trong Continue Reading
	Media *ProgressMedia `db:"-"`

	// Current unit info
	CurrentUnit *ProgressUnit `db:"-"`
}

// ProgressMedia chứa thông tin media cần thiết cho hiển thị
// Được load từ JOIN với novels/mangas/animes table
type ProgressMedia struct {
	ID       uuid.UUID
	Title    string
	Slug     string
	CoverURL *string
	Status   string // "ongoing", "completed", etc.

	// Owner info (người đăng nội dung)
	OwnerID          uuid.UUID
	OwnerDisplayName *string
	OwnerUsername    *string
	OwnerAvatarURL   *string

	// Genres (optional, for display)
	Genres []Genre
}

// ProgressUnit chứa thông tin unit (chapter/episode) hiện tại
// Được load từ JOIN với chapters/episodes table
type ProgressUnit struct {
	ID     uuid.UUID
	Number int    // chapter_number hoặc episode_number
	Title  string
	Slug   string
}

// =============================================================================
// UNIT PROGRESS ENTITY
// Lưu trữ trạng thái đọc/xem của từng chapter/episode
// =============================================================================

// UnitProgress represents a user's progress for a specific chapter/episode.
//
// Mỗi user có tối đa 1 entry cho mỗi unit (chapter/episode).
// Entry này cho biết user đã đọc/xem unit này chưa và ở trạng thái nào.
//
// Dùng để:
//   - Hiển thị icon "đã đọc" trên chapter list
//   - Tính toán completed_units trong MediaProgress
//   - Track thời gian đọc của user
type UnitProgress struct {
	ID     uuid.UUID
	UserID uuid.UUID

	// Media reference (để query tất cả units của 1 media)
	MediaType string
	MediaID   uuid.UUID

	// Unit reference
	UnitID uuid.UUID // References chapters.id or episodes.id

	// Read/Watch status
	Status UnitProgressStatus // "in_progress" or "completed"

	// Position within unit (để resume nếu cần)
	Position json.RawMessage

	// Time tracking
	StartedAt      time.Time  // Khi user mở unit lần đầu
	CompletedAt    *time.Time // Khi user đọc/xem xong (NULL nếu in_progress)
	LastAccessedAt time.Time  // Lần cuối truy cập
}

// =============================================================================
// POSITION TYPES
// Các struct định nghĩa format của Position JSON cho từng loại media
// =============================================================================

// NovelPosition lưu vị trí đọc trong chapter của novel
// node_id: ID của paragraph/block trong rich text content
// preview: 100 ký tự đầu của content tại vị trí đó (để hiển thị)
type NovelPosition struct {
	NodeID  string `json:"node_id"`
	Preview string `json:"preview"`
}

// MangaPosition lưu vị trí đọc trong chapter của manga
// page: Số trang hiện tại (1-indexed)
type MangaPosition struct {
	Page int `json:"page"`
}

// AnimePosition lưu vị trí xem trong episode của anime
// time: Thời gian dạng "HH:MM:SS" hoặc "MM:SS"
// seconds: Tổng số giây (để dễ xử lý)
type AnimePosition struct {
	Time    string `json:"time"`    // "12:34" hoặc "01:23:45"
	Seconds int    `json:"seconds"` // 754
}

// =============================================================================
// REPOSITORY INTERFACE
// =============================================================================

// MediaProgressRepository định nghĩa interface cho việc truy cập dữ liệu media progress
type MediaProgressRepository interface {
	// =========================================================================
	// MEDIA PROGRESS OPERATIONS
	// =========================================================================

	// UpsertMediaProgress tạo mới hoặc cập nhật tiến độ media
	//
	// Flow:
	//   1. Check existing entry với (user_id, media_type, media_id)
	//   2. Nếu có: UPDATE current_unit_id, position, last_accessed_at
	//   3. Nếu không: INSERT new entry
	//
	// Được gọi khi: User mở chapter/episode để đọc/xem
	UpsertMediaProgress(ctx context.Context, progress *MediaProgress) error

	// GetRecentMediaProgress lấy N media progress gần nhất
	//
	// Flow:
	//   1. SELECT FROM media_progress WHERE user_id = ?
	//   2. JOIN với novels/chapters để lấy thông tin hiển thị
	//   3. ORDER BY last_accessed_at DESC
	//   4. LIMIT ?
	//
	// Được gọi khi: Hiển thị "Continue Reading" section
	GetRecentMediaProgress(ctx context.Context, userID uuid.UUID, limit int) ([]*MediaProgress, error)

	// GetMediaProgressByUserAndMedia lấy progress của user cho 1 media cụ thể
	//
	// Được gọi khi: User vào trang detail của novel để xem tiến độ
	GetMediaProgressByUserAndMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) (*MediaProgress, error)

	// GetMediaProgressByUserID lấy tất cả progress của user với pagination
	//
	// Được gọi khi: Hiển thị trang History
	GetMediaProgressByUserID(ctx context.Context, userID uuid.UUID, filter MediaProgressFilter) ([]*MediaProgress, int64, error)

	// DeleteMediaProgress xóa progress của 1 media
	//
	// Cascade: Cũng xóa tất cả unit_progress liên quan
	DeleteMediaProgress(ctx context.Context, id uuid.UUID) error

	// DeleteAllMediaProgressByUser xóa tất cả progress của user
	//
	// Được gọi khi: User chọn "Clear All History"
	DeleteAllMediaProgressByUser(ctx context.Context, userID uuid.UUID) error

	// =========================================================================
	// UNIT PROGRESS OPERATIONS
	// =========================================================================

	// UpsertUnitProgress tạo mới hoặc cập nhật tiến độ unit
	//
	// Flow:
	//   1. Check existing entry với (user_id, unit_id)
	//   2. Nếu có: UPDATE status, position, last_accessed_at
	//   3. Nếu không: INSERT với status = in_progress
	//
	// Được gọi khi: User mở chapter/episode
	UpsertUnitProgress(ctx context.Context, progress *UnitProgress) error

	// GetUnitProgressByMedia lấy trạng thái đọc của tất cả units trong 1 media
	//
	// Flow:
	//   1. SELECT FROM unit_progress WHERE user_id = ? AND media_type = ? AND media_id = ?
	//   2. Return map[unit_id]UnitProgress
	//
	// Được gọi khi: Hiển thị chapter list với icon "đã đọc"
	GetUnitProgressByMedia(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) ([]*UnitProgress, error)

	// GetUnitProgress lấy progress của 1 unit cụ thể
	GetUnitProgress(ctx context.Context, userID uuid.UUID, unitID uuid.UUID) (*UnitProgress, error)

	// MarkUnitComplete đánh dấu unit là đã hoàn thành
	//
	// Flow:
	//   1. UPDATE unit_progress SET status = 'completed', completed_at = NOW()
	//   2. UPDATE media_progress SET completed_units = completed_units + 1
	//
	// Được gọi khi: User scroll tới cuối chapter hoặc xem xong episode
	MarkUnitComplete(ctx context.Context, userID uuid.UUID, unitID uuid.UUID) error

	// UpdateCompletedUnitsCount cập nhật số unit đã hoàn thành
	//
	// Recalculate từ unit_progress và update vào media_progress
	UpdateCompletedUnitsCount(ctx context.Context, userID uuid.UUID, mediaType string, mediaID uuid.UUID) error
}

// MediaProgressFilter định nghĩa các filter cho list query
type MediaProgressFilter struct {
	MediaType *string // Filter theo loại media: "novel", "manga", "anime"
	Limit     int     // Số lượng kết quả tối đa
	Offset    int     // Offset cho pagination
	SortBy    string  // "last_accessed_at" (default), "title", "progress"
	SortOrder string  // "asc", "desc" (default)
}
