package config

import "errors"

var (
	ErrMissingValue = errors.New("missing required environment variable")
	ErrInvalidValue = errors.New("invalid environment variable")
)
