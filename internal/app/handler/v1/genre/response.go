package genre

import "github.com/gofrs/uuid/v5"

// GenreResponse là DTO cho response genre
type GenreResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	IsActive      bool    `json:"is_active"`
	SeriesCount   int     `json:"series_count"`   // tổng số novel (sau này sẽ có cả manga và anime)
	ActiveReaders int64   `json:"active_readers"` // số độc giả đang active
	TotalViews    int64   `json:"total_views"`    // tổng lượt xem
	Trend         string  `json:"trend"`          // "rising" | "stable" | "falling"
	Description   *string `json:"description,omitempty"`
	CreatedAt     string  `json:"created_at"` // Ngày tạo
	UpdatedAt     string  `json:"updated_at"` // Ngày cập nhật
}

// GenreDetailResponse là DTO chi tiết cho genre (bao gồm thêm thông tin)
type GenreDetailResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
	IsActive    bool    `json:"is_active"`

	// Statistics
	SeriesCount   int    `json:"series_count"`
	ActiveReaders int64  `json:"active_readers"`
	TotalViews    int64  `json:"total_views"`
	Trend         string `json:"trend"`

	// Audit
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// PreviewMergeGenreResponse là response preview trước khi merge
type PreviewMergeGenreResponse struct {
	AffectedNovels []AffectedNovel `json:"affected_novels"`
	SourceGenres   []string        `json:"source_genres"`
}

// AffectedNovel là thông tin novel bị ảnh hưởng khi merge
type AffectedNovel struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	CoverImageURL *string   `json:"cover_image_url"`
}
