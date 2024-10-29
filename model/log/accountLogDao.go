package logModel

import (
	"github.com/ZiRunHua/LeapLedger/global"
	"gorm.io/gorm"
)

type AccountLogDao struct {
	db *gorm.DB
}

func NewDao(db ...*gorm.DB) *AccountLogDao {
	if len(db) > 0 {
		return &AccountLogDao{db: db[0]}
	}
	return &AccountLogDao{global.GvaDb}
}

func (d *AccountLogDao) RecordAccountLogMapping(logModel AccountLogger) (AccountLogMapping, error) {
	log := AccountLogMapping{
		AccountId: logModel.GetAccountId(),
		LogTable:  logModel.TableName(),
		LogId:     logModel.GetId(),
	}
	return log, d.db.Create(&log).Error
}
