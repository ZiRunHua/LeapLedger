package request

import "KeepAccount/global/constant"

type UserLogin struct {
	Email    string `binding:"required"`
	Password string `binding:"required"`
	PicCaptcha
}

type UserRegister struct {
	Username string `binding:"required"`
	Password string `binding:"required"`
	Email    string `binding:"required,email"`
	Captcha  string `binding:"required"`
}

type UserForgetPassword struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required"`
	Captcha  string `binding:"required"`
}

type UserUpdatePassword struct {
	Password string `binding:"required"`
	Captcha  string `binding:"required"`
}

type UserUpdateInfo struct {
	Username string `binding:"required"`
}

type UserSendEmail struct {
	PicCaptcha
	Type constant.UserAction `binding:"required,oneof=updatePassword"`
}
