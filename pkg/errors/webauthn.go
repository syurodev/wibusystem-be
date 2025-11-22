package errors

import "errors"

// WebAuthn/Passkey errors - Lỗi liên quan đến WebAuthn/Passkey authentication
var (
	ErrNoPasskeyRegistered        = errors.New("no passkey registered for this user")
	ErrCredentialNotFound         = errors.New("credential not found")
	ErrCredentialAlreadyExists    = errors.New("credential already exists")
	ErrInvalidSession             = errors.New("invalid session")
	ErrInvalidOrExpiredSession    = errors.New("invalid or expired session")
	ErrRegistrationFailed         = errors.New("passkey registration failed")
	ErrAuthenticationFailed       = errors.New("passkey authentication failed")
	ErrInvalidCredentialData      = errors.New("invalid credential data")
	ErrCredentialVerificationFailed = errors.New("credential verification failed")
)
