package db

import "errors"

var (
	ErrInvalidIndex = errors.New("index does not exist or out of range")
	ErrInvalidType  = errors.New("unexpected operation for type")
	ErrEmptyKey     = errors.New("empty key")
	ErrNotFound     = errors.New("expected key not found")
)
