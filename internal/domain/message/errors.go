package message

import "errors"

var (
	ErrEmpty  = errors.New("message has no content")
	ErrNoChat = errors.New("message has no chat")
)
