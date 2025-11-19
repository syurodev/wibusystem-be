package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// UserStatus định nghĩa trạng thái của user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User là domain model cho người dùng trong hệ thống
type User struct {
	ID            uuid.UUID
	Email         string
	EmailVerified bool
	PasswordHash  string
	FullName      *string
	AvatarURL     *string
	Phone         *string
	Status        UserStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
	Settings      map[string]any
}

// UserRepository định nghĩa interface cho việc truy cập dữ liệu user
type UserRepository interface {
	// GetByID lấy user theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail lấy user theo email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Create tạo user mới
	Create(ctx context.Context, user *User) error

	// Update cập nhật thông tin user
	Update(ctx context.Context, user *User) error

	// UpdateLastLogin cập nhật thời gian đăng nhập cuối
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}
