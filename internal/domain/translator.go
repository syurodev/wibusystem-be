package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Translator là domain model cho người dịch
type Translator struct {
	ID     uuid.UUID
	UserID *uuid.UUID

	Name string
	Slug string

	Biography json.RawMessage
	AvatarURL *string

	// Languages: ["vi", "en", "zh", "ja", "ko"]
	Languages json.RawMessage

	// Thống kê
	NovelCount         int
	ChapterCount       int
	WordCount          int64
	ContributionCount  int
	FollowerCount      int

	// Chất lượng (0-5)
	RatingAverage float64
	RatingCount   int

	IsVerified bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// TranslatorRepository định nghĩa interface cho việc truy cập dữ liệu translator
type TranslatorRepository interface {
	// GetByID lấy translator theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Translator, error)

	// GetBySlug lấy translator theo slug
	GetBySlug(ctx context.Context, slug string) (*Translator, error)

	// GetByUserID lấy translator theo user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Translator, error)

	// List lấy danh sách translators với filter
	List(ctx context.Context, filter TranslatorFilter) ([]*Translator, int64, error)

	// Create tạo translator mới
	Create(ctx context.Context, translator *Translator) error

	// Update cập nhật translator
	Update(ctx context.Context, translator *Translator) error

	// Delete xóa mềm translator
	Delete(ctx context.Context, id uuid.UUID) error

	// GetNovelTranslators lấy danh sách translators của một novel
	GetNovelTranslators(ctx context.Context, novelID uuid.UUID) ([]*NovelTranslator, error)

	// AddNovelTranslator thêm translator cho novel
	AddNovelTranslator(ctx context.Context, novelID, translatorID uuid.UUID, targetLanguage, role string, displayOrder int) error

	// RemoveNovelTranslator xóa translator khỏi novel
	RemoveNovelTranslator(ctx context.Context, novelID, translatorID uuid.UUID, targetLanguage string) error

	// UpdateStatistics cập nhật thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats TranslatorStatistics) error
}

// TranslatorFilter định nghĩa các filter cho việc query translators
type TranslatorFilter struct {
	SearchQuery *string
	Languages   []string // Filter by supported languages
	IsVerified  *bool
	MinRating   *float64
	SortBy      string // "name", "rating", "chapter_count", "follower_count"
	SortOrder   string // "asc", "desc"
	Limit       int
	Offset      int
}

// TranslatorStatistics chứa thông tin thống kê để update
type TranslatorStatistics struct {
	NovelCount        *int
	ChapterCount      *int
	WordCount         *int64
	ContributionCount *int
	FollowerCount     *int
	RatingAverage     *float64
	RatingCount       *int
}

// NovelTranslator chứa thông tin về translator và role trong novel
type NovelTranslator struct {
	Translator     *Translator
	TargetLanguage string // Ngôn ngữ đích
	Role           string // "lead_translator", "translator", "proofreader"
	DisplayOrder   int
}
