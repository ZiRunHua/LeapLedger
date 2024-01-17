package productService

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Product struct {
}
type _productService interface {
	MappingTransactionCategory(
		category *categoryModel.Category, productTransCat *productModel.TransactionCategory,
	) (*productModel.TransactionCategoryMapping, error)
	DeleteMappingTransactionCategory(
		category *categoryModel.Category, productTransCat *productModel.TransactionCategory,
	) error
}

func (proService *Product) MappingTransactionCategory(
	category *categoryModel.Category, productTransCat *productModel.TransactionCategory,
) (*productModel.TransactionCategoryMapping, error) {
	if category.IncomeExpense != productTransCat.IncomeExpense {
		return nil, errors.Wrap(global.ErrInvalidParameter, "")
	}
	mapping := &productModel.TransactionCategoryMapping{
		AccountID:  category.AccountId,
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
	err := global.GvaDb.Where(
		"category_id = ? AND ptc_id = ?", category.ID, productTransCat.ID,
	).Delete(&productModel.TransactionCategoryMapping{}).Error
	return err
}

func (proService *Product) BillImport(
	user userModel.User, account accountModel.Account, product productModel.Product, file *util.FileWithSuffix,
	tx *gorm.DB,
) error {
	importServer := newProductBillImport(user, account, product)
	if err := importServer.init(); err != nil {
		return err
	}
	return importServer.doImport(file, tx)
}
