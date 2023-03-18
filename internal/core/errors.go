package core

import "errors"

var (
	ErrNotAvailable = errors.New("not available")
	ErrNotFound     = errors.New("not found")
	ErrInvalidArg   = errors.New("invalid arguments")
)
