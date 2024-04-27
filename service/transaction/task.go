package transactionService

import (
	"KeepAccount/global/nats"
	transactionModel "KeepAccount/model/transaction"
	"gorm.io/gorm"
)

func init() {
	nats.TransSubscribe[transactionModel.StatisticData](
		nats.TaskStatisticUpdate, func(db *gorm.DB, data transactionModel.StatisticData) error {
			return GroupApp.Transaction.updateStatistic(data, db)
		},
	)
}

type updateStatisticTask struct {
	data transactionModel.StatisticData
}
