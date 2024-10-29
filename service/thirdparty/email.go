package thirdpartyService

import (
	"bytes"
	"fmt"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	userModel "github.com/ZiRunHua/LeapLedger/model/user"
	commonService "github.com/ZiRunHua/LeapLedger/service/common"
	"github.com/ZiRunHua/LeapLedger/util/rand"
	"github.com/pkg/errors"
	"os"
	"time"
)

var emailTemplate map[constant.Notification][]byte
var emailTemplateFilePath = map[constant.Notification]string{
	constant.NotificationOfCaptcha:             "/template/email/captcha.html",
	constant.NotificationOfRegistrationSuccess: "/template/email/registerSuccess.html",
	constant.NotificationOfUpdatePassword:      "/template/email/updatePassword.html",
}

func init() {
	emailTemplate = make(map[constant.Notification][]byte, len(emailTemplateFilePath))
	var err error
	for notification, path := range emailTemplateFilePath {
		if emailTemplate[notification], err = os.ReadFile(constant.DATA_PATH + path); err != nil {
			panic(err)
		}
	}
}

func (g *Group) sendCaptchaEmail(email string, action constant.UserAction) error {
	captcha := rand.String(6)
	expirationTime := time.Second * time.Duration(global.Config.Captcha.EmailCaptchaTimeOut)
	err := commonService.Common.SetEmailCaptchaCache(email, captcha, expirationTime)
	if err != nil {
		return err
	}
	minutes := int(expirationTime.Minutes())
	content := bytes.Replace(emailTemplate[constant.NotificationOfCaptcha], []byte("[Captcha]"), []byte(captcha), 1)
	content = bytes.Replace(content, []byte("[ExpirationTime]"), []byte(fmt.Sprintf("%d分钟", minutes)), 1)
	var actionName string
	switch action {
	case constant.Register:
		actionName = "注册"
	case constant.ForgetPassword:
		actionName = "忘记密码"
	case constant.UpdatePassword:
		actionName = "修改密码"
	default:
		return errors.Wrap(global.ErrUnsupportedUserAction, "发送邮箱验证码")
	}
	content = bytes.Replace(content, []byte("[Action]"), []byte(actionName), 2)
	return emailServer.Send([]string{email}, actionName+"验证码", string(content))
}

func (g *Group) sendRegisterSuccessEmail(user userModel.User) error {
	content := bytes.Replace(
		emailTemplate[constant.NotificationOfRegistrationSuccess], []byte("[username]"), []byte(user.Username), 1,
	)
	return emailServer.Send([]string{user.Email}, "注册成功", string(content))
}

func (g *Group) sendUpdatePasswordEmail(user userModel.User) error {
	content := bytes.Replace(
		emailTemplate[constant.NotificationOfUpdatePassword], []byte("[username]"), []byte(user.Username), 1,
	)
	return emailServer.Send([]string{user.Email}, "修改密码", string(content))
}

func (g *Group) sendNotificationEmail(user userModel.User, notification constant.Notification) error {
	switch notification {
	case constant.NotificationOfUpdatePassword:
		return g.sendUpdatePasswordEmail(user)
	case constant.NotificationOfRegistrationSuccess:
		return g.sendRegisterSuccessEmail(user)
	default:
		return errors.New("不支持该类型邮箱通知")
	}
}
