package script

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/contextKey"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	"context"
	"gorm.io/gorm"
)

type _category struct {
}

var Category = _category{}

type fatherTmpl struct {
	Name     string
	Ie       constant.IncomeExpense
	Children []categoryTmpl
}

func (ft *fatherTmpl) create(account accountModel.Account, tx *gorm.DB) error {
	father, err := categoryService.CreateOneFather(account, ft.Ie, ft.Name, tx)
	if err != nil {
		return err
	}
	for _, child := range ft.Children {
		_, err = child.create(father, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

type categoryTmpl struct {
	Name, Icon  string
	Ie          constant.IncomeExpense
	MappingPtcs []struct {
		ProductKey productModel.KeyValue
		Name       string
	}
}

func (ct *categoryTmpl) create(father categoryModel.Father, tx *gorm.DB) (category categoryModel.Category, err error) {
	ctx := context.WithValue(context.Background(), contextKey.Tx, tx)
	category, err = categoryService.CreateOne(father, categoryService.NewCategoryData(ct.Name, ct.Icon), ctx)
	if err != nil {
		return
	}
	var ptc productModel.TransactionCategory
	for _, mappingPtc := range ct.MappingPtcs {
		ptc, err = productModel.NewDao(tx).SelectByName(mappingPtc.ProductKey, father.IncomeExpense, mappingPtc.Name)
		if err != nil {
			return
		}
		_, err = productService.MappingTransactionCategory(category, ptc)
		if err != nil {
			return
		}
	}
	return
}
