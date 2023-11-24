package productService

import (
	"KeepAccount/global"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	"github.com/pkg/errors"
)

type Product struct {
	BillImport ProductBillImport
}
type _productService interface {
	MappingTransactionCategory(category *categoryModel.Category, productTransCat *productModel.TransactionCategory) (*productModel.TransactionCategoryMapping, error)
	DeleteMappingTransactionCategory(category *categoryModel.Category, productTransCat *productModel.TransactionCategory) error
}

func (proService *Product) MappingTransactionCategory(
	category *categoryModel.Category, productTransCat *productModel.TransactionCategory,
) (*productModel.TransactionCategoryMapping, error) {
	if category.IncomeExpense != productTransCat.IncomeExpense {
		return nil, errors.Wrap(global.ErrInvalidParameter, "")
	}
	mapping := &productModel.TransactionCategoryMapping{
		AccountID:  category.AccountID,
		CategoryID: category.ID,
		PtcID:      productTransCat.ID,
		ProductKey: productTransCat.ProductKey,
	}
	err := global.GvaDb.Model(mapping).Create(mapping).Error
	return mapping, err
}

func (proService *Product) DeleteMappingTransactionCategory(
	category *categoryModel.Category, productTransCat *productModel.TransactionCategory,
) error {
	if category.IncomeExpense != productTransCat.IncomeExpense {
		return errors.Wrap(global.ErrInvalidParameter, "")
	}
	err := global.GvaDb.Model(&productModel.TransactionCategoryMapping{}).Delete("category_id = ? AND ptc_id = ?", category.ID, productTransCat.ID).Error
	return err
}
