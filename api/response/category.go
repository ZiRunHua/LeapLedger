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

func (co *CategoryOne) SetData(category categoryModel.Category) error {
	co.Id = category.ID
	co.Name = category.Name
	co.Icon = category.Icon
	co.IncomeExpense = category.IncomeExpense
	return nil
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
