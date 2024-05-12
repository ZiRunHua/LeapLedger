package initialize

import (
	"KeepAccount/global/constant"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type _nats struct {
	ServerUrl string `yaml:"ServerUrl"`
}

var Nats *nats.Conn

type NastConn[T struct{}] struct {
	nats *nats.Conn
}

func (n *_nats) do(mode constant.ServerMode) error {
	if mode == constant.Debug {
		opts := &server.Options{
			Debug:     true,
			JetStream: true,
			Trace:     true,
			Logtime:   true,
			LogFile:   _natsLogPath}
		nastServer, err := server.NewServer(opts)
		if err != nil {
			return err
		}
		nastServer.Start()
		n.ServerUrl = nats.DefaultURL
	}
	var err error
	Nats, err = nats.Connect(n.ServerUrl)
	if err != nil {
		return err
	}
	return err
}
