package request

import "KeepAccount/global"

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
	IncomeExpense global.IncomeExpense
}
type CategoryMoveCategory struct {
	Previous *uint
	FatherId *uint
}
type CategoryMoveFather struct {
	Previous *uint
}
type CategoryGetTree struct {
	AccountId     uint                 `binding:"required"`
	IncomeExpense global.IncomeExpense `binding:"required"`
}
type CategoryGetList struct {
	AccountId     uint                 `binding:"required"`
	IncomeExpense global.IncomeExpense `binding:"required"`
}
