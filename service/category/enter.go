package categoryService

import (
	"KeepAccount/global/constant"
	_thirdpartyService "KeepAccount/service/thirdparty"
	_log "KeepAccount/util/log"
	"go.uber.org/zap"
)

type Group struct {
	Category
	Task _task
}

var GroupApp = new(Group)

var task = &_task{}
var aiService = _thirdpartyService.GroupApp.Ai
var errorLog *zap.Logger

// 初始化
func init() {
	initLog()
}
func initLog() {
	var err error
	if errorLog, err = _log.GetNewZapLogger(constant.LOG_PATH + "/service/category/error.log"); err != nil {
		panic(err)
	}
}
