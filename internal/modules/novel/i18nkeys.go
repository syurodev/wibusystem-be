package novel

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for novel module
const (
	// Success messages
	I18nCreatedSuccess = "novel.created_success"
	I18nUpdatedSuccess = "novel.updated_success"
	I18nDeletedSuccess = "novel.deleted_success"
	I18nGetSuccess     = "novel.get_success"
	I18nListSuccess    = "novel.list_success"

	// Error messages
	I18nCreateFailed = "novel.create_failed"
	I18nUpdateFailed = "novel.update_failed"
	I18nDeleteFailed = "novel.delete_failed"
	I18nGetFailed    = "novel.get_failed"
	I18nListFailed   = "novel.list_failed"

	// Validation and business errors
	I18nNotFound            = "novel.not_found"
	I18nInvalidID           = "novel.invalid_id"
	I18nInvalidInput        = "novel.invalid_input"
	I18nSlugAlreadyExists   = "novel.slug_already_exists"
	I18nInvalidStatus       = "novel.invalid_status"
	I18nInvalidOwnerID      = "novel.invalid_owner_id"
	I18nInvalidGenreID      = "novel.invalid_genre_id"
	I18nInvalidAuthorID     = "novel.invalid_author_id"
	I18nInvalidArtistID     = "novel.invalid_artist_id"
	I18nIncrementViewFailed = "novel.increment_view_failed"
	I18nViewIncremented     = "novel.view_incremented"
	I18nGetTopSuccess       = "novel.get_top_success"
)
