package novel_chapter

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for novel_chapter module
const (
	// Success messages
	I18nCreatedSuccess    = "chapter.created_success"
	I18nUpdatedSuccess    = "chapter.updated_success"
	I18nDeletedSuccess    = "chapter.deleted_success"
	I18nGetSuccess        = "chapter.get_success"
	I18nListSuccess       = "chapter.list_success"
	I18nReorderSuccess    = "chapter.reorder_success"
	I18nBulkCreateSuccess = "chapter.bulk_create_success"

	// Error messages
	I18nCreateFailed  = "chapter.create_failed"
	I18nUpdateFailed  = "chapter.update_failed"
	I18nDeleteFailed  = "chapter.delete_failed"
	I18nGetFailed     = "chapter.get_failed"
	I18nListFailed    = "chapter.list_failed"
	I18nReorderFailed = "chapter.reorder_failed"

	// Validation and business errors
	I18nNotFound             = "chapter.not_found"
	I18nInvalidID            = "chapter.invalid_id"
	I18nInvalidInput         = "chapter.invalid_input"
	I18nVolumeNotFound       = "chapter.volume_not_found"
	I18nChapterNumberExists  = "chapter.chapter_number_exists"
	I18nInvalidChapterStatus = "chapter.invalid_chapter_status"
)
