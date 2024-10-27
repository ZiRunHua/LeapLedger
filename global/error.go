package global

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrNoPermission     = errors.New("无权限")

	ErrFrequentOperation  = errors.New("操作太频繁，稍后再试")
	ErrDeviceNotSupported = errors.New("当前设备不支持")

	ErrTooManyTourists = errors.New("游客过多稍后再试")
)

// 数据校验
var (
	ErrTimeFrameIsTooLong = errors.New("时间范围过长")

	ErrAccountId = errors.New("error accountId")
)

type errDataIsEmpty struct {
	Field string
}

func (e *errDataIsEmpty) Error() string {
	return e.Field + "数据不可为空"
}

func NewErrDataIsEmpty(param string) error {
	return &errDataIsEmpty{
		Field: param,
	}
}

var ErrOperationTooFrequent = errors.New("操作过于频繁,请稍后再试！")
var ErrVerifyEmailCaptchaFail = errors.New("校验邮箱验证码失败！")
var ErrServiceClosed = errors.New("服务未开启！")
var ErrTouristHaveNoRight = errors.New("游客无权操作！")

// 对应constant.UserAction
var ErrUnsupportedUserAction = errors.New("暂不支持该操作")

// 用户
var ErrSameAsTheOldPassword = errors.New("新旧密码相同")

// 账本
var ErrAccountType = errors.New("账本类型不允许该操作")

// 交易类型
var ErrCategoryNameEmpty = errors.New("名称不可为空")
var ErrCategorySameName = errors.New("类型名称相同")

func NewErrThirdpartyApi(name, msg string) error {
	return &errThirdpartyApi{Name: name, Msg: msg}
}

type errThirdpartyApi struct {
	Name, Msg string
}

func (eta *errThirdpartyApi) Error() string {
	return fmt.Sprintf("第三方%s接口服务错误:%s", eta.Name, eta.Msg)
}
