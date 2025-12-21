package genre

import (
	novelres "system/internal/dto/novel"
)

// Alias for AffectedNovel from novel dto
type AffectedNovel = novelres.AffectedNovel

// GenreResponse là DTO cho response genre
type GenreResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	IsActive      bool    `json:"is_active"`
	SeriesCount   int     `json:"series_count"`
	ActiveReaders int64   `json:"active_readers"`
	TotalViews    int64   `json:"total_views"`
	Trend         string  `json:"trend"`
	Description   *string `json:"description,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`

	// Rank Comparison (Optional)
	CurrentRank  *int `json:"current_rank,omitempty"`
	PreviousRank *int `json:"previous_rank,omitempty"`
	RankChange   *int `json:"rank_change,omitempty"`
}

// GenreDetailResponse là DTO chi tiết cho genre
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

// SelectionResponse là DTO cho dropdown selection (ID + Name)
// Sử dụng cho các select fields khi cần chọn genre
type SelectionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
