package commonService

import (
	"KeepAccount/global"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"crypto/sha1"
	"encoding/hex"
	"github.com/pkg/errors"
	"time"
)

type common struct{}

var Common = new(common)

// CheckCaptchaStatus 判断验证码是否开启
func (cm *common) CheckCaptchaStatus(key string) bool {
	openCaptcha := global.GvaConfig.Captcha.OpenCaptcha               // 是否开启防爆次数
	openCaptchaTimeOut := global.GvaConfig.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	v, ok := global.BlackCache.Get(key)
	if !ok {
		global.BlackCache.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}
	return openCaptcha == 0 || openCaptcha < interfaceToInt(v)
}
func (cm *common) HashPassword(username string, password string) string {
	data := []byte(username + password)
	str := sha1.New()
	str.Write(data)
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
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
func (cm *common) MakeCustomClaims(userClientInfo *userModel.UserClientBaseInfo) util.CustomClaims {
	// 设置过期时间
	expirationTime := time.Now().Add(24 * time.Hour)
	return util.CustomClaims{
		UserId:    userClientInfo.UserID,
		ExpiresAt: expirationTime.Unix(),
		Issuer:    "server", // 可自定义
		Subject:   "user",   // 可自定义
	}
}

func (cm *common) GenerateJWT(custom util.CustomClaims) (string, error) {
	jwt := util.NewJWT()
	token, err := jwt.CreateToken(custom)
	if err != nil {
		return "", errors.Wrap(err, "jwt.CreateToken")
	}
	return token, err
}
