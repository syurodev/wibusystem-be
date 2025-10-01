package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	d "wibusystem/pkg/common/dto"
)

// NovelQueryRepository interface defines read-only query methods for novel data access
// This follows CQRS pattern - separating complex read operations from write operations
type NovelQueryRepository interface {
	GetNovelWithFullDetails(ctx context.Context, id uuid.UUID) (*d.NovelDetailResponse, error)
}

// novelQueryRepository implements NovelQueryRepository interface
type novelQueryRepository struct {
	pool *pgxpool.Pool
}

// NewNovelQueryRepository creates a new novel query repository instance
func NewNovelQueryRepository(pool *pgxpool.Pool) NovelQueryRepository {
	return &novelQueryRepository{pool: pool}
}

// GetNovelWithFullDetails retrieves a novel with all related data in a single optimized query
// This uses JOINs and JSON aggregation to fetch genres, creators, and characters efficiently
func (r *novelQueryRepository) GetNovelWithFullDetails(ctx context.Context, id uuid.UUID) (*d.NovelDetailResponse, error) {
	query := `
		SELECT
			n.id,
			n.name,
			n.cover_image,
			n.summary,
			n.status,
			n.published_at,
			n.original_language,
			n.source_url,
			n.isbn,
			n.age_rating,
			n.content_warnings,
			n.mature_content,
			n.is_public,
			n.is_featured,
			n.is_completed,
			n.slug,
			n.keywords,
			n.price_coins,
			n.rental_price_coins,
			n.rental_duration_days,
			n.is_premium,
			n.view_count,
			n.rating_average,
			n.rating_count,
			n.total_chapters,
			n.total_volumes,
			n.created_at,
			n.updated_at,
			-- Aggregate genres as JSON array
			COALESCE(
				json_agg(
					DISTINCT jsonb_build_object(
						'id', g.id,
						'name', g.name
					)
				) FILTER (WHERE g.id IS NOT NULL),
				'[]'::json
			) as genres,
			-- Aggregate creators as JSON array
			COALESCE(
				json_agg(
					DISTINCT jsonb_build_object(
						'id', c.id,
						'name', c.name,
						'role', nc.role
					)
				) FILTER (WHERE c.id IS NOT NULL),
				'[]'::json
			) as creators,
			-- Aggregate characters as JSON array
			COALESCE(
				json_agg(
					DISTINCT jsonb_build_object(
						'id', ch.id,
						'name', ch.name
					)
				) FILTER (WHERE ch.id IS NOT NULL),
				'[]'::json
			) as characters
		FROM novel n
		LEFT JOIN novel_genre ng ON n.id = ng.novel_id
		LEFT JOIN genre g ON ng.genre_id = g.id
		LEFT JOIN novel_creator nc ON n.id = nc.novel_id
		LEFT JOIN creator c ON nc.creator_id = c.id
		LEFT JOIN novel_character nch ON n.id = nch.novel_id
		LEFT JOIN character ch ON nch.character_id = ch.id
		WHERE n.id = $1 AND n.is_deleted = FALSE
		GROUP BY n.id
	`

	var response d.NovelDetailResponse
	var summaryBytes, contentWarningsBytes []byte
	var genresJSON, creatorsJSON, charactersJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&response.ID,
		&response.Name,
		&response.CoverImage,
		&summaryBytes,
		&response.Status,
		&response.PublishedAt,
		&response.OriginalLanguage,
		&response.SourceURL,
		&response.ISBN,
		&response.AgeRating,
		&contentWarningsBytes,
		&response.MatureContent,
		&response.IsPublic,
		&response.IsFeatured,
		&response.IsCompleted,
		&response.Slug,
		&response.Keywords,
		&response.PriceCoins,
		&response.RentalPriceCoins,
		&response.RentalDurationDays,
		&response.IsPremium,
		&response.ViewCount,
		&response.RatingAverage,
		&response.RatingCount,
		&response.ChapterCount,
		&response.VolumeCount,
		&response.CreatedAt,
		&response.UpdatedAt,
		&genresJSON,
		&creatorsJSON,
		&charactersJSON,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("novel not found")
		}
		return nil, fmt.Errorf("failed to get novel with full details: %w", err)
	}

	// Parse summary JSON
	if summaryBytes != nil {
		var summaryData map[string]interface{}
		if err := json.Unmarshal(summaryBytes, &summaryData); err == nil {
			response.Summary = summaryData
		} else {
			response.Summary = make(map[string]interface{})
		}
	} else {
		response.Summary = make(map[string]interface{})
	}

	// Parse content warnings JSON
	if contentWarningsBytes != nil {
		var warnings []string
		if err := json.Unmarshal(contentWarningsBytes, &warnings); err == nil {
			response.ContentWarnings = warnings
		} else {
			response.ContentWarnings = []string{}
		}
	} else {
		response.ContentWarnings = []string{}
	}

	// Parse genres JSON
	if err := json.Unmarshal(genresJSON, &response.Genres); err != nil {
		response.Genres = []d.GenreInfo{}
	}

	// Parse creators JSON
	if err := json.Unmarshal(creatorsJSON, &response.Creators); err != nil {
		response.Creators = []d.CreatorInfo{}
	}

	// Parse characters JSON
	if err := json.Unmarshal(charactersJSON, &response.Characters); err != nil {
		response.Characters = []d.CharacterInfo{}
	}

	// Initialize empty maps/slices for optional fields
	response.Translations = []interface{}{}
	response.Stats = make(map[string]interface{})

	return &response, nil
}