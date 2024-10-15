package tUtil

import (
	"math/rand"

	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
)

type Get struct {
}

func (g *Get) Category() categoryModel.Category {
	return testCategoryList[rand.Intn(len(testCategoryList))]
}

func (g *Get) TransInfo() transactionModel.Info {
	return build.TransInfo(testUser, g.Category())
}
