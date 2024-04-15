package thirdparty

import (
	"KeepAccount/global"
	_ "KeepAccount/global"
	_ "KeepAccount/global/constant"
	_ "KeepAccount/initialize"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestChineseSimilarityMatching(t *testing.T) {
	sourceData := []string{
		"餐饮美食", "酒店旅游", "运动户外", "美容美发", "生活服务", "爱车养车",
		"母婴亲子", "服饰装扮", "日用百货", "文化休闲", "数码电器", "教育培训",
		"家居家装", "宠物", "商业服务", "医疗健康", "其他",
	}

	targetData := []string{
		"住房", "交通", "购物", "通讯", "餐饮", "杂项",
		"医疗", "旅游", "文化休闲", "其他",
	}
	result, err := (&aiServer{}).ChineseSimilarityMatching(sourceData, targetData)
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}

type aiServer struct {
}

const ChineseSimilarityMatching

func (a *aiServer) ChineseSimilarityMatching(SourceData, TargetData []string) (SimilarityResponse, error) {
	var post placeholder

	requester, response := SimilarityRequester{}, SimilarityResponse{}
	err := requester.setRequestData(SourceData, TargetData)
	if err != nil {
		return response, err
	}
	err = requester.sendRequest()
	if err != nil {
		return response, err
	}
	response, err = requester.getResponseData()
	if err != nil {
		return response, err
	}
	return response, nil
}

type aiApiResponse struct {
	code int
	msg  string
}

func (a *aiApiResponse) isSuccess() bool { return a.code == 200 }

type AiAPICommunicator struct {
	APICommunicator
}

func (a *AiAPICommunicator) getUrl() string {
	return global.Config.ThirdParty.Ai.GetPortalSite() + a.path
}

type SimilarityRequester struct {
	AiAPICommunicator
}

func (a *SimilarityRequester) getUrl() string {
	a.path = "/similarity/matching"
	return a.AiAPICommunicator.getUrl()
}
func (s *SimilarityRequester) setRequestData(SourceData, TargetData []string) error {
	personMap := map[string]interface{}{
		"SourceData": SourceData,
		"TargetData": TargetData,
	}
	jsonData, err := json.Marshal(personMap)
	if err != nil {
		return err
	}
	s.requestData = jsonData
	return nil
}

func (s *SimilarityRequester) getResponseData() (SimilarityResponse, error) {
	var response SimilarityResponse
	err := json.Unmarshal([]byte(s.responseData), &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

type SimilarityResponse struct {
	aiApiResponse
	data []struct {
		Source, Target string
		Similarity     float32
	}
}

func (s *SimilarityRequester) sendRequest() error {
	s.path = s.getUrl()
	return s.APICommunicator.sendRequest()
}

type APICommunicator struct {
	path         string
	requestData  []byte
	responseData []byte
	state        int
}

func (r *APICommunicator) sendRequest() error {
	// 构建请求
	fmt.Println(r.path)
	req, err := http.NewRequest("POST", r.path, bytes.NewBuffer(r.requestData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 处理响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	r.responseData = body

	// 设置状态码
	r.state = resp.StatusCode

	return nil
}
func (f *APICommunicator) setRequestData() {}

func (r *APICommunicator) getResponseData() {}
func (a *APICommunicator) getUrl() string   { return "" }
