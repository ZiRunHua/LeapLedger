package categoryService

import (
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
)

type _task struct{}

func init() {
	nats.SubscribeTaskWithPayload[accountModel.Mapping](
		nats.TaskMappingCategoryToAccountMapping,
		GroupApp.MappingCategoryToAccountMapping,
	)

	nats.SubscribeTaskWithPayload[categoryModel.Category](
		nats.TaskUpdateCategoryMapping,
		GroupApp.UpdateCategoryMapping,
	)
}

func (t *_task) UpdateCategoryMapping(category categoryModel.Category) error {
	_ = nats.PublishTaskWithPayload[categoryModel.Category](nats.TaskUpdateCategoryMapping, category)
	return nil
}

func (t *_task) MappingCategoryToAccountMapping(mappingAccount accountModel.Mapping) error {
	_ = nats.PublishTaskWithPayload[accountModel.Mapping](nats.TaskMappingCategoryToAccountMapping, mappingAccount)
	return nil
}
