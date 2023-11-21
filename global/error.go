package global

import (
	"github.com/pkg/errors"
)

var (
	ErrNotInTransaction = errors.New("run error:not in transaction")
)

var (
	ErrNotBelongCurrentUser = errors.New("not belong current user")
	ErrInvalidRequest       = errors.New("invalid request")
	ErrInvalidParameter     = errors.New("invalid parameter")
)

// 数据校验
var (
	ErrDataIsEmpty = NewErrDataIsEmpty("")
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
var ErrServiceClosed = errors.New("服务未开启")

// 对应constant.UserAction
var ErrUnsupportedUserAction = errors.New("暂不支持该操作")

// 用户
var ErrSameAsTheOldPassword = errors.New("新旧密码相同")
