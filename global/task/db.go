package globalTask

import (
	"KeepAccount/global"
	"KeepAccount/global/task/model"
)

var db = global.GvaDb

func init() {
	if global.Config.Nats.IsConsumerServer {
		err := db.AutoMigrate(model.Task{}, model.RetryTask{})
		if err != nil {
			panic(err)
		}
	}
}
