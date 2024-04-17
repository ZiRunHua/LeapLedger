package thirdpartyService

import (
	"KeepAccount/global"
	"context"
	"github.com/carlmjohnson/requests"
)

const AI_SERVER_NAME = "AI"
const API_SIMILARITY_MATCHING = "/similarity/matching"

type aiApiResponse struct {
	Code int
	Msg  string
}

func (a *aiApiResponse) isSuccess() bool { return a.Code == 200 }

type aiServer struct {
}

func (as *aiServer) getBaseUrl() string {
	return global.Config.ThirdParty.Ai.GetPortalSite()
}
func (as *aiServer) IsOpen() bool {
	return global.Config.ThirdParty.Ai.IsOpen()
}
func (as *aiServer) ChineseSimilarityMatching(SourceData, TargetData []string, ctx context.Context) (map[string]string, error) {
	if false == as.IsOpen() {
		return make(map[string]string), nil
	}
	var response struct {
		aiApiResponse
		Data []struct {
			Source, Target string
			Similarity     float32
		}
	}
	err := requests.
		URL(as.getBaseUrl()).Path(API_SIMILARITY_MATCHING).
		BodyJSON(map[string]interface{}{
			"SourceData": SourceData, "TargetData": TargetData,
		}).
		ToJSON(&response).
		Fetch(ctx)

	if err != nil {
		return nil, err
	}
	if false == response.isSuccess() {
		return nil, global.NewErrThirdpartyApi(AI_SERVER_NAME, response.Msg)
	}

	result := make(map[string]string)
	minSimilarity := global.Config.ThirdParty.Ai.MinSimilarity
	for _, item := range response.Data {
		if item.Similarity >= minSimilarity {
			result[item.Source] = item.Target
		}
	}
	return result, nil
}
