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

// CharacterRepository defines CRUD and listing operations for characters
type CharacterRepository interface {
	Create(ctx context.Context, character *m.Character) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.Character, error)
	GetByName(ctx context.Context, name string) (*m.Character, error)
	Update(ctx context.Context, character *m.Character) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int, search string) ([]*m.Character, int64, error)
}

type characterRepository struct {
	pool *pgxpool.Pool
}

// NewCharacterRepository creates a Postgres-backed CharacterRepository
func NewCharacterRepository(pool *pgxpool.Pool) CharacterRepository {
	return &characterRepository{pool: pool}
}

// Create inserts a new character and returns the assigned ID and timestamps
func (r *characterRepository) Create(ctx context.Context, character *m.Character) error {
	query := `
		INSERT INTO character (name, description, image_url)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, character.Name, character.Description, character.ImageURL).Scan(
		&character.ID,
		&character.CreatedAt,
		&character.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// GetByID retrieves a character by its ID
func (r *characterRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Character, error) {
	query := `
		SELECT id, name, description, image_url, created_at, updated_at
		FROM character
		WHERE id = $1
	`

	var character m.Character
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&character.ID,
		&character.Name,
		&character.Description,
		&character.ImageURL,
		&character.CreatedAt,
		&character.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get character by ID: %w", err)
	}

	return &character, nil
}

// GetByName retrieves a character by its name
func (r *characterRepository) GetByName(ctx context.Context, name string) (*m.Character, error) {
	query := `
		SELECT id, name, description, image_url, created_at, updated_at
		FROM character
		WHERE LOWER(name) = LOWER($1)
	`

	var character m.Character
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&character.ID,
		&character.Name,
		&character.Description,
		&character.ImageURL,
		&character.CreatedAt,
		&character.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get character by name: %w", err)
	}

	return &character, nil
}

// Update modifies an existing character
func (r *characterRepository) Update(ctx context.Context, character *m.Character) error {
	query := `
		UPDATE character
		SET name = $2, description = $3, image_url = $4, updated_at = $5
		WHERE id = $1
		RETURNING updated_at
	`

	character.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, character.ID, character.Name, character.Description, character.ImageURL, character.UpdatedAt).Scan(
		&character.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	return nil
}

// Delete removes a character by ID
func (r *characterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM character WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("character not found")
	}

	return nil
}

// List retrieves paginated characters with optional search
func (r *characterRepository) List(ctx context.Context, limit, offset int, search string) ([]*m.Character, int64, error) {
	var characters []*m.Character
	var total int64

	// Base query with search condition
	whereClause := ""
	args := []interface{}{limit, offset}

	if search != "" {
		whereClause = "WHERE LOWER(name) LIKE LOWER($3) OR LOWER(description) LIKE LOWER($3)"
		args = append(args, "%"+strings.ToLower(search)+"%")
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM character %s", whereClause)
	var countArgs []interface{}
	if search != "" {
		countArgs = []interface{}{"%"+strings.ToLower(search)+"%"}
	}

	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count characters: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, name, description, image_url, created_at, updated_at
		FROM character
		%s
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list characters: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var character m.Character
		err := rows.Scan(
			&character.ID,
			&character.Name,
			&character.Description,
			&character.ImageURL,
			&character.CreatedAt,
			&character.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan character: %w", err)
		}
		characters = append(characters, &character)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("failed to iterate characters: %w", rows.Err())
	}

	return characters, total, nil
}