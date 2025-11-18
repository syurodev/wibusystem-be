package errors

import "errors"

// Author related errors
var (
	ErrAuthorNotFound = errors.New("author not found")
	ErrAuthorInUse    = errors.New("author is in use by novels")
)
