package errors

import "errors"

// Volume-related errors
var (
	ErrVolumeNotFound        = errors.New("volume not found")
	ErrVolumeNumberExists    = errors.New("volume number already exists for this novel")
	ErrVolumeInUse           = errors.New("volume is in use")
	ErrInvalidVolumeNumber   = errors.New("invalid volume number")
)
