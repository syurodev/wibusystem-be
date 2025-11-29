package novel

// CreateNovelRequest là DTO cho việc tạo novel mới
type CreateNovelRequest struct {
	Title            string  `json:"title" binding:"required,min=1,max=500"`
	Synopsis         string  `json:"synopsis" binding:"max=10000"`
	CoverImageURL    *string `json:"cover_image_url,omitempty" binding:"omitempty,url"`
	ThumbnailURL     *string `json:"thumbnail_url,omitempty" binding:"omitempty,url"`
	Status           string  `json:"status" binding:"required,oneof=draft ongoing completed hiatus dropped"`
	IsOneshot        bool    `json:"is_oneshot"`
	OriginalLanguage *string `json:"original_language,omitempty" binding:"omitempty,len=2"` // ISO 639-1
	OriginalTitle    *string `json:"original_title,omitempty" binding:"omitempty,max=500"`
	OwnerID          string   `json:"owner_id" binding:"required,uuid"`
	OwnerType        string   `json:"owner_type" binding:"required,oneof=user group"`
	GenreIDs         []string `json:"genre_ids" binding:"omitempty,dive,uuid"`
	AuthorIDs        []string `json:"author_ids" binding:"omitempty,dive,uuid"`
	ArtistIDs        []string `json:"artist_ids" binding:"omitempty,dive,uuid"`
	Metadata         *string  `json:"metadata,omitempty"` // JSON string
}

// UpdateNovelRequest là DTO cho việc cập nhật novel
type UpdateNovelRequest struct {
	Title            string  `json:"title" binding:"required,min=1,max=500"`
	Synopsis         string  `json:"synopsis" binding:"max=10000"`
	CoverImageURL    *string `json:"cover_image_url,omitempty" binding:"omitempty,url"`
	ThumbnailURL     *string `json:"thumbnail_url,omitempty" binding:"omitempty,url"`
	Status           string  `json:"status" binding:"required,oneof=draft ongoing completed hiatus dropped"`
	OriginalLanguage *string `json:"original_language,omitempty" binding:"omitempty,len=2"` // ISO 639-1
	OriginalTitle    *string `json:"original_title,omitempty" binding:"omitempty,max=500"`
	Metadata         *string `json:"metadata,omitempty"` // JSON string
}

// NovelResponse là DTO cho response novel (cho danh sách)
type NovelResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	ThumbnailURL  *string `json:"thumbnail_url,omitempty"`
	Status        string  `json:"status"`
	TotalChapters int     `json:"total_chapters"`
	ViewCount     int64   `json:"view_count"`
	FavoriteCount int     `json:"favorite_count"`
	RatingAverage float64 `json:"rating_average"`
	RatingCount   int     `json:"rating_count"`
	LastChapterAt *string `json:"last_chapter_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ListNovelsRequest là DTO cho việc lấy danh sách novels
type ListNovelsRequest struct {
	Page             int    `form:"page" binding:"omitempty,min=1"`
	Limit            int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search           string `form:"search" binding:"omitempty,max=200"`                                      // Tìm kiếm trong title và synopsis
	Status           string `form:"status" binding:"omitempty,oneof=draft ongoing completed hiatus dropped"` // Filter theo status
	OriginalLanguage string `form:"original_language" binding:"omitempty,len=2"`                             // Filter theo ngôn ngữ gốc
	SortBy           string `form:"sort_by" binding:"omitempty,oneof=created_at rating views last_chapter"`  // created_at, rating, views, last_chapter
	SortOrder        string `form:"sort_order" binding:"omitempty,oneof=asc desc"`                           // asc hoặc desc
}

// NovelDetailResponse là DTO chi tiết cho novel
type NovelDetailResponse struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Slug             string  `json:"slug"`
	Synopsis         string  `json:"synopsis"` // Extracted from JSONB
	CoverImageURL    *string `json:"cover_image_url,omitempty"`
	ThumbnailURL     *string `json:"thumbnail_url,omitempty"`
	Status           string  `json:"status"`
	OriginalLanguage *string `json:"original_language,omitempty"`
	OriginalTitle    *string `json:"original_title,omitempty"`
	TotalVolumes     int     `json:"total_volumes"`
	TotalChapters    int     `json:"total_chapters"`
	TotalWords       int64   `json:"total_words"`
	ViewCount        int64   `json:"view_count"`
	FavoriteCount    int     `json:"favorite_count"`
	RatingAverage    float64 `json:"rating_average"`
	RatingCount      int     `json:"rating_count"`
	Metadata         *string `json:"metadata,omitempty"` // JSON string
	FirstPublishedAt *string `json:"first_published_at,omitempty"`
	LastChapterAt    *string `json:"last_chapter_at,omitempty"`
	CompletedAt      *string `json:"completed_at,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}
