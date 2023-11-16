package commonService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
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
	openCaptcha := global.Config.Captcha.OpenCaptcha               // 是否开启防爆次数
	openCaptchaTimeOut := global.Config.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	v, ok := global.Cache.Get(key)
	if !ok {
		global.Cache.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
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
func (cm *common) MakeCustomClaims(userId uint) util.CustomClaims {
	// 设置过期时间
	expirationTime := time.Now().Add(24 * time.Hour)
	return util.CustomClaims{
		UserId:    userId,
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

// 设置邮件验证码缓存 如果该缓存已存在未过期 则返回ErrOperationTooFrequent 以此避免重复发送和频繁操作
func (cm *common) SetEmailCaptchaCache(email string, emailCaptcha string, expirationTime time.Duration) error {
	key := global.Cache.GetKey(constant.EmailCaptcha, email)
	if _, ok := global.Cache.Get(key); true == ok {
		return global.ErrOperationTooFrequent
	}
	key = global.Cache.GetKey(constant.EmailCaptcha, email)
	global.Cache.Set(key, emailCaptcha, expirationTime)
	return nil
}

// 检查邮件验证码错误次数 如果该邮件验证码错误次数过多 则短时禁用该邮件以保护该邮件
func (cm *common) CheckEmailCaptcha(email string, captcha string) error {
	countKey := global.Cache.GetKey(constant.CaptchaEmailErrorCount, email)
	//检查错误次数
	if global.Config.Captcha.EmailCaptcha > 0 {
		count, ok := global.Cache.Get(countKey)
		if false == ok {
			global.Cache.Set(countKey, 1, time.Second*time.Duration(global.Config.Captcha.EmailCaptchaTimeOut))
		} else {
			var intCount int
			intCount, ok = count.(int)
			if false == ok {
				panic("cache计数数据转断言int失败")
			}
			if intCount > global.Config.Captcha.EmailCaptcha {
				return global.ErrOperationTooFrequent
			} else {
				err := global.Cache.Increment(countKey, 1)
				if err != nil {
					return err
				}
			}
		}
	}
	//检查验证码
	emailKey := global.Cache.GetKey(constant.EmailCaptcha, email)
	cacheData, ok := global.Cache.Get(emailKey)
	if false == ok {
		return global.ErrVerifyEmailCaptchaFail
	}
	if val, ok := cacheData.(string); ok {
		if val != cacheData {
			return global.ErrVerifyEmailCaptchaFail
		}
	} else {
		panic("cache数据断言为字符串失败")
	}
	//成功
	_ = global.Cache.Delete(countKey)
	//不清除cache 以此来限制频繁发送相同邮箱的验证码
	//global.Cache.Delete(emailKey)
	return nil
}
