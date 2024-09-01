package templateService

import (
	"KeepAccount/global"
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	userModel "KeepAccount/model/user"
	accountService "KeepAccount/service/account"
	"KeepAccount/util/dataTool"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type template struct{}

var TemplateApp = &template{}

func (t *template) GetList() ([]accountModel.Account, error) {
	var list []accountModel.Account
	err := global.GvaDb.Where("user_id = ?", TmplUserId).Find(&list).Error
	return list, err
}

func (t *template) GetListByRank(ctx context.Context) (result []accountModel.Account, err error) {
	var list dataTool.Slice[uint, rankMember]
	list, err = rank.GetAll(ctx)
	if err != nil {
		return
	}
	ids := list.ExtractValues(func(member rankMember) uint { return member.id })
	if len(ids) == 0 {
		return
	}
	err = global.GvaDb.Where("id IN (?)", ids).Find(&result).Error
	return
}
func (t *template) rankOnceIncr(userId uint, tmplAccount accountModel.Account, ctx context.Context) error {
	member := newRankMember(tmplAccount)
	_, err := rank.OnceIncrWeight(member, userId, time.Now().Unix(), ctx)
	return err
}
func (t *template) CreateAccount(
	user userModel.User, tmplAccount accountModel.Account, ctx context.Context,
) (account accountModel.Account, err error) {
	if tmplAccount.UserId != TmplUserId {
		return account, ErrNotBelongTemplate
	}
	return account, db.Transaction(ctx, func(ctx *cusCtx.TxContext) error {
		account, _, err = accountService.ServiceGroupApp.CreateOne(
			user, tmplAccount.Name, tmplAccount.Icon, tmplAccount.Type, ctx,
		)
		if err != nil {
			return err
		}
		return t.CreateCategory(account, tmplAccount, ctx)
	})
}

func (t *template) CreateCategory(account accountModel.Account, tmplAccount accountModel.Account, ctx context.Context) error {
	return db.Transaction(ctx, func(ctx *cusCtx.TxContext) error {
		tx := db.Get(ctx)
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
		categoryDao := categoryModel.NewDao(tx)
		tmplFatherList, err = categoryDao.GetFatherList(tmplAccount, nil)
		if err != nil {
			return err
		}
		categoryDao.OrderFather(tmplFatherList)
		for _, tmplFather := range tmplFatherList {
			if err = t.createFatherCategory(account, tmplFather, ctx); err != nil {
				return err
			}
		}
		err = t.rankOnceIncr(account.UserId, tmplAccount, ctx)
		if err != nil {
			errorLog.Error("CreateAccount => rankOnceIncr", zap.Error(err))
			err = nil
		}
		return nil
	})
}

func (t *template) createFatherCategory(
	account accountModel.Account, tmplFather categoryModel.Father, ctx context.Context,
) error {
	tx := db.Get(ctx)
	father, err := categoryService.CreateOneFather(account, tmplFather.IncomeExpense, tmplFather.Name, ctx)
	if err != nil {
		return err
	}
	categoryDao := categoryModel.NewDao(tx)
	tmplCategoryList, err := categoryDao.GetListByFather(tmplFather)
	if err != nil {
		return err
	}
	categoryDao.Order(tmplCategoryList)
	var category categoryModel.Category
	var mappingList []productModel.TransactionCategoryMapping
	productDao := productModel.NewDao(tx)
	for _, tmplCategory := range tmplCategoryList {
		category, err = categoryService.CreateOne(father, categoryService.NewCategoryData(tmplCategory.Name, tmplCategory.Icon), ctx)
		if err != nil {
			return err
		}
		mappingList, err = productDao.SelectAllCategoryMappingByCategoryId(tmplCategory.ID)
		if err != nil {
			return err
		}
		for _, tmpMapping := range mappingList {
			mapping := productModel.TransactionCategoryMapping{
				AccountId:  category.AccountId,
				CategoryId: category.ID,
				PtcId:      tmpMapping.PtcId,
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
