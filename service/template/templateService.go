package templateService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	productModel "KeepAccount/model/product"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataTool"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"os"
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
	return account, db.Transaction(ctx, func(ctx *cus.TxContext) error {
		account, _, err = accountService.CreateOne(
			user, accountService.NewCreateData(tmplAccount.Name, tmplAccount.Icon, tmplAccount.Type, tmplAccount.Location), ctx,
		)
		if err != nil {
			return err
		}
		return t.CreateCategory(account, tmplAccount, ctx)
	})
}

func (t *template) CreateCategory(account accountModel.Account, tmplAccount accountModel.Account, ctx context.Context) error {
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
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
func (t *template) CreateAccountByTemplate(tmpl AccountTmpl, user userModel.User, ctx context.Context) (account accountModel.Account, accountUser accountModel.User, err error) {
	account = accountService.NewCreateData(tmpl.Name, tmpl.Icon, tmpl.Type, tmpl.Location)
	account, accountUser, err = accountService.CreateOne(user, account, ctx)
	if err != nil {
		return
	}
	var list dataTool.Slice[any, fatherTmpl] = tmpl.Category
	for _, f := range list.CopyReverse() {
		err = f.create(account, ctx)
		if err != nil {
			return
		}
	}
	return
}

func (t *template) CreateExampleAccount(user userModel.User, ctx context.Context) (account accountModel.Account, accountUser accountModel.User, err error) {
	var accountTmpl AccountTmpl
	err = accountTmpl.ReadFromJson(constant.ExampleAccountJsonPath)
	if err != nil {
		return
	}
	return t.CreateAccountByTemplate(accountTmpl, user, ctx)
}

type AccountTmpl struct {
	Name, Icon, Location string
	Type                 accountModel.Type
	Category             []fatherTmpl
}

func (at *AccountTmpl) ReadFromJson(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, at)
	if err != nil {
		return err
	}
	return nil
}

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
