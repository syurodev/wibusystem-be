package errors

import "errors"

// Chapter-related errors
var (
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrChapterNumberExists  = errors.New("chapter number already exists for this novel")
	ErrChapterInUse         = errors.New("chapter is in use")
	ErrInvalidChapterNumber = errors.New("invalid chapter number")
	ErrInvalidChapterStatus = errors.New("invalid chapter status")
)
