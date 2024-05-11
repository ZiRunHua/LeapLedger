package initialize

import (
	"github.com/nats-io/nats.go"
)

type _nats struct {
	ServerUrl string
}

var Nats *nats.Conn

type NastConn[T struct{}] struct {
	nats *nats.Conn
}

func (n *_nats) do() error {
	var err error
	Nats, err = nats.Connect(n.ServerUrl)
	if err != nil {
		return err
	}
	return nil
}
