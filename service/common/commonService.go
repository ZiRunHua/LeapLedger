package commonService

import (
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/constant"
	utilJwt "github.com/ZiRunHua/LeapLedger/util/jwtTool"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type common struct{}

var Common = new(common)

// CheckCaptchaStatus 判断验证码是否开启
func (cm *common) CheckCaptchaStatus(key string) bool {
	openCaptcha := global.Config.Captcha.OpenCaptcha               // 是否开启防爆次数
	openCaptchaTimeOut := global.Config.Captcha.OpenCaptchaTimeOut // 缓存超时时间
	v, ok := global.Cache.GetInt(key)
	if !ok {
		global.Cache.Set(key, 1, time.Second*time.Duration(openCaptchaTimeOut))
	}
	return openCaptcha == 0 || openCaptcha < v
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

const ExpiresAt time.Duration = 90 * 24 * time.Hour

func (cm *common) MakeCustomClaims(userId uint) jwt.RegisteredClaims {
	expirationTime := time.Now().Add(ExpiresAt)
	return jwt.RegisteredClaims{
		ID:        strconv.Itoa(int(userId)),
		ExpiresAt: &jwt.NumericDate{expirationTime},
		Issuer:    "server",
		Subject:   "user",
	}
}

func (cm *common) ParseToken(tokenStr string) (jwt.RegisteredClaims, error) {
	return utilJwt.ParseToken(tokenStr, []byte(global.Config.System.JwtKey))
}

func (cm *common) GenerateJWT(custom jwt.RegisteredClaims) (string, error) {
	token, err := utilJwt.CreateToken(custom, []byte(global.Config.System.JwtKey))
	if err != nil {
		return "", errors.Wrap(err, "jwt.CreateToken")
	}
	return token, err
}
func (cm *common) RefreshJWT(custom jwt.RegisteredClaims) (token string, newCustom jwt.RegisteredClaims, err error) {
	newCustom = custom
	if newCustom.ExpiresAt.Before(time.Now().Add(ExpiresAt / 3)) {
		newCustom.ExpiresAt.Time = time.Now().Add(ExpiresAt)
	}
	token, err = cm.GenerateJWT(newCustom)
	return
}

// 设置邮件验证码缓存 如果该缓存已存在未过期 则返回ErrOperationTooFrequent 以此避免重复发送和频繁操作
func (cm *common) SetEmailCaptchaCache(email string, emailCaptcha string, expirationTime time.Duration) error {
	key := global.Cache.GetKey(constant.EmailCaptcha, email)
	if _, ok := global.Cache.GetInt(key); true == ok {
		return global.ErrOperationTooFrequent
	}
	key = global.Cache.GetKey(constant.EmailCaptcha, email)
	global.Cache.Set(key, emailCaptcha, expirationTime)
	return nil
}

// 检查邮件验证码错误次数 如果该邮件验证码错误次数过多 则短时禁用该邮件以保护该邮件
func (cm *common) CheckEmailCaptcha(email string, captcha string) error {
	countKey := global.Cache.GetKey(constant.CaptchaEmailErrorCount, email)
	// 检查错误次数
	if global.Config.Captcha.EmailCaptcha > 0 {
		count, ok := global.Cache.GetInt(countKey)
		if false == ok {
			global.Cache.Set(countKey, 1, time.Second*time.Duration(global.Config.Captcha.EmailCaptchaTimeOut))
		} else {
			if count > global.Config.Captcha.EmailCaptcha {
				return global.ErrOperationTooFrequent
			} else {
				err := global.Cache.Increment(countKey, 1)
				if err != nil {
					return err
				}
			}
		}
	}
	// 检查验证码
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
	// 成功
	_ = global.Cache.Delete(countKey)
	// 不清除cache 以此来限制频繁发送相同邮箱的验证码
	// global.Cache.Delete(emailKey)
	return nil
}
