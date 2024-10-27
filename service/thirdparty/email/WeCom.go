package email

import (
	"KeepAccount/global"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type WeCom struct {
	corpId     string
	corpSecret string
	token      string
	requestUrl weComRequestUrl
}
type weComRequestUrl struct {
	getToken  string
	sendEmail string
}

func (w *WeCom) init() {
	w.corpId = global.Config.ThirdParty.WeCom.CorpId
	w.corpSecret = global.Config.ThirdParty.WeCom.CorpSecret
	if w.corpId != "" && w.corpSecret != "" {
		ServiceStatus = true
	} else {
		return
	}
	w.requestUrl.getToken = fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", w.corpId, w.corpSecret,
	)
	if err := w.getToken(); err != nil {
		print(fmt.Sprintf("初始化WoCom邮箱服务失败 err:%v", err))
	}
}

type SendRequest struct {
	To      ToField `json:"to"`
	Subject string  `json:"subject"`
	Content string  `json:"content"`
}

type ToField struct {
	Emails []string `json:"emails"`
}

type responseWeCom struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (w *WeCom) Send(emails []string, subject string, contest string) error {
	if false == ServiceStatus {
		return global.ErrServiceClosed
	}
	jsonData, err := json.Marshal(SendRequest{To: ToField{Emails: emails}, Subject: subject, Content: contest})
	if err != nil {
		return err
	}
	response, err := http.Post(w.requestUrl.sendEmail, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {

		return err
	}
	var responseData responseWeCom
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}
	if err = w.checkCode(
		response.StatusCode, responseData.Errcode, responseData.Errmsg,
	); err != nil {
		if errors.Is(err, tokenExpiredError) {
			w.getToken()
		}
		return err
	}
	return nil
}

type ResponseGetToken struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (w *WeCom) getToken() error {
	response, err := http.Get(w.requestUrl.getToken)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	var responseData ResponseGetToken
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}
	if err = w.checkCode(
		response.StatusCode, responseData.Errcode, responseData.Errmsg,
	); err != nil {
		return err
	}

	w.token = responseData.AccessToken

	w.updateAfterTokenUpdate()
	return nil
}

func (w *WeCom) updateAfterTokenUpdate() {
	w.requestUrl.sendEmail = "https://qyapi.weixin.qq.com/cgi-bin/exmail/app/compose_send?access_token=" + w.token
}

func (w *WeCom) checkCode(status int, code int, msg string) error {
	print(status)
	if status == http.StatusOK {
		switch code {
		case 0:
			return nil
		case 42001:
			return tokenExpiredError
		default:
			return &thirdPartyResponseError{
				StatusCode: status,
				ErrorCode:  code,
				Message:    msg,
			}
		}
	}
	return &thirdPartyResponseError{
		StatusCode: status,
	}
}
