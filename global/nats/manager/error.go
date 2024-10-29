package manager

import "errors"

var (
	ErrMsgHandlerNotExist = errors.New("msg handler not exist")

	ErrStreamNotExist = errors.New("stream not exist")
)
