package tUtil

import (
	"math/rand"

	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	transactionModel "github.com/ZiRunHua/LeapLedger/model/transaction"
)

type Get struct {
}

func (g *Get) Category() categoryModel.Category {
	return testCategoryList[rand.Intn(len(testCategoryList))]
}

func (g *Get) TransInfo() transactionModel.Info {
	return build.TransInfo(testUser, g.Category())
}
