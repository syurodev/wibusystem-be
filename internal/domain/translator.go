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
	UserID uuid.UUID // Required - link to user account

	Name string
	Slug string

	Bio       *string
	AvatarURL *string

	// Languages arrays
	SourceLanguages []string `db:"source_languages"` // Array of ISO 639-1 codes
	TargetLanguages []string `db:"target_languages"` // Array of ISO 639-1 codes

	// Thống kê
	TranslationCount     int   `db:"translation_count"`
	TotalWordsTranslated int64 `db:"total_words_translated"`

	// Chất lượng (0-5)
	QualityScore float64 `db:"quality_score"`

	// Metadata (JSONB)
	Metadata json.RawMessage

	// Audit fields
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
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
	AddNovelTranslator(ctx context.Context, novelID, translatorID uuid.UUID, language, role string, displayOrder int) error

	// RemoveNovelTranslator xóa translator khỏi novel
	RemoveNovelTranslator(ctx context.Context, novelID, translatorID uuid.UUID, language string) error

	// UpdateStatistics cập nhật thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats TranslatorStatistics) error
}

// TranslatorFilter định nghĩa các filter cho việc query translators
type TranslatorFilter struct {
	SearchQuery     *string
	SourceLanguages []string // Filter by source languages
	TargetLanguages []string // Filter by target languages
	MinQualityScore *float64
	SortBy          string // "name", "quality_score", "translation_count"
	SortOrder       string // "asc", "desc"
	Limit           int
	Offset          int
}

// TranslatorStatistics chứa thông tin thống kê để update
type TranslatorStatistics struct {
	TranslationCount     *int
	TotalWordsTranslated *int64
	QualityScore         *float64
}

// NovelTranslator chứa thông tin về translator và role trong novel
type NovelTranslator struct {
	Translator   *Translator
	Language     string // Target language for translation
	Role         string // "lead_translator", "translator", "proofreader"
	DisplayOrder int
}
