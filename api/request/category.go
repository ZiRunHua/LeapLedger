package request

import (
	"KeepAccount/global/constant"
)

type CategoryOne struct {
	Id            uint
	Name          string
	Icon          string
	FatherId      uint
	IncomeExpense IncomeExpense
}

type CategoryCreateOne struct {
	Name     string `binding:"required"`
	Icon     string `binding:"required"`
	FatherId uint   `binding:"required"`
}

type CategoryUpdateOne struct {
	Name *string
	Icon *string
}

type CategoryCreateOneFather struct {
	Name          string
	IncomeExpense constant.IncomeExpense
}

type CategoryMove struct {
	Previous *uint
	FatherId *uint
}

type CategoryMoveFather struct {
	Previous *uint
}

type CategoryGetTree struct {
	IncomeExpense *constant.IncomeExpense
}

type CategoryGetList struct {
	IncomeExpense *constant.IncomeExpense `binding:"omitempty"`
}

type CategoryMapping struct {
	ChildCategoryId uint
}

type CategoryGetMappingTree struct {
	MappingAccountId uint `binding:"required"`
}
