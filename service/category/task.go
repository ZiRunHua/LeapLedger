package categoryService

import (
	"github.com/ZiRunHua/LeapLedger/global/nats"
	accountModel "github.com/ZiRunHua/LeapLedger/model/account"
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
)

type _task struct{}

func init() {
	nats.SubscribeTaskWithPayloadAndProcessInTransaction[accountModel.Mapping](
		nats.TaskMappingCategoryToAccountMapping,
		GroupApp.MappingCategoryToAccountMapping,
	)

	nats.SubscribeTaskWithPayload[categoryModel.Category](
		nats.TaskUpdateCategoryMapping,
		GroupApp.UpdateCategoryMapping,
	)
}

func (t *_task) MappingCategoryToAccountMapping(mappingAccount accountModel.Mapping) error {
	_ = nats.PublishTaskWithPayload[accountModel.Mapping](nats.TaskMappingCategoryToAccountMapping, mappingAccount)
	return nil
}
