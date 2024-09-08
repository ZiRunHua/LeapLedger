package accountModel

import (
	logModel "KeepAccount/model/log"
	"gorm.io/gorm"
	"time"
)

type Mapping struct {
	ID        uint      `gorm:"primarykey"`
	MainId    uint      `gorm:"not null;uniqueIndex:idx_mapping,priority:1"`
	RelatedId uint      `gorm:"not null;uniqueIndex:idx_mapping,priority:2"`
	CreatedAt time.Time `gorm:"type:TIMESTAMP"`
	UpdatedAt time.Time `gorm:"type:TIMESTAMP"`
}

func (m *Mapping) TableName() string {
	return "account_mapping"
}

func (m *Mapping) GetMainAccount(db *gorm.DB) (result Account, err error) {
	err = db.First(&result, m.MainId).Error
	return
}

func (m *Mapping) GetRelatedAccount(db *gorm.DB) (result Account, err error) {
	err = db.First(&result, m.RelatedId).Error
	return
}

func (m *Mapping) GetLogDataModel() logModel.AccountLogDataRecordable {
	result := &logModel.AccountMappingLogData{
		MainId:    m.MainId,
		RelatedId: m.RelatedId,
	}
	return result
}

type MappingCondition struct {
	mainId    *uint
	relatedId *uint
}

func NewMappingCondition() *MappingCondition {
	return &MappingCondition{}
}

func (mc *MappingCondition) addConditionToQuery(db *gorm.DB) *gorm.DB {
	if mc.mainId != nil {
		db = db.Where("main_id = ?", mc.mainId)
	}
	if mc.relatedId != nil {
		db = db.Where("related_id = ?", mc.relatedId)
	}
	return db
}

func (mc *MappingCondition) WithMainId(mainId uint) *MappingCondition {
	mc.mainId = &mainId
	return mc
}

func (mc *MappingCondition) WithRelatedId(relatedId uint) *MappingCondition {
	mc.relatedId = &relatedId
	return mc
}
