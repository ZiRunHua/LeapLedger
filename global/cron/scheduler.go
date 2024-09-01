package cron

import (
	"KeepAccount/global/lock"
	"KeepAccount/global/nats"
	"context"
	"errors"
	"go.uber.org/zap"
)

const lockKeyPrefix = "cron:"
const publishLockKeyPrefix = lockKeyPrefix + "publish:"
const taskPublishLockKeyPrefix = publishLockKeyPrefix + "task:"

func PublishTask(task nats.Task) func() {
	return MakeOnceJob(
		lock.Key(taskPublishLockKeyPrefix+string(task)),
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
	return MakeOnceJob(
		lock.Key(taskPublishLockKeyPrefix+string(task)),
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
	return MakeOnceJob(
		lock.Key(taskPublishLockKeyPrefix+string(task)),
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

func MakeOnceJob(key lock.Key, f func() error) func() {
	return MakeJobFunc(
		func() error {
			l := lock.New(key)
			err := l.Lock(context.Background())
			if err != nil {
				if errors.Is(err, lock.ErrLockOccupied) {
					return nil
				}
				return err
			}
			defer l.Release(context.Background())
			return f()
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
