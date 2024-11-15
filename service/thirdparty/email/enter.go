package email

import (
	"fmt"

	"github.com/ZiRunHua/LeapLedger/global/cron"
)

type emailService interface {
	init()
	Send(emails []string, subject string, contest string) error
	getToken() error
}

// 初始化
var Service emailService
var ServiceStatus bool = false

func init() {
	Service = &WeCom{}
	Service.init()
}

// token过期
var tokenExpiredError = &_tokenExpiredError{}

type _tokenExpiredError struct {
}

func (e *_tokenExpiredError) Error() string {
	return "邮箱服务Token已过期"
}

// 第三方响应错误
type thirdPartyResponseError struct {
	StatusCode int    // 第三方响应的HTTP状态码
	ErrorCode  int    // 第三方响应的错误码
	Message    string // 错误消息
}

func (e *thirdPartyResponseError) Error() string {
	return fmt.Sprintf("第三方响应错误，状态码：%d，错误码：%d，消息：%s", e.StatusCode, e.ErrorCode, e.Message)
}

func init() {
	_, err := cron.Scheduler.Every(30).Minute().Do(
		cron.MakeJobFunc(
			func() error {
				if false == ServiceStatus {
					return nil
				}
				return Service.getToken()
			},
		),
	)
	if err != nil {
		panic(err)
	}
}
