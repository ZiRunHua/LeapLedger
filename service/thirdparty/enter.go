package thirdpartyService

import (
	"KeepAccount/global/constant"
	"KeepAccount/service/thirdparty/email"
	_log "KeepAccount/util/log"
	"go.uber.org/zap"
)

type Group struct {
	Ai aiServer
}

var (
	GroupApp    = new(Group)
	log         *zap.Logger
	emailServer = email.Service
)

// 初始化
func init() {
	var err error
	if log, err = _log.GetNewZapLogger(constant.LOG_PATH + "/service/thirdparty/email.log"); err != nil {
		panic(err)
	}
}
