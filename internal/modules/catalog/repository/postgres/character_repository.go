package postgres

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CharacterRepositoryPG is a PostgreSQL implementation of CharacterRepository.
type CharacterRepositoryPG struct {
	DB *pgxpool.Pool
}

// NewCharacterRepositoryPG creates a new instance of CharacterRepositoryPG.
func NewCharacterRepositoryPG(db *pgxpool.Pool) *CharacterRepositoryPG {
	return &CharacterRepositoryPG{DB: db}
}

// Create inserts a new character into the database.
func (r *CharacterRepositoryPG) Create(ctx context.Context, character *domain.Character) error {
	query := `
		INSERT INTO characters (id, name, description, image_url, novel_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Exec(ctx, query,
		character.ID, character.Name, character.Description, character.ImageURL, character.NovelID, character.CreatedAt, character.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.CharacterRepositoryPG.Create: %w", err)
	}

	return nil
}

// GetByID retrieves a character by its ID.
func (r *CharacterRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Character, error) {
	query := `SELECT id, name, description, image_url, novel_id, created_at, updated_at FROM characters WHERE id = $1`
	character := &domain.Character{}
	row := r.DB.QueryRow(ctx, query, id)

	err := row.Scan(
		&character.ID, &character.Name, &character.Description, &character.ImageURL, &character.NovelID,
		&character.CreatedAt, &character.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("character with id %s not found", id)
		}
		return nil, fmt.Errorf("postgres.CharacterRepositoryPG.GetByID: %w", err)
	}

	return character, nil
}

// Update modifies an existing character in the database.
func (r *CharacterRepositoryPG) Update(ctx context.Context, character *domain.Character) error {
	query := `
		UPDATE characters SET
			name = $2, description = $3, image_url = $4, novel_id = $5, updated_at = $6
		WHERE id = $1`

	_, err := r.DB.Exec(ctx, query,
		character.ID, character.Name, character.Description, character.ImageURL, character.NovelID, character.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.CharacterRepositoryPG.Update: %w", err)
	}

	return nil
}

// Delete removes a character from the database.
func (r *CharacterRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM characters WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.CharacterRepositoryPG.Delete: %w", err)
	}
	return nil
}

// ListByNovelID retrieves all characters for a given novel.
func (r *CharacterRepositoryPG) ListByNovelID(ctx context.Context, novelID uuid.UUID, limit, offset int) ([]*domain.Character, int64, error) {
	query := `
		SELECT id, name, description, image_url, novel_id, created_at, updated_at
		FROM characters
		WHERE novel_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, query, novelID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.CharacterRepositoryPG.ListByNovelID: %w", err)
	}
	defer rows.Close()

	characters := []*domain.Character{}
	for rows.Next() {
		character := &domain.Character{}
		err := rows.Scan(
			&character.ID, &character.Name, &character.Description, &character.ImageURL, &character.NovelID,
			&character.CreatedAt, &character.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("postgres.CharacterRepositoryPG.ListByNovelID: failed to scan character: %w", err)
		}
		characters = append(characters, character)
	}

	countQuery := `SELECT COUNT(*) FROM characters WHERE novel_id = $1`
	var total int64
	err = r.DB.QueryRow(ctx, countQuery, novelID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.CharacterRepositoryPG.ListByNovelID: failed to count characters: %w", err)
	}

	return characters, total, nil
}
