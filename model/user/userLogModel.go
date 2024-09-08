package userModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	commonModel "KeepAccount/model/common"
	"gorm.io/gorm"
	"time"
)

type Log struct {
	ID        uint                `gorm:"primarykey"`
	UserId    uint                `gorm:"comment:用户id;not null"`
	Action    constant.UserAction `gorm:"comment:操作;not null;size:32"`
	Remark    string              `gorm:"comment:备注;not null;size:255"`
	CreatedAt time.Time           `gorm:"type:TIMESTAMP"`
	UpdatedAt time.Time           `gorm:"type:TIMESTAMP"`
	DeletedAt gorm.DeletedAt      `gorm:"index;type:TIMESTAMP"`
	commonModel.BaseModel
}

func (l *Log) TableName() string {
	return "user_log"
}

func (l *Log) IsEmpty() bool {
	return l.ID == 0
}

type LogDao struct {
	db *gorm.DB
}

func NewLogDao(db *gorm.DB) *LogDao {
	if db == nil {
		db = global.GvaDb
	}
	return &LogDao{db}
}

type LogAddData struct {
	Action constant.UserAction
	Remark string
}

func (l *LogDao) Add(user User, data *LogAddData) (*Log, error) {
	log := &Log{
		UserId: user.ID,
		Action: data.Action,
		Remark: data.Remark,
	}
	err := l.db.Create(&log).Error
	return log, err
}
