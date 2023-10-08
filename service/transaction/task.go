package transactionService

import (
	"time"
)

var chanUpdateStatistic = make(chan updateStatisticTask, 100)

func init() {
	go func() {
		for true {
			select {
			case task := <-chanUpdateStatistic:
				if err := task.handleTask(); err != nil {
					panic(err)
				}
			case <-time.After(1 * time.Second):
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

type updateStatisticTask struct {
	data UpdateStatisticData
}

func (ust *updateStatisticTask) handleTask() error {
	return GroupApp.Transaction.updateStatistic(
		&ust.data,
	)
}
