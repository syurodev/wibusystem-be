package creator

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

	// Filter by first content posted date
	if filter.FirstContentPostedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("us.first_content_posted_at >= $%d", argNum))
		args = append(args, *filter.FirstContentPostedFrom)
		argNum++
	}
	if filter.FirstContentPostedTo != nil {
		conditions = append(conditions, fmt.Sprintf("us.first_content_posted_at <= $%d", argNum))
		args = append(args, *filter.FirstContentPostedTo)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Determine sort column and order
	sortColumn := "us.last_content_updated_at"
	switch filter.SortBy {
	case "follower_count":
		sortColumn = "us.follower_count"
	case "works_count":
		sortColumn = "(COALESCE(us.novel_count, 0) + COALESCE(us.manga_count, 0) + COALESCE(us.anime_count, 0))"
	case "novel_count":
		sortColumn = "us.novel_count"
	case "created_at":
		sortColumn = "u.created_at"
	case "last_content_updated_at":
		sortColumn = "us.last_content_updated_at"
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM identify.users u
		LEFT JOIN identify.user_statistics us ON u.id = us.user_id
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

	// Fetch creators with LEFT JOIN to user_statistics
	selectQuery := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.email_verified, u.full_name, u.avatar_url,
			u.phone, u.status, u.created_at, u.updated_at, u.last_login_at,
			u.display_name, u.username, u.bio, u.is_verified,
			COALESCE(us.follower_count, 0) AS follower_count,
			COALESCE(us.novel_count, 0) AS novel_count,
			COALESCE(us.manga_count, 0) AS manga_count,
			COALESCE(us.anime_count, 0) AS anime_count,
			us.last_content_updated_at
		FROM identify.users u
		LEFT JOIN identify.user_statistics us ON u.id = us.user_id
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
		var followerCount, novelCount, mangaCount, animeCount int
		var lastContentUpdatedAt *interface{}

		err := rows.Scan(
			&user.ID, &user.Email, &user.EmailVerified, &user.FullName, &user.AvatarURL,
			&user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
			&user.DisplayName, &user.Username, &bioJSON, &user.IsVerified,
			&followerCount, &novelCount, &mangaCount, &animeCount, &lastContentUpdatedAt,
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
			User:                 user,
			FollowerCount:        followerCount,
			NovelCount:           novelCount,
			MangaCount:           mangaCount,
			AnimeCount:           animeCount,
			WorksCount:           novelCount + mangaCount + animeCount,
			LastContentUpdatedAt: nil, // Will be set if lastContentUpdatedAt is not nil
			TotalViews:           0,   // Will be populated from ClickHouse separately
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
	// First ensure record exists
	if err := r.EnsureUserStatisticsExists(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE identify.user_statistics
		SET last_content_updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// IncrementNovelCount increments the novel_count for a user
func (r *creatorRepository) IncrementNovelCount(ctx context.Context, userID uuid.UUID) error {
	// First ensure record exists
	if err := r.EnsureUserStatisticsExists(ctx, userID); err != nil {
		return err
	}

	query := `
		UPDATE identify.user_statistics
		SET novel_count = COALESCE(novel_count, 0) + 1,
		    last_content_updated_at = NOW()
		WHERE user_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// DecrementNovelCount decrements the novel_count for a user
func (r *creatorRepository) DecrementNovelCount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE identify.user_statistics
		SET novel_count = GREATEST(COALESCE(novel_count, 0) - 1, 0)
		WHERE user_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

// EnsureUserStatisticsExists ensures a user_statistics record exists for the user
func (r *creatorRepository) EnsureUserStatisticsExists(ctx context.Context, userID uuid.UUID) error {
	query := `
		INSERT INTO identify.user_statistics (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
