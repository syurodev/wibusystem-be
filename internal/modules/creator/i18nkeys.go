package creator

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// I18n key constants cho Creator module
const (
	I18nCreatorListSuccess = "creator.list_success"
	I18nCreatorListFailed  = "creator.list_failed"
)
