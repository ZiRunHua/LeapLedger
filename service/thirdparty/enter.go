package thirdpartyService

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
	"github.com/ZiRunHua/LeapLedger/service/thirdparty/email"
	_log "github.com/ZiRunHua/LeapLedger/util/log"
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
