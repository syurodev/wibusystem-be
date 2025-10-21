package domain

import "errors"

var (
	// General
	ErrCannotModifyDeleted = errors.New("cannot modify deleted entity")

	// Slug
	ErrInvalidTitleForSlug = errors.New("invalid title for slug generation")

	// Novel errors
	ErrInvalidNovelID        = errors.New("invalid novel ID")
	ErrInvalidNovelTitle     = errors.New("invalid novel title")
	ErrNovelTitleTooShort    = errors.New("novel title too short")
	ErrNovelTitleTooLong     = errors.New("novel title too long")
	ErrInvalidNovelStatus    = errors.New("invalid novel status")
	ErrInvalidOwnershipType  = errors.New("invalid ownership type")
	ErrInvalidAccessLevel    = errors.New("invalid access level")
	ErrInvalidAgeRating      = errors.New("invalid age rating")
	ErrInvalidSlug           = errors.New("invalid slug")
	ErrSlugTooShort          = errors.New("slug too short")
	ErrSlugTooLong           = errors.New("slug too long")
	ErrInvalidOwnerID        = errors.New("invalid owner ID")
	ErrInvalidCreatorID      = errors.New("invalid creator ID")
	ErrInvalidLanguage       = errors.New("invalid language code")
	ErrInvalidPricing        = errors.New("invalid pricing")
	ErrNovelDeleted          = errors.New("novel is deleted")
	ErrInvalidISBN           = errors.New("invalid ISBN format")
	ErrNegativePrice         = errors.New("price cannot be negative")
	ErrInvalidRentalDuration = errors.New("invalid rental duration")

	// Volume errors
	ErrInvalidVolumeID      = errors.New("invalid volume ID")
	ErrInvalidVolumeNovelID = errors.New("invalid novel ID for volume")
	ErrInvalidVolumeNumber  = errors.New("invalid volume number")
	ErrVolumeDeleted        = errors.New("volume is deleted")
	ErrVolumeTitleTooLong   = errors.New("volume title too long")

	// Chapter errors
	ErrInvalidChapterID       = errors.New("invalid chapter ID")
	ErrInvalidChapterVolumeID = errors.New("invalid volume ID for chapter")
	ErrInvalidChapterNumber   = errors.New("invalid chapter number")
	ErrChapterDeleted         = errors.New("chapter is deleted")
	ErrChapterTitleTooLong    = errors.New("chapter title too long")
	ErrInvalidChapterVersion  = errors.New("invalid chapter version")

	// Creator errors
	ErrCreatorNameRequired = errors.New("creator name is required")
	ErrCreatorNameTooLong  = errors.New("creator name is too long")

	// Genre errors
	ErrGenreNameRequired = errors.New("genre name is required")
	ErrGenreNameTooLong  = errors.New("genre name is too long")

	// Character errors
	ErrCharacterNameRequired = errors.New("character name is required")
	ErrCharacterNameTooLong  = errors.New("character name is too long")
)
