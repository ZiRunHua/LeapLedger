package query

import (
	"github.com/ZiRunHua/LeapLedger/global"
	commonModel "github.com/ZiRunHua/LeapLedger/model/common"

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
	err := global.GvaDb.Where(query, args...).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, errors.Wrap(err, "exist")
	}
}

func GetAmountCountStatistic(query *gorm.DB) (result global.AmountCount, err error) {
	err = query.Select("COUNT(*) as Count,SUM(amount) as Amount").Scan(&result).Error
	return
}
