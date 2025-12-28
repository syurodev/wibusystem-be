package artist

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for artist module
const (
	// Success messages
	I18nCreatedSuccess = "artist.created_success"
	I18nUpdatedSuccess = "artist.updated_success"
	I18nDeletedSuccess = "artist.deleted_success"
	I18nGetSuccess     = "artist.get_success"
	I18nListSuccess    = "artist.list_success"
	I18nMergeSuccess   = "artist.merge_success"
	I18nPreviewSuccess = "artist.preview_success"

	// Error messages
	I18nCreateFailed  = "artist.create_failed"
	I18nUpdateFailed  = "artist.update_failed"
	I18nDeleteFailed  = "artist.delete_failed"
	I18nGetFailed     = "artist.get_failed"
	I18nListFailed    = "artist.list_failed"
	I18nMergeFailed   = "artist.merge_failed"
	I18nPreviewFailed = "artist.preview_failed"

	// Validation and business errors
	I18nNotFound          = "artist.not_found"
	I18nInvalidID         = "artist.invalid_id"
	I18nInvalidInput      = "artist.invalid_input"
	I18nSlugAlreadyExists = "artist.slug_already_exists"
	I18nInUse             = "artist.in_use"
)
