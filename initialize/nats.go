package initialize

import (
	"KeepAccount/global/constant"
	"github.com/nats-io/nats.go"
)

type _nats struct {
	ServerUrl        string             `yaml:"ServerUrl"`
	Subjects         []constant.Subject `yaml:"Subjects"`
	subjectMap       map[constant.Subject]struct{}
	IsConsumerServer bool
}

// NatsDb is used to record and retry failure messages
// Enabled in consumer server

const nastStoreDir = constant.RUNTIME_DATA_PATH + "/nats"

func (n *_nats) do() error {
	err := n.init()
	if err != nil {
		return err
	}
	Nats, err = nats.Connect(n.ServerUrl)
	if err != nil {
		return err
	}
	return err
}

func (n *_nats) init() error {
	n.subjectMap = make(map[constant.Subject]struct{})
	for _, subject := range n.Subjects {
		n.subjectMap[subject] = struct{}{}
	}
	n.IsConsumerServer = len(n.subjectMap) > 0
	return nil
}

const allTask constant.Subject = "all"

func (n *_nats) CanSubscribe(subj constant.Subject) bool {
	if _, ok := n.subjectMap[allTask]; ok {
		return true
	}
	_, ok := n.subjectMap[subj]
	return ok
}
