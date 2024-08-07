package globalTask

import (
	"KeepAccount/global/task/model"
	"KeepAccount/initialize"
	"context"
)

type transactionHandle[Data any] func(Data, context.Context) error

func init() {
	natsConn = initialize.Nats
	natsServer = initialize.NatsServer
	natsLogger = initialize.NatsLogger
	cronLogger = initialize.CronLogger
	Scheduler = initialize.Scheduler
	if Scheduler == nil {
		panic("init Scheduler")
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
	_, err := Scheduler.Every(1).Second().Do(NewTransactionCron(cronOfPublishRetryTask))
	if err != nil {
		panic(err)
	}
}

func Shutdown() {
	Scheduler.Stop()
	if natsConn != nil {
		natsConn.Close()
	}
	if natsServer != nil {
		natsServer.Shutdown()
	}
}
