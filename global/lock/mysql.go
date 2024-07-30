package lock

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var (
	mdb *gorm.DB
)

type lockTable struct {
	Key        string `gorm:"primarykey"`
	Value      string
	Expiration int64
}

func (lt *lockTable) TableName() string { return "lock" }

func newMysqlLock(client *gorm.DB, key string, expiration time.Duration) *MysqlLock {
	return &MysqlLock{
		client:     client,
		key:        key,
		value:      uuid.New().String(),
		expiration: expiration,
	}
}

type MysqlLock struct {
	client     *gorm.DB
	key        string
	value      string
	expiration time.Duration
}

func (ml *MysqlLock) Lock(ctx context.Context) error {
	err := ml.client.WithContext(ctx).Create(&lockTable{
		Key:        ml.key,
		Value:      ml.value,
		Expiration: time.Now().Add(ml.expiration).Unix(),
	}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrLockOccupied
		}
		return err
	}
	return nil
}

func (ml *MysqlLock) Release(ctx context.Context) error {
	result := ml.client.WithContext(ctx).Where("`key` = ? AND `value` = ?", ml.key, ml.value).Delete(&lockTable{})
	err := result.Error
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return ErrLockNotExist
	}
	return nil
}
