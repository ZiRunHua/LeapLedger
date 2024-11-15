package websocket

import (
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
			panic(err)
		}
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
