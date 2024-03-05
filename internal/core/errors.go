package core

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidArg     = errors.New("invalid arguments")
	ErrNotImplemented = errors.New("not implemented")
	ErrAlreadyExists  = errors.New("already exists")
)
