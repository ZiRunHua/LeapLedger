package logModel

import (
	"KeepAccount/global"
	"gorm.io/gorm"
)

type accountLogDao struct {
	db *gorm.DB
}

func (d *dao) NewAccountDao(db *gorm.DB) *accountLogDao {
	if db == nil {
		db = global.GvaDb
	}
	return &accountLogDao{db}
}

func (d *accountLogDao) RecordAccountLogMapping(logModel AccountLogger) (AccountLogMapping, error) {
	log := AccountLogMapping{
		AccountId: logModel.GetAccountId(),
		LogTable:  logModel.TableName(),
		LogId:     logModel.GetId(),
	}
	return log, d.db.Create(&log).Error
}
