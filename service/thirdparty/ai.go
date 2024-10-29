package thirdpartyService

import (
	"context"
	"strings"

	"KeepAccount/global"
	"github.com/go-resty/resty/v2"
)

const AI_SERVER_NAME = "AI"
const API_SIMILARITY_MATCHING = "/similarity/matching"

type aiServer struct{}

func (as *aiServer) getBaseUrl() string {
	return global.Config.ThirdParty.Ai.GetPortalSite()
}

func (as *aiServer) IsOpen() bool {
	return global.Config.ThirdParty.Ai.IsOpen()
}

func (as *aiServer) ChineseSimilarityMatching(sourceStr string, targetList []string, ctx context.Context) (
	target string, err error,
) {
	if false == as.IsOpen() {
		for _, targetStr := range targetList {
			if strings.Compare(sourceStr, targetStr) == 0 {
				return targetStr, nil
			}
		}
		return target, nil
	}
	responseData, err := as.requestChineseSimilarity([]string{sourceStr}, targetList, ctx)
	if err != nil {
		return
	}
	minSimilarity := global.Config.ThirdParty.Ai.MinSimilarity
	if len(responseData) > 0 && responseData[0].Similarity >= minSimilarity {
		return responseData[0].Target, nil
	}
	return
}

func (as *aiServer) BatchChineseSimilarityMatching(sourceList, targetList []string, ctx context.Context) (
	map[string]string, error,
) {
	if false == as.IsOpen() {
		targetNameMap := make(map[string]struct{})
		for _, targetStr := range targetList {
			targetNameMap[targetStr] = struct{}{}
		}
		result := make(map[string]string)
		for _, sourceStr := range sourceList {
			if _, exist := targetNameMap[sourceStr]; !exist {
				continue
			}
			result[sourceStr] = sourceStr
		}
		return result, nil
	}
	responseData, err := as.requestChineseSimilarity(sourceList, targetList, ctx)
	if err != nil {
		return nil, err
	}

	minSimilarity, result := global.Config.ThirdParty.Ai.MinSimilarity, make(map[string]string)
	for _, item := range responseData {
		if item.Similarity >= minSimilarity {
			result[item.Source] = item.Target
		}
	}
	return result, nil
}

type chineseSimilarityResponse []struct {
	Source, Target string
	Similarity     float32
}

func (as *aiServer) requestChineseSimilarity(
	SourceList, TargetList []string,
	ctx context.Context) (chineseSimilarityResponse, error) {
	var response struct {
		aiApiResponse
		Data chineseSimilarityResponse
	}
	_, err := resty.New().R().SetContext(ctx).SetBody(
		map[string]interface{}{
			"SourceData": SourceList, "TargetData": TargetList,
		},
	).SetResult(&response).Post(as.getBaseUrl() + API_SIMILARITY_MATCHING)

	if err != nil {
		return nil, err
	}
	if false == response.isSuccess() {
		return nil, global.NewErrThirdpartyApi(AI_SERVER_NAME, response.Msg)
	}
	return response.Data, nil
}

type aiApiResponse struct {
	Code int
	Msg  string
}

func (a *aiApiResponse) isSuccess() bool { return a.Code == 200 }
