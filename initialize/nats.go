package initialize

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type _nats struct {
}

var Nast *nats.Conn

type NastConn[T struct{}] struct {
	nats *nats.Conn
}

func (m *_nats) do() error {
	opts := &server.Options{
		Port: 4222,
	}
	nastServer, err := server.NewServer(opts)
	nastServer.Start()
	Nast, err = nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}
	return nil
}
