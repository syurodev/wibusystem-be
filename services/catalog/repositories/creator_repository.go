package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreatorRepository defines CRUD and listing operations for creators
type CreatorRepository interface {
	Create(ctx context.Context, creator *m.Creator) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.Creator, error)
	GetByName(ctx context.Context, name string) (*m.Creator, error)
	Update(ctx context.Context, creator *m.Creator) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int, search string) ([]*m.Creator, int64, error)
}

type creatorRepository struct {
	pool *pgxpool.Pool
}

// NewCreatorRepository creates a Postgres-backed CreatorRepository
func NewCreatorRepository(pool *pgxpool.Pool) CreatorRepository {
	return &creatorRepository{pool: pool}
}

// Create inserts a new creator and returns the assigned ID and timestamps
func (r *creatorRepository) Create(ctx context.Context, creator *m.Creator) error {
	query := `
		INSERT INTO creator (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, creator.Name, creator.Description).Scan(
		&creator.ID,
		&creator.CreatedAt,
		&creator.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create creator: %w", err)
	}

	return nil
}

// GetByID retrieves a creator by its ID
func (r *creatorRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Creator, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM creator
		WHERE id = $1
	`

	var creator m.Creator
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&creator.ID,
		&creator.Name,
		&creator.Description,
		&creator.CreatedAt,
		&creator.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get creator by ID: %w", err)
	}

	return &creator, nil
}

// GetByName retrieves a creator by its name
func (r *creatorRepository) GetByName(ctx context.Context, name string) (*m.Creator, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM creator
		WHERE LOWER(name) = LOWER($1)
	`

	var creator m.Creator
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&creator.ID,
		&creator.Name,
		&creator.Description,
		&creator.CreatedAt,
		&creator.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get creator by name: %w", err)
	}

	return &creator, nil
}

// Update modifies an existing creator
func (r *creatorRepository) Update(ctx context.Context, creator *m.Creator) error {
	query := `
		UPDATE creator
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
		RETURNING updated_at
	`

	creator.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, creator.ID, creator.Name, creator.Description, creator.UpdatedAt).Scan(
		&creator.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update creator: %w", err)
	}

	return nil
}

// Delete removes a creator by ID
func (r *creatorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM creator WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete creator: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("creator not found")
	}

	return nil
}

// List retrieves paginated creators with optional search
func (r *creatorRepository) List(ctx context.Context, limit, offset int, search string) ([]*m.Creator, int64, error) {
	var creators []*m.Creator
	var total int64

	// Base query with search condition
	whereClause := ""
	args := []interface{}{limit, offset}

	if search != "" {
		whereClause = "WHERE LOWER(name) LIKE LOWER($3) OR LOWER(description) LIKE LOWER($3)"
		args = append(args, "%"+strings.ToLower(search)+"%")
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM creator %s", whereClause)
	var countArgs []interface{}
	if search != "" {
		countArgs = []interface{}{"%"+strings.ToLower(search)+"%"}
	}

	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count creators: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at
		FROM creator
		%s
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list creators: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var creator m.Creator
		err := rows.Scan(
			&creator.ID,
			&creator.Name,
			&creator.Description,
			&creator.CreatedAt,
			&creator.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan creator: %w", err)
		}
		creators = append(creators, &creator)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("failed to iterate creators: %w", rows.Err())
	}

	return creators, total, nil
}