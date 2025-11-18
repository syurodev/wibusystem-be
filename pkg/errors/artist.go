package errors

import "errors"

// Artist related errors
var (
	ErrArtistNotFound = errors.New("artist not found")
	ErrArtistInUse    = errors.New("artist is in use by novels")
)
