package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// creatorRepository implements CreatorRepository using pgx
type creatorRepository struct {
	pool *pgxpool.Pool
}

// NewCreatorRepository creates a new instance of creatorRepository
func NewCreatorRepository(pool *pgxpool.Pool) domain.CreatorRepository {
	return &creatorRepository{pool: pool}
}

// ListCreators returns paginated list of creators with filters
func (r *creatorRepository) ListCreators(ctx context.Context, filter domain.CreatorListFilter) (*domain.CreatorListResult, error) {
	// Build dynamic query
	var conditions []string
	var args []any
	argNum := 1

	// Base condition: only active users
	conditions = append(conditions, "u.status = 'active'")

	// Filter by role (via user_global_roles)
	if filter.Role != nil && *filter.Role != "" {
		conditions = append(conditions, fmt.Sprintf(`
			EXISTS (
				SELECT 1 FROM identify.user_global_roles ugr
				JOIN identify.roles r ON ugr.role_id = r.id
				WHERE ugr.user_id = u.id AND r.name = $%d
			)
		`, argNum))
		args = append(args, *filter.Role)
		argNum++
	}

	// Search by display_name or username
	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + strings.ToLower(*filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf(`
			(LOWER(u.display_name) LIKE $%d OR LOWER(u.username) LIKE $%d)
		`, argNum, argNum))
		args = append(args, searchPattern)
		argNum++
	}

	// Filter by created date range
	if filter.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("u.created_at >= $%d", argNum))
		args = append(args, *filter.CreatedFrom)
		argNum++
	}
	if filter.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("u.created_at <= $%d", argNum))
		args = append(args, *filter.CreatedTo)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Determine sort column and order
	sortColumn := "u.last_content_updated_at"
	switch filter.SortBy {
	case "follower_count":
		sortColumn = "u.follower_count"
	case "works_count":
		sortColumn = "u.works_count"
	case "created_at":
		sortColumn = "u.created_at"
	case "last_content_updated_at":
		sortColumn = "u.last_content_updated_at"
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM identify.users u
		WHERE %s
	`, whereClause)

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count creators: %w", err)
	}

	// Calculate pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Fetch creators
	selectQuery := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.email_verified, u.full_name, u.avatar_url,
			u.phone, u.status, u.created_at, u.updated_at, u.last_login_at,
			u.display_name, u.username, u.bio, u.is_verified,
			u.follower_count, u.works_count, u.last_content_updated_at
		FROM identify.users u
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereClause, sortColumn, sortOrder, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query creators: %w", err)
	}
	defer rows.Close()

	var creators []domain.CreatorWithStats
	for rows.Next() {
		var user domain.User
		var bioJSON []byte

		err := rows.Scan(
			&user.ID, &user.Email, &user.EmailVerified, &user.FullName, &user.AvatarURL,
			&user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
			&user.DisplayName, &user.Username, &bioJSON, &user.IsVerified,
			&user.FollowerCount, &user.WorksCount, &user.LastContentUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan creator: %w", err)
		}

		// Parse bio JSON
		if len(bioJSON) > 0 {
			if err := json.Unmarshal(bioJSON, &user.Bio); err != nil {
				// Ignore JSON parse errors, bio will be nil
				user.Bio = nil
			}
		}

		creators = append(creators, domain.CreatorWithStats{
			User:       user,
			TotalViews: 0, // Will be populated from ClickHouse separately
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate creators: %w", err)
	}

	return &domain.CreatorListResult{
		Creators:   creators,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateLastContentUpdatedAt updates the last content update timestamp
func (r *creatorRepository) UpdateLastContentUpdatedAt(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE identify.users
		SET last_content_updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// IncrementWorksCount increments the works_count for a user
func (r *creatorRepository) IncrementWorksCount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE identify.users
		SET works_count = COALESCE(works_count, 0) + 1
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// DecrementWorksCount decrements the works_count for a user
func (r *creatorRepository) DecrementWorksCount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE identify.users
		SET works_count = GREATEST(COALESCE(works_count, 0) - 1, 0)
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
