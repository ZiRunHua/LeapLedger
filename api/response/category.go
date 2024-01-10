package response

import (
	"KeepAccount/global/constant"
	categoryModel "KeepAccount/model/category"
)

func CategoryModelToResponse(category *categoryModel.Category) *CategoryOne {
	if category == nil {
		return &CategoryOne{}
	}
	return &CategoryOne{
		Id:            category.ID,
		Name:          category.Name,
		Icon:          category.Icon,
		IncomeExpense: category.IncomeExpense,
	}
}

type CategoryOne struct {
	Id            uint
	Name          string
	Icon          string
	IncomeExpense constant.IncomeExpense
}

type FatherOne struct {
	Id            uint
	Name          string
	IncomeExpense constant.IncomeExpense
	Children      []CategoryOne
}

type CategoryTree struct {
	Tree []FatherOne
}
