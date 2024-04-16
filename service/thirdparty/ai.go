package thirdpartyService

import (
	"KeepAccount/global"
	"github.com/carlmjohnson/requests"
	"github.com/gin-gonic/gin"
)

const AI_SERVER_NAME = "AI"
const API_SIMILARITY_MATCHING = "/similarity/matching"

type aiApiResponse struct {
	code int
	msg  string
}

func (a *aiApiResponse) isSuccess() bool { return a.code == 200 }

type aiServer struct {
}

func (as *aiServer) getBaseUrl() string {
	return global.Config.ThirdParty.Ai.GetPortalSite()
}

func (as *aiServer) ChineseSimilarityMatching(SourceData, TargetData []string, ctx *gin.Context) (map[string]string, error) {
	var response struct {
		aiApiResponse
		data []struct {
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
		return nil, global.NewErrThirdpartyApi(AI_SERVER_NAME, response.msg)
	}

	result := make(map[string]string)
	minSimilarity := global.Config.ThirdParty.Ai.MinSimilarity
	for _, item := range response.data {
		if item.Similarity >= minSimilarity {
			result[item.Source] = item.Target
		}
	}
	return result, nil
}
