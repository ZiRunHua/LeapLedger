package logService

import (
	logModel "KeepAccount/model/log"
	"gorm.io/gorm"
)

type log struct{}

var Log = new(log)

type _log interface {
	RecordAccountLog(provider logModel.AccountLogDataProvider, baseInfo logModel.BaseAccountLog, tx *gorm.DB) (
		AccountLog logModel.AccountLogger, logMapping logModel.AccountLogMapping, err error,
	)
}

func (logSvc *log) RecordAccountLog(
	provider logModel.AccountLogDataProvider,
	baseInfo logModel.BaseAccountLog,
	tx *gorm.DB,
) (AccountLog logModel.AccountLogger, logMapping logModel.AccountLogMapping, err error) {
	data := provider.GetLogDataModel()
	AccountLog, err = data.Record(baseInfo, tx)
	if err != nil {
		return
	}
	logMapping, err = AccountLog.RecordMapping(tx)
	return
}
