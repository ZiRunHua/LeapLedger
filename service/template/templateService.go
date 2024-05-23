package templateService

import (
	"KeepAccount/global"
	"KeepAccount/global/contextKey"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	userModel "KeepAccount/model/user"
	accountService "KeepAccount/service/account"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type template struct{}

func (t *template) GetList() ([]accountModel.Account, error) {
	list := []accountModel.Account{}
	err := global.GvaDb.Where("user_id = ?", tempUser.ID).Find(&list).Error
	return list, err
}

func (t *template) CreateAccount(
	user userModel.User, tmplAccount accountModel.Account, ctx context.Context,
) (account accountModel.Account, err error) {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
	if tmplAccount.UserId != tempUser.ID {
		return account, ErrNotBelongTemplate
	}
	account, _, err = accountService.ServiceGroupApp.Base.CreateOne(
		user, tmplAccount.Name, tmplAccount.Icon, tmplAccount.Type, tx,
	)
	if err != nil {
		return
	}
	err = t.CreateCategory(account, tmplAccount, ctx)
	if err != nil {
		return
	}
	return
}

func (t *template) CreateCategory(account accountModel.Account, tmplAccount accountModel.Account, ctx context.Context) error {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
	var err error
	if err = account.ForShare(tx); err != nil {
		return err
	}
	var existCategory bool
	existCategory, err = categoryModel.NewDao(tx).Exist(account)
	if existCategory == true {
		return errors.WithStack(errors.New("交易类型已存在"))
	}
	var tmplFatherList []categoryModel.Father
	tmplFatherList, err = categoryModel.NewDao(tx).GetFatherList(tmplAccount, nil)
	if err != nil {
		return err
	}
	for _, tmplFather := range tmplFatherList {
		if err = t.CreateFatherCategory(account, tmplFather, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (t *template) CreateFatherCategory(
	account accountModel.Account, tmplFather categoryModel.Father, ctx context.Context,
) error {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
	father, err := categoryService.CreateOneFather(account, tmplFather.IncomeExpense, tmplFather.Name, tx)
	if err != nil {
		return err
	}

	tmplCategoryList, err := categoryModel.NewDao(tx).GetListByFather(&tmplFather)
	if err != nil {
		return err
	}

	var category categoryModel.Category
	var mappingList []productModel.TransactionCategoryMapping
	productDao := productModel.NewDao(tx)
	for _, tmplCategory := range tmplCategoryList {
		category, err = categoryService.CreateOne(father, categoryService.NewCategoryData(tmplCategory), ctx)
		if err != nil {
			return err
		}
		mappingList, err = productDao.SelectAllCategoryMappingByCategoryId(tmplCategory.ID)
		if err != nil {
			return err
		}
		for _, tmpMapping := range mappingList {
			mapping := productModel.TransactionCategoryMapping{
				AccountID:  category.AccountId,
				CategoryID: category.ID,
				PtcID:      tmpMapping.PtcID,
				ProductKey: tmpMapping.ProductKey,
			}
			err = tx.Create(&mapping).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}
