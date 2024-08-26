package nats

import "errors"

var (
	ErrNatsNotWork        = errors.New("nats not work")
	ErrMsgHandlerNotExist = errors.New("msg handler not exist")
	ErrMsgNotExist        = errors.New("msg not exist")

	ErrStreamNotExist = errors.New("stream not exist")
)
