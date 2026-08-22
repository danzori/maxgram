package max

import "errors"

var (
	ErrDisconnected = errors.New("max: not connected")
	ErrRemote       = errors.New("max: remote error")

	ErrAuthRejected = errors.New("max: authorization rejected")
	ErrAuthTimeout  = errors.New("max: authorization did not complete in time")
	ErrSnapshot     = errors.New("max: cannot read the authorization snapshot")
)
