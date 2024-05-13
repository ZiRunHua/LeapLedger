package initialize

import (
	"KeepAccount/global/constant"
	_natsLogger "github.com/nats-io/nats-server/v2/logger"
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
			JetStream: true,
			Trace:     true,
			Logtime:   true,
			LogFile:   _natsLogPath}
		nastServer, err := server.NewServer(opts)
		if err != nil {
			return err
		}
		nastServer.SetLoggerV2(_natsLogger.NewFileLogger(_natsLogPath, true, false, true, true, _natsLogger.LogUTC(false)), false, true, false)
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
