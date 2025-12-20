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
	ID            uuid.UUID      `json:"id"`
	Email         string         `json:"email"`
	EmailVerified bool           `json:"email_verified"`
	PasswordHash  string         `json:"-"`
	FullName      *string        `json:"full_name"`
	AvatarURL     *string        `json:"avatar_url"`
	Phone         *string        `json:"phone"`
	Status        UserStatus     `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	Settings      map[string]any `json:"settings"`
	// Profile fields
	DisplayName *string `json:"display_name"`
	Username    *string `json:"username"`
	Bio         []any   `json:"bio"` // JSONB array (platejs TNode[])
	IsVerified  bool    `json:"is_verified"`
}

// UserRepository định nghĩa interface cho việc truy cập dữ liệu user
type UserRepository interface {
	// GetByID lấy user theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)

	// GetByEmail lấy user theo email
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername lấy user theo username
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Create tạo user mới
	Create(ctx context.Context, user *User) error

	// Update cập nhật thông tin user
	Update(ctx context.Context, user *User) error

	// UpdateLastLogin cập nhật thời gian đăng nhập cuối
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	// GetGlobalPermissions lấy danh sách global permissions của user
	GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)

	// GetOrganizationPermissions lấy danh sách permissions của user trong organization
	GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)

	// GetGlobalRoles lấy danh sách global roles của user
	GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error)

	// GetOrganizationRoles lấy danh sách roles của user trong organization
	GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)

	// UpdateFirstContentPostedAt updates first content posted time if not set
	UpdateFirstContentPostedAt(ctx context.Context, userID uuid.UUID, postedAt time.Time) error
}
