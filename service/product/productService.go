package productService

import (
	"KeepAccount/global"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	"KeepAccount/util"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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

func (proService *Product) BillImport(
	accountUser accountModel.User, account accountModel.Account, product productModel.Product,
	file *util.FileWithSuffix,
	tx *gorm.DB,
) error {
	if accountUser.AccountId != account.ID {
		return global.ErrAccountId
	}
	err := accountUser.CheckTransAddByUserId(accountUser.UserId)
	if err != nil {
		return err
	}
	importServer := newProductBillImport(accountUser, account, product)
	if err = importServer.init(); err != nil {
		return err
	}
	return importServer.doImport(file, tx)
}
