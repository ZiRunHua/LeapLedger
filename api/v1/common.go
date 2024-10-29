package v1

import (
	"github.com/ZiRunHua/LeapLedger/api/request"
	"github.com/ZiRunHua/LeapLedger/api/response"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/nats"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"github.com/pkg/errors"
)

type CommonApi struct {
}

var captchaStore = base64Captcha.DefaultMemStore

// Captcha
//
//	@Tags		Common
//	@Produce	json
//	@Success	200	{object}	response.Data{Data=response.CommonCaptcha}
//	@Router		/public/captcha [get]
func (p *PublicApi) Captcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(
		global.Config.Captcha.ImgHeight, global.Config.Captcha.ImgWidth, global.Config.Captcha.KeyLong, 0.7,
		80,
	)
	cp := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, _, err := cp.Generate()
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

// SendEmailCaptcha
//
//	@Tags		Common
//	@Accept		json
//	@Produce	json
//	@Param		body	body		request.CommonSendEmailCaptcha	true	"data"
//	@Success	200		{object}	response.Data{Data=response.ExpirationTime}
//	@Router		/public/captcha/email/send [post]
func (p *PublicApi) SendEmailCaptcha(ctx *gin.Context) {
	var requestData request.CommonSendEmailCaptcha
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	if false == captchaStore.Verify(requestData.CaptchaId, requestData.Captcha, true) {
		response.FailWithMessage("验证码错误", ctx)
		return
	}

	isSuccess := nats.PublishTaskWithPayload(
		nats.TaskSendCaptchaEmail, nats.PayloadSendCaptchaEmail{
			Email: requestData.Email, Action: requestData.Type,
		},
	)
	if !isSuccess {
		response.FailToError(ctx, errors.New("发送失败"))
	}
	response.OkWithData(response.ExpirationTime{ExpirationTime: global.Config.Captcha.EmailCaptchaTimeOut}, ctx)
}
