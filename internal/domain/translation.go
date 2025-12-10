package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// TranslationStatus định nghĩa trạng thái của bản dịch
type TranslationStatus string

const (
	TranslationStatusDraft         TranslationStatus = "draft"
	TranslationStatusPendingReview TranslationStatus = "pending_review"
	TranslationStatusApproved      TranslationStatus = "approved"
	TranslationStatusRejected      TranslationStatus = "rejected"
	TranslationStatusPublished     TranslationStatus = "published"
)

// ContributionType định nghĩa loại đóng góp
type ContributionType string

const (
	ContributionTypeNewTranslation ContributionType = "new_translation"
	ContributionTypeImprovement    ContributionType = "improvement"
	ContributionTypeProofreading   ContributionType = "proofreading"
	ContributionTypeCorrection     ContributionType = "correction"
)

// ChapterTranslation là domain model cho bản dịch chính thức của chapter
type ChapterTranslation struct {
	ID        uuid.UUID
	ChapterID uuid.UUID
	Chapter   *NovelChapter // Optional: được load bởi JOIN query

	// Language code (ISO 639-1)
	Language string

	// Title và content bằng ngôn ngữ đích
	Title   string
	Content json.RawMessage

	// Translator notes
	TranslatorNotes json.RawMessage

	// Organization (nhóm dịch, nullable nếu cá nhân dịch)
	OrganizationID *uuid.UUID
	Organization   *Organization // Optional: được load bởi JOIN query

	// Version tracking
	Version int

	// Status
	Status TranslationStatus

	// Quality metrics
	WordCount      int
	CharacterCount int
	QualityScore   float64
	ReviewerRating float64

	// Thống kê
	ViewCount         int64
	LikeCount         int
	CommentCount      int
	ContributionCount int

	// Review tracking
	ReviewedBy  *uuid.UUID
	ReviewNotes *string
	ReviewedAt  *time.Time

	// Publishing
	PublishedAt *time.Time

	// Audit
	CreatedBy uuid.UUID
	Creator   *User // Optional: được load bởi JOIN query
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ChapterTranslationRepository định nghĩa interface cho việc truy cập dữ liệu translation
type ChapterTranslationRepository interface {
	// GetByID lấy translation theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*ChapterTranslation, error)

	// GetByChapterAndLanguage lấy translation theo chapter và language
	GetByChapterAndLanguage(ctx context.Context, chapterID uuid.UUID, language string) (*ChapterTranslation, error)

	// GetByChapterID lấy tất cả translations của một chapter
	GetByChapterID(ctx context.Context, chapterID uuid.UUID) ([]*ChapterTranslation, error)

	// GetByOrganizationID lấy danh sách translations của organization
	GetByOrganizationID(ctx context.Context, organizationID uuid.UUID, filter TranslationFilter) ([]*ChapterTranslation, error)

	// GetByCreatorID lấy danh sách translations của người tạo
	GetByCreatorID(ctx context.Context, creatorID uuid.UUID, filter TranslationFilter) ([]*ChapterTranslation, error)

	// Create tạo translation mới
	Create(ctx context.Context, translation *ChapterTranslation) error

	// Update cập nhật translation
	Update(ctx context.Context, translation *ChapterTranslation) error

	// Delete xóa mềm translation
	Delete(ctx context.Context, id uuid.UUID) error

	// Publish xuất bản translation
	Publish(ctx context.Context, id uuid.UUID) error

	// IncrementViewCount tăng view count
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// UpdateStatistics cập nhật thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats TranslationStatisticsUpdate) error
}

// TranslationFilter định nghĩa các filter cho việc query translations
type TranslationFilter struct {
	Language  *string
	Status    *TranslationStatus
	SortBy    string // "created_at", "views", "rating"
	SortOrder string // "asc", "desc"
	Limit     int
	Offset    int
}

// TranslationStatisticsUpdate chứa thông tin thống kê để update
type TranslationStatisticsUpdate struct {
	ViewCount     *int64
	LikeCount     *int
	RatingAverage *float64
	RatingCount   *int
}

// TranslationContribution là domain model cho đóng góp bản dịch từ community
type TranslationContribution struct {
	ID        uuid.UUID
	ChapterID uuid.UUID
	Chapter   *NovelChapter // Optional

	// Contributor
	ContributorID uuid.UUID
	Contributor   *User // Optional: được load bởi JOIN query

	// Target language
	Language string

	// Contribution type
	ContributionType ContributionType

	// Content
	Title            *string
	Content          json.RawMessage
	ContributorNotes *string

	// Status and review
	Status TranslationStatus

	// Reviewer information
	ReviewedBy   *uuid.UUID
	ReviewedAt   *time.Time
	ReviewNotes  *string

	// If approved, link to official translation
	OfficialTranslationID *uuid.UUID

	// Credit and rewards
	CreditPoints int
	IsCredited   bool

	// Metrics
	WordCount      int
	CharacterCount int

	// Community feedback
	UpvoteCount   int
	DownvoteCount int

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// TranslationContributionRepository định nghĩa interface cho việc truy cập dữ liệu contributions
type TranslationContributionRepository interface {
	// GetByID lấy contribution theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*TranslationContribution, error)

	// GetByChapterID lấy danh sách contributions của một chapter
	GetByChapterID(ctx context.Context, chapterID uuid.UUID, filter ContributionFilter) ([]*TranslationContribution, error)

	// GetByContributorID lấy danh sách contributions của contributor
	GetByContributorID(ctx context.Context, contributorID uuid.UUID, filter ContributionFilter) ([]*TranslationContribution, error)

	// GetPendingReview lấy danh sách contributions đang chờ review
	GetPendingReview(ctx context.Context, language *string, limit, offset int) ([]*TranslationContribution, error)

	// Create tạo contribution mới
	Create(ctx context.Context, contribution *TranslationContribution) error

	// Update cập nhật contribution
	Update(ctx context.Context, contribution *TranslationContribution) error

	// Delete xóa mềm contribution
	Delete(ctx context.Context, id uuid.UUID) error

	// Approve phê duyệt contribution
	Approve(ctx context.Context, id, reviewerID uuid.UUID, reviewNotes string, creditPoints int) error

	// Reject từ chối contribution
	Reject(ctx context.Context, id, reviewerID uuid.UUID, reviewNotes string) error

	// Vote vote cho contribution (upvote/downvote)
	Vote(ctx context.Context, contributionID, userID uuid.UUID, isUpvote bool) error

	// UpdateStatistics cập nhật thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats ContributionStatisticsUpdate) error
}

// ContributionFilter định nghĩa các filter cho việc query contributions
type ContributionFilter struct {
	Language         *string
	Status           *TranslationStatus
	ContributionType *ContributionType
	SortBy           string // "created_at", "upvotes", "credit_points"
	SortOrder        string // "asc", "desc"
	Limit            int
	Offset           int
}

// ContributionStatisticsUpdate chứa thông tin thống kê để update
type ContributionStatisticsUpdate struct {
	UpvoteCount   *int
	DownvoteCount *int
	CreditPoints  *int
}

// TranslationHistory là domain model cho lịch sử thay đổi bản dịch
type TranslationHistory struct {
	ID            uuid.UUID
	TranslationID uuid.UUID
	Version       int

	// Snapshot of content at this version
	Title   string
	Content json.RawMessage

	// Who made the change
	ChangedBy          *uuid.UUID
	ChangeDescription  *string

	// Metrics at this version
	WordCount int

	CreatedAt time.Time
}

// TranslationHistoryRepository định nghĩa interface cho việc truy cập translation history
type TranslationHistoryRepository interface {
	// GetByTranslationID lấy lịch sử của một translation
	GetByTranslationID(ctx context.Context, translationID uuid.UUID, limit, offset int) ([]*TranslationHistory, error)

	// GetVersion lấy một version cụ thể
	GetVersion(ctx context.Context, translationID uuid.UUID, version int) (*TranslationHistory, error)

	// Create tạo history record (thường được gọi bởi trigger)
	Create(ctx context.Context, history *TranslationHistory) error
}
