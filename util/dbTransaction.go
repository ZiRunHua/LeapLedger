package util

import (
	commonModel "KeepAccount/model/common"
	"gorm.io/gorm"
)

func SetTxOfModels(tx *gorm.DB, models ...commonModel.Model) {
	for _, model := range models {
		if model == nil {
			continue
		}
		model.SetTx(tx)
	}
}
