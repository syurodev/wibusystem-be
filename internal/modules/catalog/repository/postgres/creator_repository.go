package postgres

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreatorRepositoryPG is a PostgreSQL implementation of CreatorRepository.
type CreatorRepositoryPG struct {
	DB *pgxpool.Pool
}

// NewCreatorRepositoryPG creates a new instance of CreatorRepositoryPG.
func NewCreatorRepositoryPG(db *pgxpool.Pool) *CreatorRepositoryPG {
	return &CreatorRepositoryPG{DB: db}
}

// Create inserts a new creator into the database.
func (r *CreatorRepositoryPG) Create(ctx context.Context, creator *domain.Creator) error {
	query := `
		INSERT INTO creators (id, name, bio, profile_image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.DB.Exec(ctx, query,
		creator.ID, creator.Name, creator.Bio, creator.ProfileImage, creator.CreatedAt, creator.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.CreatorRepositoryPG.Create: %w", err)
	}

	return nil
}

// GetByID retrieves a creator by their ID.
func (r *CreatorRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Creator, error) {
	query := `SELECT id, name, bio, profile_image, novel_count, followers, created_at, updated_at FROM creators WHERE id = $1`
	creator := &domain.Creator{}
	row := r.DB.QueryRow(ctx, query, id)

	err := row.Scan(
		&creator.ID, &creator.Name, &creator.Bio, &creator.ProfileImage,
		&creator.NovelCount, &creator.Followers, &creator.CreatedAt, &creator.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("creator with id %s not found", id)
		}
		return nil, fmt.Errorf("postgres.CreatorRepositoryPG.GetByID: %w", err)
	}

	return creator, nil
}

// Update modifies an existing creator in the database.
func (r *CreatorRepositoryPG) Update(ctx context.Context, creator *domain.Creator) error {
	query := `
		UPDATE creators SET
			name = $2, bio = $3, profile_image = $4, novel_count = $5, followers = $6, updated_at = $7
		WHERE id = $1`

	_, err := r.DB.Exec(ctx, query,
		creator.ID, creator.Name, creator.Bio, creator.ProfileImage,
		creator.NovelCount, creator.Followers, creator.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.CreatorRepositoryPG.Update: %w", err)
	}

	return nil
}

// Delete removes a creator from the database.
func (r *CreatorRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM creators WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.CreatorRepositoryPG.Delete: %w", err)
	}
	return nil
}

// List retrieves a paginated list of creators.
func (r *CreatorRepositoryPG) List(ctx context.Context, limit, offset int) ([]*domain.Creator, int64, error) {
	query := `SELECT id, name, bio, profile_image, novel_count, followers, created_at, updated_at FROM creators ORDER BY name ASC LIMIT $1 OFFSET $2`
	rows, err := r.DB.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.CreatorRepositoryPG.List: %w", err)
	}
	defer rows.Close()

	creators := []*domain.Creator{}
	for rows.Next() {
		creator := &domain.Creator{}
		err := rows.Scan(
			&creator.ID, &creator.Name, &creator.Bio, &creator.ProfileImage,
			&creator.NovelCount, &creator.Followers, &creator.CreatedAt, &creator.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("postgres.CreatorRepositoryPG.List: failed to scan creator: %w", err)
		}
		creators = append(creators, creator)
	}

	countQuery := `SELECT COUNT(*) FROM creators`
	var total int64
	err = r.DB.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.CreatorRepositoryPG.List: failed to count creators: %w", err)
	}

	return creators, total, nil
}
