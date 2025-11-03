package errors

import "errors"

// OAuth2 Client errors - Lỗi OAuth2 Client
var (
	ErrClientNotFound      = errors.New("oauth2 client not found")
	ErrInvalidClient       = errors.New("invalid client credentials")
	ErrClientInactive      = errors.New("client is inactive")
	ErrClientAlreadyExists = errors.New("client already exists")
	ErrInvalidClientID     = errors.New("invalid client id")
	ErrInvalidRedirectURI  = errors.New("invalid redirect uri")
)

// OAuth2 Authorization errors - Lỗi OAuth2 Authorization
var (
	ErrAuthRequestNotFound = errors.New("authorization request not found")
	ErrAuthRequestExpired  = errors.New("authorization request expired")
	ErrInvalidAuthRequest  = errors.New("invalid authorization request")
)

// OAuth2 Token errors - Lỗi OAuth2 Token
var (
	ErrInvalidGrant         = errors.New("invalid grant")
	ErrInvalidScope         = errors.New("invalid scope")
	ErrUnsupportedGrantType = errors.New("unsupported grant type")
)
