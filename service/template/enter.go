package templateService

import (
	"KeepAccount/global/constant"
	userModel "KeepAccount/model/user"
	_categoryService "KeepAccount/service/category"
	_log "KeepAccount/util/log"
	"go.uber.org/zap"
)

type Group struct {
	template
}

var (
	GroupApp = &Group{}

	errorLog *zap.Logger

	TmplUserId uint = 1
)

func init() {
	var err error
	if errorLog, err = _log.GetNewZapLogger(constant.LOG_PATH + "/service/template/error.log"); err != nil {
		panic(err)
	}
	initRank()
}

const (
	TmplUserEmail    = "template@gmail.com"
	TmplUserPassword = "1999123456"
	TmplUserName     = "template"
)

func SetTmplUser(user userModel.User) {
	TmplUserId = user.ID
}

var categoryService = _categoryService.GroupApp
