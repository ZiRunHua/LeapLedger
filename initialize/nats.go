package initialize

import (
	"KeepAccount/global/constant"
	_natsLogger "github.com/nats-io/nats-server/v2/logger"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type _nats struct {
	ServerUrl  string             `yaml:"ServerUrl"`
	Subjects   []constant.Subject `yaml:"Subjects"`
	subjectMap map[constant.Subject]struct{}
}

var Nats *nats.Conn

type NastConn[T struct{}] struct {
	nats *nats.Conn
}

func (n *_nats) do(mode constant.ServerMode) error {
	n.init()
	if mode == constant.Debug {
		opts := &server.Options{
			JetStream: true,
			Trace:     true,
			Logtime:   true,
			LogFile:   _natsServerLogPath}
		nastServer, err := server.NewServer(opts)
		if err != nil {
			return err
		}
		nastServer.SetLoggerV2(_natsLogger.NewFileLogger(_natsServerLogPath, true, false, true, true, _natsLogger.LogUTC(false)), false, true, false)
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

func (n *_nats) init() {
	n.subjectMap = make(map[constant.Subject]struct{})
	for _, subject := range n.Subjects {
		n.subjectMap[subject] = struct{}{}
	}
}

func (n *_nats) CanSubscribe(subj constant.Subject) bool {
	_, ok := n.subjectMap[subj]
	return ok
}
