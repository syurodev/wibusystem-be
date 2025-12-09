package i18nkeys

// WebAuthn/Passkey message keys
const (
	// Success messages
	WebAuthnPasskeyRegistered      = "auth.passkey_registered"
	WebAuthnPasskeyDeleted         = "auth.passkey_deleted"
	WebAuthnPasskeyUpdated         = "auth.passkey_updated"
	WebAuthnCredentialsListed      = "auth.credentials_listed"
	WebAuthnAuthenticationSuccess  = "auth.authentication_success"
	WebAuthnCredentialDeleted      = "auth.credential_deleted"
	WebAuthnCredentialUpdated      = "auth.credential_updated"

	// Error messages - Registration
	WebAuthnRegistrationBeginFailed = "auth.registration_begin_failed"
	WebAuthnRegistrationFailed      = "auth.registration_failed"

	// Error messages - Authentication
	WebAuthnAuthenticationBeginFailed = "auth.authentication_begin_failed"
	WebAuthnAuthenticationFailed      = "auth.authentication_failed"

	// Error messages - Session
	WebAuthnInvalidOrExpiredSession = "auth.invalid_or_expired_session"
	WebAuthnSessionCreationFailed   = "auth.session_creation_failed"

	// Error messages - Credential
	WebAuthnCredentialAlreadyExists     = "auth.credential_already_exists"
	WebAuthnCredentialVerificationFailed = "auth.credential_verification_failed"
	WebAuthnCredentialNotFound          = "auth.credential_not_found"
	WebAuthnInvalidCredentialData       = "auth.invalid_credential_data"
	WebAuthnInvalidCredentialID         = "auth.invalid_credential_id"
	WebAuthnListCredentialsFailed       = "auth.list_credentials_failed"
	WebAuthnDeleteCredentialFailed      = "auth.delete_credential_failed"
	WebAuthnUpdateCredentialFailed      = "auth.update_credential_failed"

	// Error messages - User
	WebAuthnNoPasskeyRegistered = "auth.no_passkey_registered"
	WebAuthnMissingUserID       = "auth.missing_user_id"
)
