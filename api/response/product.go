package response

import (
	"KeepAccount/global/constant"
)

// swagger:response
type ProductGetOne struct {
	UniqueKey string
	Name      string
}

// swagger:response
type ProductGetList struct {
	List []ProductGetOne
}

// swagger:response
type ProductGetTransactionCategory struct {
	Id            uint
	Name          string
	IncomeExpense constant.IncomeExpense
}

// swagger:response
type ProductGetTransactionCategoryList struct {
	List []ProductGetTransactionCategory
}

// swagger:response
type ProductGetMappingTree struct {
	Tree []ProductGetMappingTreeFather
}

// swagger:response
type ProductGetMappingTreeFather struct {
	FatherId uint
	Children []uint
}
