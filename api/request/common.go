package request

import "KeepAccount/global"

type IncomeExpense struct {
	IncomeExpense global.IncomeExpense `json:"Income_expense"`
}
type Name struct {
	Name string
}
type Id struct {
	Id uint
}

type PageData struct {
	page  int
	limit int
}
