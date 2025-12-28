package errors

import "errors"

// Common errors
var (
	ErrResourceNotFound  = errors.New("resource not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrInternalServer    = errors.New("internal server error")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrSlugAlreadyExists = errors.New("slug already exists")
)

// Artist errors (deprecated - use AppError)
var (
	ErrArtistNotFound = errors.New("artist not found")
	ErrArtistInUse    = errors.New("artist is in use")
)

// Author errors (deprecated - use AppError)
var (
	ErrAuthorNotFound = errors.New("author not found")
	ErrAuthorInUse    = errors.New("author is in use")
)

// Genre errors (deprecated - use AppError)
var (
	ErrGenreNotFound           = errors.New("genre not found")
	ErrGenreInUse              = errors.New("genre is in use")
	ErrGenreHasChildren        = errors.New("genre has children")
	ErrCircularParentReference = errors.New("circular parent reference")
	ErrParentGenreNotFound     = errors.New("parent genre not found")
)

// Novel errors (deprecated - use AppError)
var (
	ErrNovelNotFound = errors.New("novel not found")
)

// Volume errors (deprecated - use AppError)
var (
	ErrVolumeNotFound      = errors.New("volume not found")
	ErrVolumeNumberExists  = errors.New("volume number already exists")
	ErrVolumeHasChapters   = errors.New("volume has chapters")
	ErrInvalidVolumeNumber = errors.New("invalid volume number")
)

// Chapter errors (deprecated - use AppError)
var (
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrChapterNumberExists  = errors.New("chapter number already exists")
	ErrInvalidChapterStatus = errors.New("invalid chapter status")
)

// User errors (deprecated - use AppError)
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUsernameImmutable     = errors.New("username is immutable")
	ErrUsernameLengthInvalid = errors.New("username length is invalid")
	ErrUsernameTaken         = errors.New("username is already taken")
	ErrWeakPassword          = errors.New("password is too weak")
	ErrInvalidPassword       = errors.New("invalid password")
)

// Auth errors (deprecated - use AppError)
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenAlreadyUsed   = errors.New("token already used")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// WebAuthn errors (deprecated - use AppError)
var (
	ErrInvalidOrExpiredSession      = errors.New("invalid or expired session")
	ErrInvalidSession               = errors.New("invalid session")
	ErrCredentialAlreadyExists      = errors.New("credential already exists")
	ErrCredentialVerificationFailed = errors.New("credential verification failed")
	ErrNoPasskeyRegistered          = errors.New("no passkey registered")
	ErrAuthenticationFailed         = errors.New("authentication failed")
	ErrCredentialNotFound           = errors.New("credential not found")
)
