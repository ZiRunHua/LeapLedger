package initialize

type _captcha struct {
	KeyLong            int
	ImgWidth           int
	ImgHeight          int
	OpenCaptcha        int // 防爆破验证码开启此数，0代表每次登录都需要验证码，其他数字代表错误密码此数，如3代表错误三次后出现验证码
	OpenCaptchaTimeOut int // 防爆破验证码超时时间，单位：s(秒)
}
