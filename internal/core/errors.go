package core

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrInvalidArg = errors.New("invalid arguments")
	// ErrNotAvailable = errors.New("not available")
)
