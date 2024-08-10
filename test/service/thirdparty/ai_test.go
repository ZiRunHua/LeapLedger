package thirdparty

import (
	_ "KeepAccount/test/initialize"
)
import (
	_thirdpartyService "KeepAccount/service/thirdparty"
	"context"
	"testing"
)

var aiService = _thirdpartyService.GroupApp.Ai

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
	result, err := aiService.BatchChineseSimilarityMatching(sourceData, targetData, context.TODO())
	if err != nil {
		t.Error(err)
	}
	t.Log(result)
}
