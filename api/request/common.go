package request

import (
	"KeepAccount/global/constant"
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
