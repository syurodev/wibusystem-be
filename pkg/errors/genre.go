package errors

import "errors"

// Genre errors - Lỗi liên quan đến genre
var (
	ErrGenreNotFound           = errors.New("genre not found")
	ErrSlugAlreadyExists       = errors.New("slug already exists")
	ErrParentGenreNotFound     = errors.New("parent genre not found")
	ErrCircularParentReference = errors.New("circular parent reference not allowed")
	ErrGenreInUse              = errors.New("genre is in use by novels")
	ErrGenreHasChildren        = errors.New("genre has children")
)
