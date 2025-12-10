package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// AssignmentStatus định nghĩa trạng thái của organization assignment
type AssignmentStatus string

const (
	AssignmentStatusActive    AssignmentStatus = "active"
	AssignmentStatusInactive  AssignmentStatus = "inactive"
	AssignmentStatusSuspended AssignmentStatus = "suspended"
)

// NovelOrganizationAssignment là domain model cho việc gán organization dịch novel
type NovelOrganizationAssignment struct {
	ID uuid.UUID

	NovelID        uuid.UUID
	Novel          *Novel // Optional: loaded by JOIN
	OrganizationID uuid.UUID
	Organization   *Organization // Optional: loaded by JOIN

	// Language for translation
	Language string // ISO 639-1 code

	Status             AssignmentStatus
	HasExclusiveRights bool // Claim exclusive translation rights

	// Statistics
	ChaptersTranslated int
	ChaptersProofread  int

	// Metadata (JSONB)
	Metadata json.RawMessage

	// Timestamps
	AssignedAt     time.Time
	LastActivityAt *time.Time

	// Audit
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NovelOrganizationAssignmentFilter định nghĩa các filter cho query
type NovelOrganizationAssignmentFilter struct {
	NovelID            *uuid.UUID
	OrganizationID     *uuid.UUID
	Language           *string
	Status             *AssignmentStatus
	HasExclusiveRights *bool
	SortBy             string
	SortOrder          string
	Limit              int
	Offset             int
}

// NovelOrganizationAssignmentRepository định nghĩa interface cho việc truy cập dữ liệu
type NovelOrganizationAssignmentRepository interface {
	// GetByID lấy assignment theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*NovelOrganizationAssignment, error)

	// GetByNovelAndLanguage lấy assignment theo novel và language
	GetByNovelAndLanguage(ctx context.Context, novelID uuid.UUID, language string) ([]*NovelOrganizationAssignment, error)

	// GetByOrganization lấy danh sách assignments của organization
	GetByOrganization(ctx context.Context, orgID uuid.UUID, filter NovelOrganizationAssignmentFilter) ([]*NovelOrganizationAssignment, int64, error)

	// List lấy danh sách assignments với filter
	List(ctx context.Context, filter NovelOrganizationAssignmentFilter) ([]*NovelOrganizationAssignment, int64, error)

	// Create tạo assignment mới
	Create(ctx context.Context, assignment *NovelOrganizationAssignment) error

	// Update cập nhật assignment
	Update(ctx context.Context, assignment *NovelOrganizationAssignment) error

	// Delete xóa mềm assignment
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error

	// UpdateStatistics cập nhật số liệu thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, chaptersTranslated, chaptersProofread int) error

	// ClaimExclusiveRights claim quyền dịch độc quyền
	ClaimExclusiveRights(ctx context.Context, id uuid.UUID) error

	// RevokeExclusiveRights thu hồi quyền dịch độc quyền
	RevokeExclusiveRights(ctx context.Context, id uuid.UUID) error
}
