package websocket

import (
	"net"
	"time"

	"github.com/ZiRunHua/LeapLedger/api/v1/ws/msg"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Use(handler func(conn *websocket.Conn, ctx *gin.Context) error) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.JSONP(500, "error")
			return
		}
		conn.SetPingHandler(
			func(message string) error {
				err := conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(1))
				if err == websocket.ErrCloseSent {
					return nil
				} else if e, ok := err.(net.Error); ok && e.Temporary() {
					return nil
				}
				return err
			},
		)
		conn.SetPongHandler(nil)
		conn.SetCloseHandler(
			func(code int, text string) error {
				return nil
			},
		)
		defer conn.Close()
		err = handler(conn, ctx)
		if err != nil {
			err = msg.SendError(conn, err)
			if err != nil {
				panic(err)
			}
		}
	}
}
