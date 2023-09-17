package response

import "KeepAccount/global"

type ProductGetOne struct {
	UniqueKey string
	Name      string
}

type ProductGetList struct {
	List []ProductGetOne
}

type ProductGetTransactionCategory struct {
	Id            uint
	Name          string
	IncomeExpense global.IncomeExpense
}

type ProductGetTransactionCategoryList struct {
	List []ProductGetTransactionCategory
}
type ProductGetMappingTree struct {
	Tree []ProductGetMappingTreeFather
}
type ProductGetMappingTreeFather struct {
	FatherId uint
	Children []uint
}
