package response

import (
	"github.com/ZiRunHua/LeapLedger/global/constant"
)

type ProductOne struct {
	UniqueKey string
	Name      string
}

type ProductList struct {
	List []ProductOne
}

type ProductTransactionCategory struct {
	Id            uint
	Name          string
	IncomeExpense constant.IncomeExpense
}

type ProductTransactionCategoryList struct {
	List []ProductTransactionCategory
}

type ProductMappingTree struct {
	Tree []ProductMappingTreeFather
}

type ProductMappingTreeFather struct {
	FatherId uint
	Children []uint
}
