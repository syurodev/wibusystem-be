package novel_volume

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for novel_volume module
const (
	// Success messages
	I18nCreatedSuccess = "volume.created_success"
	I18nUpdatedSuccess = "volume.updated_success"
	I18nDeletedSuccess = "volume.deleted_success"
	I18nGetSuccess     = "volume.get_success"
	I18nListSuccess    = "volume.list_success"
	I18nReorderSuccess = "volume.reorder_success"

	// Error messages
	I18nCreateFailed  = "volume.create_failed"
	I18nUpdateFailed  = "volume.update_failed"
	I18nDeleteFailed  = "volume.delete_failed"
	I18nGetFailed     = "volume.get_failed"
	I18nListFailed    = "volume.list_failed"
	I18nReorderFailed = "volume.reorder_failed"

	// Validation and business errors
	I18nNotFound           = "volume.not_found"
	I18nInvalidID          = "volume.invalid_id"
	I18nInvalidInput       = "volume.invalid_input"
	I18nNovelNotFound      = "volume.novel_not_found"
	I18nVolumeNumberExists = "volume.volume_number_exists"
	I18nHasChapters        = "volume.has_chapters"
)
