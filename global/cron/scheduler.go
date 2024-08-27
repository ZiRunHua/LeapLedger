package cron

import (
	"KeepAccount/global/nats"
	"errors"
	"go.uber.org/zap"
)

func PublishTask(task nats.Task) func() {
	return MakeJobFunc(
		func() error {
			isSuccess := nats.PublishTask(task)
			if !isSuccess {
				return errors.New("publish fail")
			}
			return nil
		},
	)
}

func PublishTaskWithPayload[T nats.PayloadType](task nats.Task, payload T) func() {
	return MakeJobFunc(
		func() error {
			isSuccess := nats.PublishTaskWithPayload[T](task, payload)
			if !isSuccess {
				return errors.New("publish fail")
			}
			return nil
		},
	)
}

func PublishTaskWithMakePayload[T nats.PayloadType](task nats.Task, makePayload func() (T, error)) func() {
	return MakeJobFunc(
		func() error {
			payload, err := makePayload()
			if err != nil {
				return err
			}
			isSuccess := nats.PublishTaskWithPayload[T](task, payload)
			if !isSuccess {
				return errors.New("publish fail")
			}
			return nil
		},
	)
}

func MakeJobFunc(f func() error) func() {
	return func() {
		defer func() {
			r := recover()
			if r != nil {
				logger.Error("job exec panic", zap.Any("panic", r))
			}
		}()
		err := f()
		if err != nil {
			logger.Error("job exec error", zap.Error(err))
		}
	}
}
