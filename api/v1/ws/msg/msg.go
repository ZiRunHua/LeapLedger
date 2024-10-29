// Package msg is the middle layer of client websocket communication
// This package is used to standardize websocket message formats with clients
package msg

import "errors"

type MsgType string
type MsgHandler func([]byte) error

var ErrMsgHandleNotExist = errors.New("msg handel not exist")
