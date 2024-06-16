package globalTask

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/task/model"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

var taskServer _taskService

type _taskService struct {
}

func (ts *_taskService) addFailedTask(Subject constant.Subject, Data string, execErr error) (task model.Task, err error) {
	task = model.Task{
		Subject: Subject,
		Data:    Data,
		Error:   execErr.Error(),
		Status:  model.StatusOfFailed,
	}
	err = db.Create(&task).Error
	return
}

func (ts *_taskService) addRetryTask(task model.Task) (retryTask model.RetryTask, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		retryTask, err = task.Retry(time.Now().Add(backOff(0)), tx)
		return err
	})
	return
}

var ErrRepublishFailure = errors.New("republish task failed")

func (ts *_taskService) republishTask(retryTask model.RetryTask, db *gorm.DB) error {
	task, err := retryTask.GetTask(db)
	if err != nil {
		return err
	}
	if task.Status != model.StatusOfRetrying {
		return nil
	}
	if false == publishRetryTask(task.Subject, retryTask) {
		return ErrRepublishFailure
	}
	return retryTask.PublishRetry(time.Now().Add(backOff(retryTask.Count+1)), db)
}
