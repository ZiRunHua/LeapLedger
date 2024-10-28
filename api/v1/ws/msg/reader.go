package msg

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"KeepAccount/util/fileTool"
	"github.com/gorilla/websocket"
)

func NewReader() Reader {
	return &reader{MsgHandle: make(map[MsgType]MsgHandler)}
}

func RegisterHandle[T any](reader Reader, msgType MsgType, handler func(data T) error) {
	reader.registerHandle(
		msgType, func(bytes []byte) error {
			var data T
			err := json.Unmarshal(bytes, &data)
			if err != nil {
				return err
			}
			return handler(data)
		},
	)
}

func ForReadAndHandleJsonMsg(reader Reader, conn *websocket.Conn) error {
	var err error
	for {
		err = reader.readJsonMsgToHandle(conn)
		if err != nil {
			return err
		}
	}
}

func ReadBytes(reader Reader, conn *websocket.Conn) ([]byte, error) {
	return reader.readBytes(conn)
}

func ReadFile(reader Reader, conn *websocket.Conn) (io.Reader, error) {
	return reader.readFile(conn)
}

type readMsg struct {
	Type MsgType
	Data json.RawMessage
}

type Reader interface {
	registerHandle(msgType MsgType, handler MsgHandler)
	readJsonMsgToHandle(conn *websocket.Conn) error
	readBytes(conn *websocket.Conn) ([]byte, error)
	readFile(conn *websocket.Conn) (io.Reader, error)
}

type reader struct {
	Reader
	MsgHandle map[MsgType]MsgHandler
	lock      sync.Mutex
}

func (mr *reader) registerHandle(msgType MsgType, handler MsgHandler) {
	mr.lock.Lock()
	defer mr.lock.Unlock()
	mr.MsgHandle[msgType] = handler
}

func (mr *reader) readJsonMsgToHandle(conn *websocket.Conn) error {
	var msg readMsg
	_, str, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	err = json.Unmarshal(str, &msg)
	if err != nil {
		return err
	}
	handle, exist := mr.MsgHandle[msg.Type]
	if !exist {
		return ErrMsgHandleNotExist
	}
	return handle(msg.Data)
}

func (mr *reader) readBytes(conn *websocket.Conn) ([]byte, error) {
	_, str, err := conn.ReadMessage()
	return str, err
}
func (mr *reader) readFile(conn *websocket.Conn) (io.Reader, error) {
	_, fileData, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(fileData)
	return fileTool.GetUTF8Reader(r, fileData)
}
