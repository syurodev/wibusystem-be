package user

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
	I18nAuthForbidden     = i18nutil.AuthForbidden
)

// i18n message keys for user/account module
const (
	// Success messages
	I18nProfileFetchedSuccess  = "account.profile_fetched_success"
	I18nProfileUpdatedSuccess  = "account.profile_updated_success"
	I18nPasswordChangedSuccess = "account.password_changed_success"
	I18nSettingsFetchedSuccess = "account.settings_fetched_success"
	I18nSettingsUpdatedSuccess = "account.settings_updated_success"
	I18nSessionsFetchedSuccess = "account.sessions_fetched_success"
	I18nSessionDeletedSuccess  = "account.session_deleted_success"

	// Error messages
	I18nFetchUserError      = "account.fetch_user_error"
	I18nFetchSessionsError  = "account.fetch_sessions_error"
	I18nUpdateProfileError  = "account.update_profile_error"
	I18nUpdatePasswordError = "account.update_password_error"
	I18nUpdateSettingsError = "account.update_settings_error"
	I18nDeleteSessionError  = "account.delete_session_error"
	I18nHashError           = "account.hash_error"

	// Validation errors
	I18nInvalidUserID         = "account.invalid_user_id"
	I18nInvalidRequest        = "account.invalid_request"
	I18nInvalidSessionID      = "account.invalid_session_id"
	I18nInvalidPassword       = "account.invalid_password"
	I18nWeakPassword          = "account.weak_password"
	I18nUsernameImmutable     = "account.username_immutable"
	I18nUsernameLengthInvalid = "account.username_length_invalid"
	I18nUsernameTaken         = "account.username_taken"
	I18nForbidden             = "account.forbidden"
)
