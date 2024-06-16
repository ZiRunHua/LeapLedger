package categoryService

import (
	gTask "KeepAccount/global/task"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
)

type _task struct{}

func init() {
	gTask.Subscribe[accountModel.Mapping](gTask.TaskMappingCategoryToAccountMapping,
		GroupApp.MappingCategoryToAccountMapping,
	)

	gTask.Subscribe[categoryModel.Category](gTask.TaskUpdateCategoryMapping,
		GroupApp.UpdateCategoryMapping,
	)
}

func (t *_task) UpdateCategoryMapping(category categoryModel.Category) error {
	_ = gTask.Publish[categoryModel.Category](gTask.TaskUpdateCategoryMapping, category)
	return nil
}

func (t *_task) MappingCategoryToAccountMapping(mappingAccount accountModel.Mapping) error {
	_ = gTask.Publish[accountModel.Mapping](gTask.TaskMappingCategoryToAccountMapping, mappingAccount)
	return nil
}
