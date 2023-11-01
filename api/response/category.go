package response

import "KeepAccount/global/constant"

type CategoryOne struct {
	Id            uint
	Name          string
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
