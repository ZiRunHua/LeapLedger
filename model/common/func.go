package commonModel

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func SelectByIdOfModel(model Model, id uint, forUpdate bool) error {
	if forUpdate {
		return model.GetDb().Set("gorm:query_option", "FOR UPDATE").First(model, id).Error
	}
	return model.GetDb().First(model, id).Error
}

func ExistOfModel(model Model, query interface{}, args ...interface{}) (bool, error) {
	err := model.GetDb().Where(query, args...).Take(model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, errors.Wrap(err, "exist")
	}
}
