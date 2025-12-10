package genre

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for genre module
const (
	// Success messages
	I18nCreatedSuccess = "genre.created_success"
	I18nUpdatedSuccess = "genre.updated_success"
	I18nDeletedSuccess = "genre.deleted_success"
	I18nGetSuccess     = "genre.get_success"
	I18nListSuccess    = "genre.list_success"

	// Error messages
	I18nCreateFailed = "genre.create_failed"
	I18nUpdateFailed = "genre.update_failed"
	I18nDeleteFailed = "genre.delete_failed"
	I18nGetFailed    = "genre.get_failed"
	I18nListFailed   = "genre.list_failed"

	// Validation and business errors
	I18nNotFound          = "genre.not_found"
	I18nInvalidID         = "genre.invalid_id"
	I18nInvalidInput      = "genre.invalid_input"
	I18nSlugAlreadyExists = "genre.slug_already_exists"
	I18nParentNotFound    = "genre.parent_not_found"
	I18nInvalidParentID   = "genre.invalid_parent_id"
	I18nInUse             = "genre.in_use"
	I18nCircularReference = "genre.circular_reference"
	I18nHasChildren       = "genre.has_children"
)


