package initialize

type _captcha struct {
	KeyLong             int `yaml:"KeyLong"`
	ImgWidth            int `yaml:"ImgWidth"`
	ImgHeight           int `yaml:"ImgHeight"`
	OpenCaptcha         int `yaml:"OpenCaptcha"`        // 防爆破验证码开启此数，0代表每次登录都需要验证码，其他数字代表错误密码此数，如3代表错误三次后出现验证码
	OpenCaptchaTimeOut  int `yaml:"OpenCaptchaTimeOut"` // 防爆破验证码超时时间，单位：s(秒)
	EmailCaptcha        int `yaml:"EmailCaptcha"`
	EmailCaptchaTimeOut int `yaml:"EmailCaptchaTimeOut"`
}
