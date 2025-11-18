package errors

import "errors"

// Novel related errors
var (
	ErrNovelNotFound = errors.New("novel not found")
	ErrNovelInUse    = errors.New("novel is in use")
)
