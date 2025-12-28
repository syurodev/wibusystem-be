// ============================================================================
// User Repository (Ent Implementation)
// ============================================================================
//
// Repository này triển khai UserRepository interface sử dụng Ent ORM.
// Một số methods sử dụng raw SQL để gọi PostgreSQL functions.
//
// ============================================================================

package user

import (
	"context"
	"database/sql"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/user"
)

// userEntRepository triển khai UserRepository sử dụng Ent
type userEntRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewUserEntRepository tạo một instance mới của userEntRepository
// db parameter is the underlying *sql.DB from Ent driver for raw queries
func NewUserEntRepository(client *ent.Client, db *sql.DB) domain.UserRepository {
	return &userEntRepository{client: client, db: db}
}

// GetByID lấy user từ database theo ID
func (r *userEntRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := database.GetClientFromContext(ctx, r.client).User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entUserToDomain(u), nil
}

// GetByEmail lấy user từ database theo email
func (r *userEntRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := database.GetClientFromContext(ctx, r.client).User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entUserToDomain(u), nil
}

// GetByUsername lấy user từ database theo username
func (r *userEntRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, err := database.GetClientFromContext(ctx, r.client).User.Query().
		Where(user.UsernameEQ(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entUserToDomain(u), nil
}

// Create tạo user mới trong database
func (r *userEntRepository) Create(ctx context.Context, u *domain.User) error {
	builder := database.GetClientFromContext(ctx, r.client).User.Create().
		SetID(u.ID).
		SetEmail(u.Email).
		SetEmailVerified(u.EmailVerified).
		SetPasswordHash(u.PasswordHash).
		SetStatus(user.Status(u.Status))

	if u.FullName != nil {
		builder.SetFullName(*u.FullName)
	}
	if u.AvatarURL != nil {
		builder.SetAvatarURL(*u.AvatarURL)
	}
	if u.Phone != nil {
		builder.SetPhone(*u.Phone)
	}
	if u.Settings != nil {
		builder.SetSettings(u.Settings)
	}
	if u.DisplayName != nil {
		builder.SetDisplayName(*u.DisplayName)
	}
	if u.Username != nil {
		builder.SetUsername(*u.Username)
	}
	if u.Bio != nil {
		builder.SetBio(u.Bio)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật thông tin user
func (r *userEntRepository) Update(ctx context.Context, u *domain.User) error {
	builder := database.GetClientFromContext(ctx, r.client).User.UpdateOneID(u.ID).
		SetEmail(u.Email).
		SetEmailVerified(u.EmailVerified).
		SetPasswordHash(u.PasswordHash).
		SetStatus(user.Status(u.Status)).
		SetIsVerified(u.IsVerified)

	if u.FullName != nil {
		builder.SetFullName(*u.FullName)
	} else {
		builder.ClearFullName()
	}
	if u.AvatarURL != nil {
		builder.SetAvatarURL(*u.AvatarURL)
	} else {
		builder.ClearAvatarURL()
	}
	if u.Phone != nil {
		builder.SetPhone(*u.Phone)
	} else {
		builder.ClearPhone()
	}
	if u.Settings != nil {
		builder.SetSettings(u.Settings)
	} else {
		builder.ClearSettings()
	}
	if u.DisplayName != nil {
		builder.SetDisplayName(*u.DisplayName)
	} else {
		builder.ClearDisplayName()
	}
	if u.Username != nil {
		builder.SetUsername(*u.Username)
	} else {
		builder.ClearUsername()
	}
	if u.Bio != nil {
		builder.SetBio(u.Bio)
	} else {
		builder.ClearBio()
	}

	_, err := builder.Save(ctx)
	return err
}

// UpdateLastLogin cập nhật thời gian đăng nhập cuối
func (r *userEntRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).User.UpdateOneID(userID).
		SetLastLoginAt(time.Now()).
		Save(ctx)
	return err
}

// GetGlobalPermissions lấy danh sách global permissions của user
// Note: Sử dụng raw SQL vì gọi PostgreSQL function
func (r *userEntRepository) GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var permissions []string
	rows, err := r.db.QueryContext(ctx,
		"SELECT permission_name FROM identify.get_user_global_permissions($1)",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
// Note: Sử dụng raw SQL vì gọi PostgreSQL function
func (r *userEntRepository) GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	var permissions []string
	rows, err := r.db.QueryContext(ctx,
		"SELECT permission_name FROM identify.get_user_organization_permissions($1, $2)",
		userID, organizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, rows.Err()
}

// GetGlobalRoles lấy danh sách global roles của user
// Note: Sử dụng raw SQL vì gọi PostgreSQL function
func (r *userEntRepository) GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var roles []string
	rows, err := r.db.QueryContext(ctx,
		"SELECT role_name FROM identify.get_user_global_roles($1)",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
// Note: Sử dụng raw SQL vì gọi PostgreSQL function
func (r *userEntRepository) GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	var roles []string
	rows, err := r.db.QueryContext(ctx,
		"SELECT role_name FROM identify.get_user_organization_roles($1, $2)",
		userID, organizationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
// Note: Sử dụng raw SQL để chỉ update nếu chưa set
func (r *userEntRepository) UpdateFirstContentPostedAt(ctx context.Context, userID uuid.UUID, postedAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE identify.users 
		 SET first_content_posted_at = $2 
		 WHERE id = $1 AND first_content_posted_at IS NULL`,
		userID, postedAt,
	)
	return err
}

// =============================================================================
// Helper Functions
// =============================================================================

// entUserToDomain converts Ent User entity to domain.User
func entUserToDomain(u *ent.User) *domain.User {
	return &domain.User{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		PasswordHash:  u.PasswordHash,
		FullName:      u.FullName,
		AvatarURL:     u.AvatarURL,
		Phone:         u.Phone,
		Status:        domain.UserStatus(u.Status),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		LastLoginAt:   u.LastLoginAt,
		Settings:      u.Settings,
		DisplayName:   u.DisplayName,
		Username:      u.Username,
		Bio:           u.Bio,
		IsVerified:    u.IsVerified,
	}
}
