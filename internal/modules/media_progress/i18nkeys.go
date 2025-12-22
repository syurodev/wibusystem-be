package media_progress

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for media_progress module
const (
	// Success messages
	I18nGetSuccess          = "media_progress.get_success"
	I18nUpdateSuccess       = "media_progress.update_success"
	I18nDeleteSuccess       = "media_progress.delete_success"
	I18nClearSuccess        = "media_progress.clear_success"
	I18nMarkCompleteSuccess = "media_progress.mark_complete_success"

	// Error messages
	I18nGetFailed          = "media_progress.get_failed"
	I18nUpdateFailed       = "media_progress.update_failed"
	I18nDeleteFailed       = "media_progress.delete_failed"
	I18nClearFailed        = "media_progress.clear_failed"
	I18nMarkCompleteFailed = "media_progress.mark_complete_failed"

	// Validation errors
	I18nNotFound  = "media_progress.not_found"
	I18nInvalidID = "media_progress.invalid_id"
)
