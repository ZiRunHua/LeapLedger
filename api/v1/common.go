package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"net/http"
	"time"
)

type CommonApi struct {
}

var store = base64Captcha.DefaultMemStore

func (p *PublicApi) Captcha(c *gin.Context) {
	// 判断验证码是否开启
	openCaptcha := global.GvaConfig.Captcha.OpenCaptcha               // 是否开启防爆次数
	openCaptchaTimeOut := global.GvaConfig.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	key := c.ClientIP()
	v, ok := global.BlackCache.Get(key)
	if !ok {
		global.BlackCache.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}

	var oc bool
	if openCaptcha == 0 || openCaptcha < interfaceToInt(v) {
		oc = true
	}
	// 字符,公式,验证码配置
	// 生成默认数字的driver
	driver := base64Captcha.NewDriverDigit(
		global.GvaConfig.Captcha.ImgHeight, global.GvaConfig.Captcha.ImgWidth, global.GvaConfig.Captcha.KeyLong, 0.7,
		80,
	)
	cp := base64Captcha.NewCaptcha(driver, store)
	id, b64s, err := cp.Generate()
	if err != nil {
		response.FailWithMessage("验证码获取失败", c)
		return
	}
	response.OkWithDetailed(
		response.CommonCaptcha{
			CaptchaId:     id,
			PicPath:       b64s,
			CaptchaLength: global.GvaConfig.Captcha.KeyLong,
			OpenCaptcha:   oc,
		}, "验证码获取成功", c,
	)
}

// 类型转换
func interfaceToInt(v interface{}) (i int) {
	switch v := v.(type) {
	case int:
		i = v
	default:
		i = 0
	}
	return
}

func (p *PublicApi) Login(ctx *gin.Context) {
	var request request.UserLogin
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key := ctx.ClientIP()
	//oc := commonService.CheckCaptchaStatus(key)
	//if oc || store.Verify(request.CaptchaId, request.Captcha, true) {
	//	response.FailWithMessage("验证码错误", ctx)
	//	return
	//}
	client := current.GetClient(ctx)
	currentAccount, token, err := userService.Login(request.Username, request.Password, client)
	if err != nil {
		// 验证码次数+1
		global.BlackCache.Increment(key, 1)
		response.FailWithMessage("用户名不存在或者密码错误", ctx)
		return
	}
	if token == "" {
		response.FailWithMessage("token获取失败", ctx)
	}
	response.OkWithDetailed(
		response.Login{
			Token: token, CurrentAccount: response.AccountModelToResponse(currentAccount),
		}, "登录成功", ctx,
	)
}
