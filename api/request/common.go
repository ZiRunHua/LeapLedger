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
	StartTime time.Time
	EndTime   time.Time
}

func (t *TimeFrame) CheckTimeFrame() error {
	if t.EndTime.Before(t.StartTime) {
		return errors.New("时间范围错误")
	}
	if t.StartTime.AddDate(2, 2, 2).Before(t.EndTime) {
		return global.ErrTimeFrameIsTooLong
	}
	return nil
}

// 格式化日时间 将StartTime置为当日第一秒 endTime置为当日最后一秒
func (t *TimeFrame) FormatDayTime() (startTime time.Time, endTime time.Time) {
	year, month, day := t.StartTime.Date()
	startTime = time.Date(year, month, day, 0, 0, 0, 0, t.StartTime.Location())
	year, month, day = t.EndTime.Date()
	endTime = time.Date(year, month, day, 23, 59, 59, 0, t.EndTime.Location())
	return
}

func (t *TimeFrame) SetLocal(l *time.Location) {
	t.StartTime, t.EndTime = t.StartTime.In(l), t.EndTime.In(l)
}

func (t *TimeFrame) ToUTC() {
	t.StartTime, t.EndTime = t.StartTime.UTC(), t.EndTime.UTC()
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
