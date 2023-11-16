package thirdpartyService

import (
	"KeepAccount/global/constant"
	"KeepAccount/service/thirdparty/email"
	"KeepAccount/util"
	"fmt"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type Group struct {
}

var App = new(Group)
var log *zap.Logger

// 初始化
func init() {
	var err error
	if log, err = util.Log.GetNewZapLogger(constant.LOG_PAYH + "/service/thirdparty/email.log"); err != nil {
		panic(err)
	}
	go startService()
}

func startService() {
	for {
		select {
		case tas := <-emailTaskChannel:
			if err := email.Service.Send(tas.Emails, tas.Subject, tas.Content); err != nil {
				tas.retry(err)
			}
		case <-time.After(time.Second * 5): // 每秒钟检查一次通道
			time.Sleep(time.Second * 5)
		}
	}
}

type tasker interface {
	canRetry() bool
	retry(error)
	handleError(err error)
}

type task struct {
	tasker
	retryCount int
	createdAt  time.Time
	err        error
}

func (t *task) handleError(err error) {
	if err != nil {
		fmt.Println(err)
		log.Error(reflect.TypeOf(*t).Name(), zap.Error(err))
	}
	t.err = err
}

func (t *task) canRetry() bool {
	if t.retryCount >= 3 {
		return false
	}
	t.retryCount++
	return true
}
