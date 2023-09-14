package commonModel

import (
	"KeepAccount/global"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func ExistOfModel(model Model, query interface{}, args ...interface{}) (bool, error) {
	err := global.GvaDb.Where(query, args...).Take(model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, errors.Wrap(err, "exist")
	}
}
