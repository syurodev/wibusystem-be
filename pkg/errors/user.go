package errors

import "errors"

// User domain errors - Lỗi liên quan đến User
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserInactive      = errors.New("user is inactive")
	ErrUserBlocked       = errors.New("user is blocked")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrInvalidUserID     = errors.New("invalid user id")
)
