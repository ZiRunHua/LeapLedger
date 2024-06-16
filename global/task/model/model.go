package model

import (
	"KeepAccount/global/constant"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Task struct {
	ID          uint `gorm:"primarykey"`
	Subject     constant.Subject
	Data        string
	Status      TaskStatus
	Error       string
	completedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (t *Task) Complete(db *gorm.DB) error {
	err := db.Model(&RetryTask{}).Delete("task_id = ?", t.ID).Error
	if err != nil {
		return err
	}
	return db.Model(t).Updates(map[string]interface{}{
		"status":       StatusOfComplete,
		"error":        "",
		"completed_at": time.Now(),
	}).Error
}

func (t *Task) Die(db *gorm.DB) error {
	return db.Model(t).Update("status", StatusOfDie).Error
}

func (t *Task) Retry(execTime time.Time, db *gorm.DB) (retryTask RetryTask, err error) {
	err = db.Model(t).Update("status", StatusOfRetrying).Error
	if err != nil {
		return
	}
	retryTask = RetryTask{
		TaskId:       t.ID,
		Count:        0,
		NextExecTime: execTime,
	}
	err = db.Create(&retryTask).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		err = db.First(&retryTask, t.ID).Error
		if err != nil {
			return
		}
		err = db.Model(&retryTask).Updates(map[string]interface{}{
			"count":          0,
			"next_exec_time": execTime,
		}).Error
		if err != nil {
			return
		}
	} else if err != nil {
		return
	}
	return
}

type TaskStatus uint8

const (
	StatusOfNormal   TaskStatus = 0
	StatusOfFailed   TaskStatus = 4
	StatusOfRetrying TaskStatus = 8
	StatusOfDie      TaskStatus = 16
	StatusOfComplete TaskStatus = 32
)

type RetryTask struct {
	TaskId       uint `gorm:"primarykey"`
	Status       RetryStatus
	Count        uint8
	NextExecTime time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type RetryStatus uint8

const (
	RetryStatusOfNormal    RetryStatus = 0
	RetryStatusOfPublished RetryStatus = 8
	RetryStatusOfAbnormal  RetryStatus = 32
)

func (rt *RetryTask) PublishRetry(nextExecTime time.Time, db *gorm.DB) error {
	rt.Count++
	rt.NextExecTime = nextExecTime
	rt.Status = RetryStatusOfPublished
	return db.Model(rt).Select("status", "count", "next_exec_time").Updates(rt).Error
}

func (rt *RetryTask) Abnormal(db *gorm.DB) error {
	return db.Model(rt).Update("status", RetryStatusOfAbnormal).Error
}

func (rt *RetryTask) GetTask(db *gorm.DB) (task Task, err error) {
	err = db.First(&task, rt.TaskId).Error
	return
}
