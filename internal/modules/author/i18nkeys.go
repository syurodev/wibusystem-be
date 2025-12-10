package author

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for author module
const (
	// Success messages
	I18nCreatedSuccess = "author.created_success"
	I18nUpdatedSuccess = "author.updated_success"
	I18nDeletedSuccess = "author.deleted_success"
	I18nGetSuccess     = "author.get_success"
	I18nListSuccess    = "author.list_success"
	I18nMergeSuccess   = "author.merge_success"
	I18nPreviewSuccess = "author.preview_success"

	// Error messages
	I18nCreateFailed  = "author.create_failed"
	I18nUpdateFailed  = "author.update_failed"
	I18nDeleteFailed  = "author.delete_failed"
	I18nGetFailed     = "author.get_failed"
	I18nListFailed    = "author.list_failed"
	I18nMergeFailed   = "author.merge_failed"
	I18nPreviewFailed = "author.preview_failed"

	// Validation and business errors
	I18nNotFound          = "author.not_found"
	I18nInvalidID         = "author.invalid_id"
	I18nInvalidInput      = "author.invalid_input"
	I18nSlugAlreadyExists = "author.slug_already_exists"
	I18nInUse             = "author.in_use"
)


