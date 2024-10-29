package productService

import (
	"context"
	"github.com/ZiRunHua/LeapLedger/global"
	"github.com/ZiRunHua/LeapLedger/global/db"
	categoryModel "github.com/ZiRunHua/LeapLedger/model/category"
	productModel "github.com/ZiRunHua/LeapLedger/model/product"
	"github.com/pkg/errors"
)

type Product struct {
}

func (proService *Product) MappingTransactionCategory(
	category categoryModel.Category, productTransCat productModel.TransactionCategory, ctx context.Context,
) (*productModel.TransactionCategoryMapping, error) {
	if category.IncomeExpense != productTransCat.IncomeExpense {
		return nil, errors.Wrap(global.ErrInvalidParameter, "")
	}
	mapping := &productModel.TransactionCategoryMapping{
		AccountId:  category.AccountId,
		CategoryId: category.ID,
		PtcId:      productTransCat.ID,
		ProductKey: productTransCat.ProductKey,
	}
	err := db.Get(ctx).Model(mapping).Create(mapping).Error
	return mapping, err
}

func (proService *Product) DeleteMappingTransactionCategory(
	category categoryModel.Category, productTransCat productModel.TransactionCategory, ctx context.Context,
) error {
	if category.IncomeExpense != productTransCat.IncomeExpense {
		return errors.Wrap(global.ErrInvalidParameter, "")
	}
	err := db.Get(ctx).Where(
		"category_id = ? AND ptc_id = ?", category.ID, productTransCat.ID,
	).Delete(&productModel.TransactionCategoryMapping{}).Error
	return err
}
