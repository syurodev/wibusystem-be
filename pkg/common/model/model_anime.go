package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Anime represents an anime series master table
type Anime struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	Status          string           `json:"status" db:"status"` // content_status enum
	CoverImage      *string          `json:"cover_image,omitempty" db:"cover_image"`
	BroadcastSeason *string          `json:"broadcast_season,omitempty" db:"broadcast_season"` // season_name enum
	BroadcastYear   *int             `json:"broadcast_year,omitempty" db:"broadcast_year"`
	Summary         *json.RawMessage `json:"summary,omitempty" db:"summary"` // JSONB field
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

// AnimeSeason represents seasons of an anime series
type AnimeSeason struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	AnimeID            uuid.UUID `json:"anime_id" db:"anime_id"`
	SeasonNumber       int       `json:"season_number" db:"season_number"`
	SeasonTitle        *string   `json:"season_title,omitempty" db:"season_title"`
	PriceCoins         *int      `json:"price_coins,omitempty" db:"price_coins"`
	RentalPriceCoins   *int      `json:"rental_price_coins,omitempty" db:"rental_price_coins"`
	RentalDurationDays *int      `json:"rental_duration_days,omitempty" db:"rental_duration_days"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// AnimeEpisode represents individual episodes within an anime season
type AnimeEpisode struct {
	ID              uuid.UUID `json:"id" db:"id"`
	SeasonID        uuid.UUID `json:"season_id" db:"season_id"`
	EpisodeNumber   int       `json:"episode_number" db:"episode_number"`
	Title           *string   `json:"title,omitempty" db:"title"`
	DurationSeconds *int      `json:"duration_seconds,omitempty" db:"duration_seconds"`
	VideoURL        *string   `json:"video_url,omitempty" db:"video_url"`
	IsPublic        bool      `json:"is_public" db:"is_public"`
	PriceCoins      *int      `json:"price_coins,omitempty" db:"price_coins"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// EpisodeSubtitle represents subtitle files for episodes
type EpisodeSubtitle struct {
	ID           uuid.UUID `json:"id" db:"id"`
	EpisodeID    uuid.UUID `json:"episode_id" db:"episode_id"`
	LanguageCode string    `json:"language_code" db:"language_code"`
	SubtitleURL  *string   `json:"subtitle_url,omitempty" db:"subtitle_url"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
