package max

import "errors"

var (
	ErrDisconnected = errors.New("max: not connected")
	ErrRemote       = errors.New("max: remote error")
)
