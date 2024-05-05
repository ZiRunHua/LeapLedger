package categoryService

import (
	"KeepAccount/global/constant"
	_thirdpartyService "KeepAccount/service/thirdparty"
	_log "KeepAccount/util/log"
	"go.uber.org/zap"
)

type Group struct {
	Category
}

var GroupApp = new(Group)

var aiService = _thirdpartyService.GroupApp.Ai
var errorLog *zap.Logger

// 初始化
func init() {
	var err error
	if errorLog, err = _log.GetNewZapLogger(constant.LOG_PAYH + "/service/category/error.log"); err != nil {
		panic(err)
	}
}
