package accountModel

import (
	"KeepAccount/global"
	logModel "KeepAccount/model/log"
	"time"
)

type Mapping struct {
	ID        uint `gorm:"primarykey"`
	MainId    uint `gorm:"not null;uniqueIndex:idx_mapping,priority:1"`
	RelatedId uint `gorm:"not null;uniqueIndex:idx_mapping,priority:2"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Mapping) TableName() string {
	return "account_mapping"
}

func (m *Mapping) GetMainAccount() (result Account, err error) {
	err = global.GvaDb.First(&result, m.MainId).Error
	return
}

func (m *Mapping) GetRelatedAccount() (result Account, err error) {
	err = global.GvaDb.First(&result, m.RelatedId).Error
	return
}

func (m *Mapping) GetLogDataModel() logModel.AccountLogDataRecordable {
	result := &logModel.AccountMappingLogData{
		MainId:    m.MainId,
		RelatedId: m.RelatedId,
	}
	return result
}
