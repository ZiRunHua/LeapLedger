package request

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"github.com/pkg/errors"
	"time"
)

type IncomeExpense struct {
	IncomeExpense constant.IncomeExpense `json:"Income_expense"`
}
type Name struct {
	Name string
}
type Id struct {
	Id uint
}

type PageData struct {
	Offset int `binding:"gte=0"`
	Limit  int `binding:"gt=0"`
}

type PicCaptcha struct {
	Captcha   string `binding:"required"`
	CaptchaId string `binding:"required"`
}

type CommonSendEmailCaptcha struct {
	Email string              `binding:"required,email"`
	Type  constant.UserAction `binding:"required,oneof=register forgetPassword"`
	PicCaptcha
}

type TimeFrame struct {
	StartTime int64 `binding:"omitempty,gt=0"`
	EndTime   int64 `binding:"omitempty,gt=0"`
}

func (t *TimeFrame) CheckTimeFrame() error {
	if t.StartTime == 0 || t.EndTime == 0 || t.EndTime < t.StartTime {
		return errors.New("时间范围错误")
	}
	if t.EndTime-t.StartTime >= 63244800 {
		return global.ErrTimeFrameIsTooLong
	}
	return nil
}

// 格式化日时间 将时间转为time.Time类型 并将StartTime置为当日第一秒 endTime置为当日最后一秒
func (t *TimeFrame) FormatDayTime() (startTime time.Time, endTime time.Time) {
	startTime = time.Unix(t.StartTime, 0)
	endTime = time.Unix(t.EndTime, 0)
	startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.Local)
	endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, time.Local)
	return
}
