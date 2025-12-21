// ============================================================================
// User Repository
// ============================================================================
//
// Repository này triển khai UserRepository interface từ domain package.
// Quản lý dữ liệu user trong hệ thống authentication/authorization.
//
// CRUD Operations:
//   - GetByID: Lấy user theo ID
//   - GetByEmail: Lấy user theo email (cho login)
//   - GetByUsername: Lấy user theo username
//   - Create: Tạo user mới
//   - Update: Cập nhật thông tin user
//
// Authentication:
//   - UpdateLastLogin: Cập nhật thời gian đăng nhập cuối
//
// Authorization:
//   - GetGlobalPermissions: Lấy global permissions của user
//   - GetOrganizationPermissions: Lấy permissions trong organization
//   - GetGlobalRoles: Lấy global roles của user
//   - GetOrganizationRoles: Lấy roles trong organization
//
// Statistics:
//   - UpdateFirstContentPostedAt: Cập nhật thời điểm post content đầu tiên
//
// SQL queries được load từ thư mục queries/ sử dụng go:embed.
//
// ============================================================================

package user

import (
	"context"
	_ "embed"
	"encoding/json"
	"system/internal/domain"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL queries embedded từ files
//
//go:embed queries/get_by_id.sql
var getByIDQuery string

//go:embed queries/get_by_email.sql
var getByEmailQuery string

//go:embed queries/get_by_username.sql
var getByUsernameQuery string

//go:embed queries/create.sql
var createQuery string

//go:embed queries/update.sql
var updateQuery string

//go:embed queries/update_last_login.sql
var updateLastLoginQuery string

//go:embed queries/get_global_permissions.sql
var getGlobalPermissionsQuery string

//go:embed queries/get_organization_permissions.sql
var getOrganizationPermissionsQuery string

//go:embed queries/get_global_roles.sql
var getGlobalRolesQuery string

//go:embed queries/get_organization_roles.sql
var getOrganizationRolesQuery string

//go:embed queries/update_first_content_posted_at.sql
var updateFirstContentPostedAtQuery string

// userRepository triển khai UserRepository sử dụng pgx
type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository tạo một instance mới của userRepository
func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &userRepository{pool: pool}
}

// GetByID lấy user từ database theo ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, getByIDQuery, id)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByEmail lấy user từ database theo email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, getByEmailQuery, email)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByUsername lấy user từ database theo username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	rows, err := r.pool.Query(ctx, getByUsernameQuery, username)
	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Create tạo user mới trong database
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	settingsJSON, err := json.Marshal(user.Settings)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, createQuery,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Phone,
		user.Status,
		string(settingsJSON),
	)

	return err
}

// Update cập nhật thông tin user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	settingsJSON, err := json.Marshal(user.Settings)
	if err != nil {
		return err
	}

	bioJSON, err := json.Marshal(user.Bio)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, updateQuery,
		user.ID,
		user.Email,
		user.EmailVerified,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Phone,
		user.Status,
		string(settingsJSON),
		user.DisplayName,
		user.Username,
		string(bioJSON),
	)

	return err
}

// UpdateLastLogin cập nhật thời gian đăng nhập cuối
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, updateLastLoginQuery, userID)
	return err
}

// GetGlobalPermissions lấy danh sách global permissions của user
func (r *userRepository) GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, getGlobalPermissionsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}

	return permissions, rows.Err()
}

// GetOrganizationPermissions lấy danh sách permissions của user trong organization
func (r *userRepository) GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, getOrganizationPermissionsQuery, userID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var result struct {
			PermissionName string `db:"permission_name"`
		}
		err := rows.Scan(&result.PermissionName)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, result.PermissionName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetGlobalRoles lấy danh sách global roles của user
func (r *userRepository) GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, getGlobalRolesQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetOrganizationRoles lấy danh sách roles của user trong organization
func (r *userRepository) GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, getOrganizationRolesQuery, userID, organizationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// UpdateFirstContentPostedAt updates first content posted time if not set
func (r *userRepository) UpdateFirstContentPostedAt(ctx context.Context, userID uuid.UUID, postedAt time.Time) error {
	// Note: We use execute, if row not updated (because already set) it's fine.
	_, err := r.pool.Exec(ctx, updateFirstContentPostedAtQuery, userID, postedAt)
	return err
}
