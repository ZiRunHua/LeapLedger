package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
)

type CommonApi struct {
}

var captchaStore = base64Captcha.DefaultMemStore

func (p *PublicApi) Captcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(
		global.Config.Captcha.ImgHeight, global.Config.Captcha.ImgWidth, global.Config.Captcha.KeyLong, 0.7,
		80,
	)
	cp := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, err := cp.Generate()
	if err != nil {
		response.FailWithMessage("验证码获取失败", c)
		return
	}
	response.OkWithDetailed(
		response.CommonCaptcha{
			CaptchaId:     id,
			PicBase64:     b64s,
			CaptchaLength: global.Config.Captcha.KeyLong,
		}, "验证码获取成功", c,
	)
}

func (u *PublicApi) SendEmailCaptcha(ctx *gin.Context) {
	var requestData request.CommonSendEmailCaptcha
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	if false == captchaStore.Verify(requestData.CaptchaId, requestData.Captcha, true) {
		response.FailWithMessage("验证码错误", ctx)
		return
	}

	err := thirdpartyService.SendCaptchaEmail(requestData.Email, requestData.Type)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.ExpirationTime{ExpirationTime: global.Config.Captcha.EmailCaptchaTimeOut}, ctx)
}
