package logModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/util/gormFunc"
	"gorm.io/gorm"
)

func CurrentInit() error {
	tables := []interface{}{
		AccountMappingLog{}, AccountLogMapping{},
	}
	err := global.GvaDb.AutoMigrate(tables...)
	if err != nil {
		return err
	}
	for _, table := range tables {
		_ = gormFunc.AlterIdToHeader(table, global.GvaDb)
	}
	return nil
}

/*账本*/
type AccountLog[T AccountLogDataRecordable] struct {
	BaseAccountLog `gorm:"embedded"`
	Data           T `gorm:"type:json"`
}

type BaseAccountLog struct {
	ID        uint                  `gorm:"primarykey"`
	UserId    uint                  `gorm:"index:idx_user_id;not null"`
	AccountId uint                  `gorm:"index:idx_account_id;not null"`
	Operation constant.LogOperation `gorm:"not null"`
}

func (b *BaseAccountLog) GetId() uint {
	return b.ID
}
func (b *BaseAccountLog) GetAccountId() uint {
	return b.AccountId
}

type AccountLogger interface {
	TableName() string
	GetId() uint
	GetAccountId() uint
	RecordMapping(tx *gorm.DB) (AccountLogMapping, error)
}

type AccountLogDataRecordable interface {
	Record(baseLog BaseAccountLog, tx *gorm.DB) (AccountLogger, error)
}

type AccountLogDataProvider interface {
	GetLogDataModel() AccountLogDataRecordable
}
