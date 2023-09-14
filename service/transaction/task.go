package transactionService

import (
	"KeepAccount/global/nats"
	"gorm.io/gorm"
)

func init() {
	nats.TransSubscribe[UpdateStatisticData](
		nats.TaskStatisticUpdate, func(db *gorm.DB, data UpdateStatisticData) error {
			return GroupApp.Transaction.updateStatistic(data, db)
		},
	)
}

type updateStatisticTask struct {
	data UpdateStatisticData
}
