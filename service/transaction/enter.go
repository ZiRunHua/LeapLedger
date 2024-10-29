package transactionService

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
	_log "github.com/ZiRunHua/LeapLedger/util/log"
	"go.uber.org/zap"
)

type Group struct {
	Transaction
	Timing Timing
}

var (
	GroupApp = new(Group)

	errorLog *zap.Logger
	task     = &_task{}
	server   = &Transaction{}
)

// 初始化
func init() {
	var err error
	if errorLog, err = _log.GetNewZapLogger(constant.LOG_PATH + "/service/transaction/error.log"); err != nil {
		panic(err)
	}
}
