package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// TransferStatus định nghĩa trạng thái của ownership transfer
type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusApproved  TransferStatus = "approved"
	TransferStatusRejected  TransferStatus = "rejected"
	TransferStatusCancelled TransferStatus = "cancelled"
	TransferStatusCompleted TransferStatus = "completed"
)

// OwnershipTransfer là domain model cho việc chuyển quyền sở hữu novel
type OwnershipTransfer struct {
	ID uuid.UUID

	NovelID uuid.UUID
	Novel   *Novel // Optional: loaded by JOIN

	// From owner
	FromOwnerType string // "user" or "organization"
	FromOwnerID   uuid.UUID

	// To owner
	ToOwnerType string // "user" or "organization"
	ToOwnerID   uuid.UUID

	Status TransferStatus
	Reason *string

	// Approval workflow
	RequiresApproval bool
	ReviewedBy       *uuid.UUID
	ReviewNotes      *string
	ReviewedAt       *time.Time

	// Timestamps
	RequestedAt *time.Time
	CompletedAt *time.Time

	// Audit
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// OwnershipTransferFilter định nghĩa các filter cho query
type OwnershipTransferFilter struct {
	NovelID       *uuid.UUID
	FromOwnerType *string
	FromOwnerID   *uuid.UUID
	ToOwnerType   *string
	ToOwnerID     *uuid.UUID
	Status        *TransferStatus
	SortBy        string
	SortOrder     string
	Limit         int
	Offset        int
}

// OwnershipTransferRepository định nghĩa interface cho việc truy cập dữ liệu
type OwnershipTransferRepository interface {
	// GetByID lấy transfer theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*OwnershipTransfer, error)

	// List lấy danh sách transfers với filter
	List(ctx context.Context, filter OwnershipTransferFilter) ([]*OwnershipTransfer, int64, error)

	// GetPendingByNovelID lấy các transfer đang pending cho novel
	GetPendingByNovelID(ctx context.Context, novelID uuid.UUID) ([]*OwnershipTransfer, error)

	// Create tạo transfer request mới
	Create(ctx context.Context, transfer *OwnershipTransfer) error

	// Update cập nhật transfer
	Update(ctx context.Context, transfer *OwnershipTransfer) error

	// Delete xóa mềm transfer
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error

	// Approve duyệt transfer request
	Approve(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, notes *string) error

	// Reject từ chối transfer request
	Reject(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, notes *string) error

	// Complete hoàn thành transfer (sau khi đã được approve)
	Complete(ctx context.Context, id uuid.UUID) error
}
