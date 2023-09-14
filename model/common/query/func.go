package query

import (
	"KeepAccount/global"
	commonModel "KeepAccount/model/common"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func FirstByPrimaryKey[T commonModel.Model](id interface{}) (T, error) {
	var result T
	err := global.GvaDb.First(&result, id).Error
	return result, err
}

func FirstByField[T commonModel.Model](field string, value interface{}) (T, error) {
	var result T
	err := global.GvaDb.Where(map[string]interface{}{field: value}).First(&result).Error
	return result, err
}

func Exist[T commonModel.Model](query interface{}, args ...interface{}) (bool, error) {
	var result T
	err := result.GetDb().Where(query, args...).Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, errors.Wrap(err, "exist")
	}
}
