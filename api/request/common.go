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
	StartTime time.Time `binding:"gt=0"`
	EndTime   time.Time `binding:"gt=0"`
}

func (t *TimeFrame) CheckTimeFrame() error {
	if t.EndTime.Before(t.StartTime) {
		return errors.New("时间范围错误")
	}
	if t.StartTime.AddDate(2, 2, 2).After(t.EndTime) {
		return global.ErrTimeFrameIsTooLong
	}
	return nil
}

// 格式化日时间 将时间转为time.Time类型 并将StartTime置为当日第一秒 endTime置为当日最后一秒
func (t *TimeFrame) FormatDayTime() (startTime time.Time, endTime time.Time) {
	startTime = time.Date(t.StartTime.Year(), t.StartTime.Month(), t.StartTime.Day(), 0, 0, 0, 0, time.Local)
	endTime = time.Date(t.EndTime.Year(), t.EndTime.Month(), t.EndTime.Day(), 23, 59, 59, 0, time.Local)
	return
}

// 信息类型
type InfoType string

// 今日交易统计
var TodayTransTotal InfoType = "todayTransTotal"

// 本月交易统计
var CurrentMonthTransTotal InfoType = "currentMonthTransTotal"

// 最近交易数据
var RecentTrans InfoType = "recentTrans"

type AccountId struct {
	AccountId uint
}
