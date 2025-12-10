package auth

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
	I18nAuthForbidden     = i18nutil.AuthForbidden
)

// i18n message keys for Auth module
const (
	// Registration
	I18nRegistrationSuccess = "auth.registration_success"
	I18nRegistrationFailed  = "auth.registration_failed"
	I18nEmailAlreadyExists  = "auth.email_already_exists"
	I18nWeakPassword        = "auth.weak_password"
	I18nInvalidEmail        = "auth.invalid_email"

	// Authentication
	I18nInvalidCredentials = "auth.invalid_credentials"
	I18nInvalidToken       = "auth.invalid_token"
	I18nTokenExpired       = "auth.token_expired"
	I18nTokenAlreadyUsed   = "auth.token_already_used"
	I18nUnauthorized       = "auth.unauthorized"

	// Email verification
	I18nEmailVerified      = "auth.email_verified"
	I18nVerificationFailed = "auth.verification_failed"

	// Password reset
	I18nResetEmailSent = "auth.reset_email_sent"
	I18nResetFailed    = "auth.reset_failed"
	I18nPasswordReset  = "auth.password_reset"

	// Session
	I18nSessionExpired        = "auth.session_expired"
	I18nSessionNotFound       = "auth.session_not_found"
	I18nSessionInvalid        = "auth.session_invalid"
	I18nSessionCreationFailed = "auth.session_creation_failed"

	// Consent
	I18nConsentNotFound = "auth.consent_not_found"
	I18nConsentRevoked  = "auth.consent_revoked"
	I18nConsentDenied   = "auth.consent_denied"

	// Passkey/WebAuthn
	I18nPasskeyRegistered            = "auth.passkey_registered"
	I18nPasskeyDeleted               = "auth.passkey_deleted"
	I18nPasskeyUpdated               = "auth.passkey_updated"
	I18nPasskeyListRetrieved         = "auth.passkey_list_retrieved"
	I18nRegistrationBeginFailed      = "auth.registration_begin_failed"
	I18nPasskeyRegistrationFailed    = "auth.passkey_registration_failed"
	I18nAuthenticationBeginFailed    = "auth.authentication_begin_failed"
	I18nAuthenticationFailed         = "auth.authentication_failed"
	I18nAuthenticationSuccess        = "auth.authentication_success"
	I18nInvalidOrExpiredSession      = "auth.invalid_or_expired_session"
	I18nCredentialAlreadyExists      = "auth.credential_already_exists"
	I18nCredentialVerificationFailed = "auth.credential_verification_failed"
	I18nCredentialNotFound           = "auth.credential_not_found"
	I18nNoPasskeyRegistered          = "auth.no_passkey_registered"
	I18nCredentialsListed            = "auth.credentials_listed"
	I18nCredentialDeleted            = "auth.credential_deleted"
	I18nCredentialUpdated            = "auth.credential_updated"
	I18nListCredentialsFailed        = "auth.list_credentials_failed"
	I18nDeleteCredentialFailed       = "auth.delete_credential_failed"
	I18nUpdateCredentialFailed       = "auth.update_credential_failed"

	// Validation
	I18nMissingUserID         = "auth.missing_user_id"
	I18nInvalidUserID         = "auth.invalid_user_id"
	I18nInvalidCredentialData = "auth.invalid_credential_data"
	I18nInvalidCredentialID   = "auth.invalid_credential_id"
	I18nMissingAuthHeader     = "auth.missing_authorization_header"
	I18nValidationFailed      = i18nutil.ValidationFailed
)
