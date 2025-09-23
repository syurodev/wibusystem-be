package model

import (
	"time"

	"github.com/google/uuid"
)

// Manga represents a manga series master table
type Manga struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	Status     string          `json:"status" db:"status"`                         // content_status enum
	CoverImage *string         `json:"cover_image,omitempty" db:"cover_image"`
	Summary    *ContentSummary `json:"summary,omitempty" db:"summary"`             // JSONB field
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// MangaVolume represents volumes of a manga series
type MangaVolume struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	MangaID            uuid.UUID `json:"manga_id" db:"manga_id"`
	VolumeNumber       int       `json:"volume_number" db:"volume_number"`
	VolumeTitle        *string   `json:"volume_title,omitempty" db:"volume_title"`
	CoverImage         *string   `json:"cover_image,omitempty" db:"cover_image"`
	Description        *string   `json:"description,omitempty" db:"description"`
	PriceCoins         *int      `json:"price_coins,omitempty" db:"price_coins"`
	RentalPriceCoins   *int      `json:"rental_price_coins,omitempty" db:"rental_price_coins"`
	RentalDurationDays *int      `json:"rental_duration_days,omitempty" db:"rental_duration_days"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// MangaChapter represents chapters of a manga
type MangaChapter struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	VolumeID      uuid.UUID  `json:"volume_id" db:"volume_id"`
	ChapterNumber int        `json:"chapter_number" db:"chapter_number"`
	Title         *string    `json:"title,omitempty" db:"title"`
	ReleasedAt    *time.Time `json:"released_at,omitempty" db:"released_at"`
	IsPublic      bool       `json:"is_public" db:"is_public"`
	PriceCoins    *int       `json:"price_coins,omitempty" db:"price_coins"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// MangaPage represents individual pages of a manga chapter
type MangaPage struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ChapterID  uuid.UUID `json:"chapter_id" db:"chapter_id"`
	PageNumber int       `json:"page_number" db:"page_number"`
	ImageURL   *string   `json:"image_url,omitempty" db:"image_url"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}