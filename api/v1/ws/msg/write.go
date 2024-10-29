package msg

import (
	"github.com/gorilla/websocket"
)

type Msg struct {
	Type MsgType
	Data any
}

func Send[T any](conn *websocket.Conn, t MsgType, data T) error {
	return conn.WriteJSON(Msg{Type: t, Data: data})
}

const msgTypeErr MsgType = "error"

func SendError(conn *websocket.Conn, err error) error {
	return conn.WriteJSON(Msg{Type: msgTypeErr, Data: err.Error()})
}
