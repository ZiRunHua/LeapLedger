package script

import (
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	"KeepAccount/util/dataTool"
	"context"
)

type fatherTmpl struct {
	Name     string
	Ie       constant.IncomeExpense
	Children []categoryTmpl
}

func (ft *fatherTmpl) create(account accountModel.Account, ctx context.Context) error {
	father, err := categoryService.CreateOneFather(account, ft.Ie, ft.Name, ctx)
	if err != nil {
		return err
	}
	var list dataTool.Slice[any, categoryTmpl] = ft.Children
	for _, child := range list.CopyReverse() {
		_, err = child.create(father, ctx)
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

func (ct *categoryTmpl) create(father categoryModel.Father, ctx context.Context) (category categoryModel.Category, err error) {
	category, err = categoryService.CreateOne(father, categoryService.NewCategoryData(ct.Name, ct.Icon), ctx)
	if err != nil {
		return
	}
	var ptc productModel.TransactionCategory
	for _, mappingPtc := range ct.MappingPtcs {
		ptc, err = productModel.NewDao(db.Get(ctx)).SelectByName(mappingPtc.ProductKey, father.IncomeExpense, mappingPtc.Name)
		if err != nil {
			return
		}
		_, err = productService.MappingTransactionCategory(category, ptc, ctx)
		if err != nil {
			return
		}
	}
	return
}
