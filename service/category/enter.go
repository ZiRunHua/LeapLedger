package categoryService

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
	_thirdpartyService "github.com/ZiRunHua/LeapLedger/service/thirdparty"
	_log "github.com/ZiRunHua/LeapLedger/util/log"
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
