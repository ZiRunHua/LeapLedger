package globalTask

import (
	"KeepAccount/global/task/model"
	"KeepAccount/initialize"
	"context"
)

type transactionHandle[Data any] func(Data, context.Context) error

func init() {
	natsConn = initialize.Nats
	natsLogger = initialize.NatsLogger
	cronLogger = initialize.CronLogger
	scheduler = initialize.Scheduler
	if scheduler == nil {
		panic("init scheduler")
	}
	initDb()
	initCron()
}
func initDb() {
	db = initialize.Db
	if initialize.Config.Nats.IsConsumerServer {
		err := db.AutoMigrate(model.Task{}, model.RetryTask{})
		if err != nil {
			panic(err)
		}
	}
}

func initCron() {
	_, err := scheduler.Every(1).Second().Do(NewTransactionCron(cronOfPublishRetryTask))
	if err != nil {
		panic(err)
	}
}
