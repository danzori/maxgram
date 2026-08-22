package topic

import "errors"

var (
	ErrNotFound     = errors.New("topic not found")
	ErrCreateDenied = errors.New("bot is not allowed to manage forum topics")
)
