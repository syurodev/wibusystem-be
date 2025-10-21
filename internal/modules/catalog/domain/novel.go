// Package domain contains business entities for the Catalog module.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Novel status constants
type NovelStatus string

const (
	NovelStatusDraft     NovelStatus = "DRAFT"
	NovelStatusOngoing   NovelStatus = "ONGOING"
	NovelStatusCompleted NovelStatus = "COMPLETED"
	NovelStatusHiatus    NovelStatus = "HIATUS"
	NovelStatusCancelled NovelStatus = "CANCELLED"
)

// Ownership type constants
type OwnershipType string

const (
	OwnershipPersonal      OwnershipType = "PERSONAL"
	OwnershipTenant        OwnershipType = "TENANT"
	OwnershipCollaborative OwnershipType = "COLLABORATIVE"
)

// Access level constants
type AccessLevel string

const (
	AccessPrivate    AccessLevel = "PRIVATE"
	AccessTenantOnly AccessLevel = "TENANT_ONLY"
	AccessPublic     AccessLevel = "PUBLIC"
)

// Age rating constants
type AgeRating string

const (
	AgeRatingG    AgeRating = "G"     // General Audiences
	AgeRatingPG   AgeRating = "PG"    // Parental Guidance
	AgeRatingPG13 AgeRating = "PG-13" // Parents Strongly Cautioned
	AgeRatingR    AgeRating = "R"     // Restricted
	AgeRatingNC17 AgeRating = "NC-17" // Adults Only
)

// Novel represents a novel series in the catalog
type Novel struct {
	// Core identification
	ID     uuid.UUID   `json:"id"`
	Status NovelStatus `json:"status"`

	// Content fields
	Title      string          `json:"title"`
	CoverImage string          `json:"cover_image,omitempty"`
	Summary    json.RawMessage `json:"summary,omitempty"`

	// Ownership model
	OwnershipType          OwnershipType `json:"ownership_type"`
	PrimaryOwnerID         uuid.UUID     `json:"primary_owner_id"`
	OriginalCreatorID      uuid.UUID     `json:"original_creator_id"`
	AccessLevel            AccessLevel   `json:"access_level"`
	LastModifiedByUserID   *uuid.UUID    `json:"last_modified_by_user_id,omitempty"`
	OwnershipTransferredAt *time.Time    `json:"ownership_transferred_at,omitempty"`

	// Publishing information
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	OriginalLanguage string     `json:"original_language"`
	SourceURL        string     `json:"source_url,omitempty"`
	ISBN             string     `json:"isbn,omitempty"`

	// Content rating and warnings
	AgeRating       AgeRating       `json:"age_rating,omitempty"`
	ContentWarnings json.RawMessage `json:"content_warnings,omitempty"`
	MatureContent   bool            `json:"mature_content"`

	// Visibility and status flags
	IsPublic        bool       `json:"is_public"`
	IsFeatured      bool       `json:"is_featured"`
	IsCompleted     bool       `json:"is_completed"`
	IsDeleted       bool       `json:"is_deleted"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	DeletedByUserID *uuid.UUID `json:"deleted_by_user_id,omitempty"`

	// SEO and discovery
	Slug            string          `json:"slug"`
	Tags            json.RawMessage `json:"tags,omitempty"`
	Keywords        string          `json:"keywords,omitempty"`
	MetaDescription string          `json:"meta_description,omitempty"`

	// Analytics and engagement
	ViewCount     int64   `json:"view_count"`
	LikeCount     int64   `json:"like_count"`
	BookmarkCount int64   `json:"bookmark_count"`
	CommentCount  int64   `json:"comment_count"`
	RatingAverage float64 `json:"rating_average"`
	RatingCount   int     `json:"rating_count"`

	// Pricing and monetization
	PriceCoins         int  `json:"price_coins,omitempty"`
	RentalPriceCoins   int  `json:"rental_price_coins,omitempty"`
	RentalDurationDays int  `json:"rental_duration_days,omitempty"`
	IsPremium          bool `json:"is_premium"`

	// Content metadata
	TotalChapters        int `json:"total_chapters"`
	TotalVolumes         int `json:"total_volumes"`
	EstimatedReadingTime int `json:"estimated_reading_time,omitempty"`
	WordCount            int `json:"word_count,omitempty"`

	// Audit timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var languageRegex = regexp.MustCompile(`^[a-z]{2}$`) // ISO 639-1

// NewNovel creates a new novel entity
func NewNovel(
	title string,
	slug string,
	ownershipType OwnershipType,
	primaryOwnerID uuid.UUID,
	originalCreatorID uuid.UUID,
	originalLanguage string,
) (*Novel, error) {
	novel := &Novel{
		ID:                uuid.New(),
		Status:            NovelStatusDraft,
		Title:             title,
		Slug:              slug,
		OwnershipType:     ownershipType,
		PrimaryOwnerID:    primaryOwnerID,
		OriginalCreatorID: originalCreatorID,
		AccessLevel:       AccessPrivate,
		OriginalLanguage:  originalLanguage,
		IsPublic:          false,
		IsFeatured:        false,
		IsCompleted:       false,
		IsDeleted:         false,
		MatureContent:     false,
		IsPremium:         false,
		ViewCount:         0,
		LikeCount:         0,
		BookmarkCount:     0,
		CommentCount:      0,
		RatingAverage:     0,
		RatingCount:       0,
		TotalChapters:     0,
		TotalVolumes:      0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := novel.Validate(); err != nil {
		return nil, err
	}

	return novel, nil
}

// Validate validates the novel entity
func (n *Novel) Validate() error {
	if n.ID == uuid.Nil {
		return ErrInvalidNovelID
	}

	if err := ValidateNovelTitle(n.Title); err != nil {
		return err
	}

	if err := ValidateSlug(n.Slug); err != nil {
		return err
	}

	if !n.Status.IsValid() {
		return ErrInvalidNovelStatus
	}

	if !n.OwnershipType.IsValid() {
		return ErrInvalidOwnershipType
	}

	if !n.AccessLevel.IsValid() {
		return ErrInvalidAccessLevel
	}

	if n.PrimaryOwnerID == uuid.Nil {
		return ErrInvalidOwnerID
	}

	if n.OriginalCreatorID == uuid.Nil {
		return ErrInvalidCreatorID
	}

	if err := ValidateLanguageCode(n.OriginalLanguage); err != nil {
		return err
	}

	if n.AgeRating != "" && !n.AgeRating.IsValid() {
		return ErrInvalidAgeRating
	}

	if err := n.ValidatePricing(); err != nil {
		return err
	}

	return nil
}

// ValidateNovelTitle validates novel title
func ValidateNovelTitle(title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrInvalidNovelTitle
	}
	if len(title) < 1 {
		return ErrNovelTitleTooShort
	}
	if len(title) > 500 {
		return ErrNovelTitleTooLong
	}
	return nil
}

// ValidateSlug validates slug format
func ValidateSlug(slug string) error {
	if slug == "" {
		return ErrInvalidSlug
	}
	if len(slug) < 3 {
		return ErrSlugTooShort
	}
	if len(slug) > 200 {
		return ErrSlugTooLong
	}
	if !slugRegex.MatchString(slug) {
		return ErrInvalidSlug
	}
	return nil
}

// ValidateLanguageCode validates ISO 639-1 language code
func ValidateLanguageCode(code string) error {
	if !languageRegex.MatchString(code) {
		return ErrInvalidLanguage
	}
	return nil
}

// ValidatePricing validates pricing fields
func (n *Novel) ValidatePricing() error {
	if n.PriceCoins < 0 {
		return ErrNegativePrice
	}
	if n.RentalPriceCoins < 0 {
		return ErrNegativePrice
	}
	if n.RentalDurationDays < 0 {
		return ErrInvalidRentalDuration
	}
	if n.RentalPriceCoins > 0 && n.RentalDurationDays == 0 {
		return ErrInvalidRentalDuration
	}
	return nil
}

// IsValid checks if status is valid
func (s NovelStatus) IsValid() bool {
	switch s {
	case NovelStatusDraft, NovelStatusOngoing, NovelStatusCompleted,
		NovelStatusHiatus, NovelStatusCancelled:
		return true
	}
	return false
}

// IsValid checks if ownership type is valid
func (o OwnershipType) IsValid() bool {
	switch o {
	case OwnershipPersonal, OwnershipTenant, OwnershipCollaborative:
		return true
	}
	return false
}

// IsValid checks if access level is valid
func (a AccessLevel) IsValid() bool {
	switch a {
	case AccessPrivate, AccessTenantOnly, AccessPublic:
		return true
	}
	return false
}

// IsValid checks if age rating is valid
func (a AgeRating) IsValid() bool {
	switch a {
	case AgeRatingG, AgeRatingPG, AgeRatingPG13, AgeRatingR, AgeRatingNC17:
		return true
	}
	return false
}

// UpdateTitle updates the novel title
func (n *Novel) UpdateTitle(title string) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if err := ValidateNovelTitle(title); err != nil {
		return err
	}
	n.Title = title
	n.UpdatedAt = time.Now()
	return nil
}

// UpdateStatus updates the novel status
func (n *Novel) UpdateStatus(status NovelStatus) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if !status.IsValid() {
		return ErrInvalidNovelStatus
	}
	n.Status = status
	n.UpdatedAt = time.Now()
	return nil
}

// Publish marks the novel as published
func (n *Novel) Publish() error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	now := time.Now()
	n.PublishedAt = &now
	n.IsPublic = true
	if n.Status == NovelStatusDraft {
		n.Status = NovelStatusOngoing
	}
	n.UpdatedAt = now
	return nil
}

// Complete marks the novel as completed
func (n *Novel) Complete() error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	n.IsCompleted = true
	n.Status = NovelStatusCompleted
	n.UpdatedAt = time.Now()
	return nil
}

// SetFeatured sets the featured status
func (n *Novel) SetFeatured(featured bool) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	n.IsFeatured = featured
	n.UpdatedAt = time.Now()
	return nil
}

// SetAccessLevel updates the access level
func (n *Novel) SetAccessLevel(level AccessLevel) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if !level.IsValid() {
		return ErrInvalidAccessLevel
	}
	n.AccessLevel = level
	n.UpdatedAt = time.Now()
	return nil
}

// SetMatureContent sets the mature content flag
func (n *Novel) SetMatureContent(mature bool) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	n.MatureContent = mature
	n.UpdatedAt = time.Now()
	return nil
}

// SetAgeRating sets the age rating
func (n *Novel) SetAgeRating(rating AgeRating) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if rating != "" && !rating.IsValid() {
		return ErrInvalidAgeRating
	}
	n.AgeRating = rating
	n.UpdatedAt = time.Now()
	return nil
}

// UpdatePricing updates pricing information
func (n *Novel) UpdatePricing(price, rentalPrice, rentalDays int, premium bool) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if price < 0 || rentalPrice < 0 || rentalDays < 0 {
		return ErrNegativePrice
	}
	if rentalPrice > 0 && rentalDays == 0 {
		return ErrInvalidRentalDuration
	}
	n.PriceCoins = price
	n.RentalPriceCoins = rentalPrice
	n.RentalDurationDays = rentalDays
	n.IsPremium = premium
	n.UpdatedAt = time.Now()
	return nil
}

// IncrementViewCount increments the view count
func (n *Novel) IncrementViewCount() {
	n.ViewCount++
	n.UpdatedAt = time.Now()
}

// IncrementLikeCount increments the like count
func (n *Novel) IncrementLikeCount() {
	n.LikeCount++
	n.UpdatedAt = time.Now()
}

// DecrementLikeCount decrements the like count
func (n *Novel) DecrementLikeCount() {
	if n.LikeCount > 0 {
		n.LikeCount--
		n.UpdatedAt = time.Now()
	}
}

// IncrementBookmarkCount increments the bookmark count
func (n *Novel) IncrementBookmarkCount() {
	n.BookmarkCount++
	n.UpdatedAt = time.Now()
}

// DecrementBookmarkCount decrements the bookmark count
func (n *Novel) DecrementBookmarkCount() {
	if n.BookmarkCount > 0 {
		n.BookmarkCount--
		n.UpdatedAt = time.Now()
	}
}

// UpdateRating updates the rating information
func (n *Novel) UpdateRating(newRating float64, totalRatings int) error {
	if newRating < 0 || newRating > 5 {
		return errors.New("rating must be between 0 and 5")
	}
	n.RatingAverage = newRating
	n.RatingCount = totalRatings
	n.UpdatedAt = time.Now()
	return nil
}

// UpdateContentMetadata updates content metadata
func (n *Novel) UpdateContentMetadata(chapters, volumes, readingTime, wordCount int) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if chapters < 0 || volumes < 0 || readingTime < 0 || wordCount < 0 {
		return errors.New("metadata values cannot be negative")
	}
	n.TotalChapters = chapters
	n.TotalVolumes = volumes
	n.EstimatedReadingTime = readingTime
	n.WordCount = wordCount
	n.UpdatedAt = time.Now()
	return nil
}

// TransferOwnership transfers ownership to another user/tenant
func (n *Novel) TransferOwnership(newOwnerID uuid.UUID, transferredByUserID uuid.UUID) error {
	if n.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if newOwnerID == uuid.Nil {
		return ErrInvalidOwnerID
	}
	now := time.Now()
	n.PrimaryOwnerID = newOwnerID
	n.LastModifiedByUserID = &transferredByUserID
	n.OwnershipTransferredAt = &now
	n.UpdatedAt = now
	return nil
}

// Delete soft deletes the novel
func (n *Novel) Delete(deletedByUserID uuid.UUID) error {
	if n.IsDeleted {
		return ErrNovelDeleted
	}
	now := time.Now()
	n.IsDeleted = true
	n.DeletedAt = &now
	n.DeletedByUserID = &deletedByUserID
	n.UpdatedAt = now
	return nil
}

// Restore restores a soft deleted novel
func (n *Novel) Restore() error {
	if !n.IsDeleted {
		return errors.New("novel is not deleted")
	}
	n.IsDeleted = false
	n.DeletedAt = nil
	n.DeletedByUserID = nil
	n.UpdatedAt = time.Now()
	return nil
}

// CanBeModifiedBy checks if a user can modify the novel
func (n *Novel) CanBeModifiedBy(userID uuid.UUID) bool {
	if n.IsDeleted {
		return false
	}
	return n.OriginalCreatorID == userID || n.PrimaryOwnerID == userID
}

// IsOwnedBy checks if the novel is owned by a specific user/tenant
func (n *Novel) IsOwnedBy(ownerID uuid.UUID) bool {
	return n.PrimaryOwnerID == ownerID
}

// IsPublished checks if the novel is published
func (n *Novel) IsPublished() bool {
	return n.PublishedAt != nil && !n.IsDeleted
}

// IsAccessibleBy checks if a user can access the novel
func (n *Novel) IsAccessibleBy(userID uuid.UUID, tenantID *uuid.UUID) bool {
	if n.IsDeleted {
		return false
	}

	// Owner and creator always have access
	if n.IsOwnedBy(userID) || n.OriginalCreatorID == userID {
		return true
	}

	// Check access level
	switch n.AccessLevel {
	case AccessPublic:
		return true
	case AccessTenantOnly:
		return tenantID != nil && *tenantID == n.PrimaryOwnerID
	case AccessPrivate:
		return false
	}

	return false
}

// GenerateSlugFromTitle generates a URL-friendly slug from the title
func GenerateSlugFromTitle(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Remove special characters and replace spaces with hyphens
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 200 {
		slug = slug[:200]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}

// Clone creates a deep copy of the novel
func (n *Novel) Clone() *Novel {
	clone := *n

	if n.PublishedAt != nil {
		t := *n.PublishedAt
		clone.PublishedAt = &t
	}

	if n.DeletedAt != nil {
		t := *n.DeletedAt
		clone.DeletedAt = &t
	}

	if n.OwnershipTransferredAt != nil {
		t := *n.OwnershipTransferredAt
		clone.OwnershipTransferredAt = &t
	}

	if n.LastModifiedByUserID != nil {
		id := *n.LastModifiedByUserID
		clone.LastModifiedByUserID = &id
	}

	if n.DeletedByUserID != nil {
		id := *n.DeletedByUserID
		clone.DeletedByUserID = &id
	}

	return &clone
}

// String returns a string representation of the novel
func (n *Novel) String() string {
	return fmt.Sprintf("Novel{ID: %s, Title: %s, Status: %s, Slug: %s}",
		n.ID, n.Title, n.Status, n.Slug)
}
