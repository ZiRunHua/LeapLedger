package request

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"github.com/pkg/errors"
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
