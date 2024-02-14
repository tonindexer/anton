package core

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidArg    = errors.New("invalid arguments")
	ErrAlreadyExists = errors.New("already exists")
)
