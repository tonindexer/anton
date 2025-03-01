package core

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidArg     = errors.New("invalid arguments")
	ErrNotImplemented = errors.New("not implemented")
	ErrNotAvailable   = errors.New("not available")
	ErrAlreadyExists  = errors.New("already exists")
)
