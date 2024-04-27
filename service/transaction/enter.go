package transactionService

import (
	"KeepAccount/global/constant"
	_log "KeepAccount/util/log"
	"go.uber.org/zap"
)

type Group struct {
	Transaction
}

var GroupApp = new(Group)
var errorLog *zap.Logger

// 初始化
func init() {
	var err error
	if errorLog, err = _log.GetNewZapLogger(constant.LOG_PAYH + "/service/transaction/error.log"); err != nil {
		panic(err)
	}
}
