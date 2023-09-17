package request

type UserLogin struct {
	Username  string `form:"username"`
	Password  string `form:"password"`
	Captcha   string
	CaptchaId string
}
