package request

import (
	"KeepAccount/global/constant"
)

type CategoryOne struct {
	Id            uint
	Name          string
	FatherId      uint
	IncomeExpense IncomeExpense
}

type CategoryCreateOne struct {
	Name     string
	FatherId uint
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
	AccountId     uint                   `binding:"required"`
	IncomeExpense constant.IncomeExpense `binding:"required"`
}
type CategoryGetList struct {
	AccountId     uint                   `binding:"required"`
	IncomeExpense constant.IncomeExpense `binding:"required"`
}
