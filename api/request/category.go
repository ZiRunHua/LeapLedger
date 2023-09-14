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
	AccountId     uint
	Name          string
	IncomeExpense constant.IncomeExpense
}

type CategoryMoveCategory struct {
	Previous *uint
	FatherId *uint
}

type CategoryMoveFather struct {
	Previous *uint
}

type CategoryGetTree struct {
	AccountId     uint `binding:"required"`
	IncomeExpense *constant.IncomeExpense
}

type CategoryGetList struct {
	AccountId     uint                    `binding:"required"`
	IncomeExpense *constant.IncomeExpense `binding:"omitempty"`
}

type CategoryMapping struct {
	ChildCategoryId uint
}

type CategoryGetMappingTree struct {
	ParentAccountId uint `binding:"required"`
	ChildAccountId  uint `binding:"required"`
}
