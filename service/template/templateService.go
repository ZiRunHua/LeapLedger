package templateService

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	accountService "KeepAccount/service/account"
	categoryService "KeepAccount/service/category"
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
	user userModel.User, tmplAccount accountModel.Account, tx *gorm.DB,
) (account accountModel.Account, err error) {
	if tmplAccount.UserId != tempUser.ID {
		return account, ErrNotBelongTemplate
	}
	account, _, err = accountService.ServiceGroupApp.Base.CreateOne(
		user, tmplAccount.Name, tmplAccount.Icon, tmplAccount.Type, tx,
	)
	if err != nil {
		return
	}
	return
}

func (t *template) CreateCategory(account accountModel.Account, tmplAccount accountModel.Account, tx *gorm.DB) error {
	var err error
	if err = account.ForUpdate(tx); err != nil {
		return err
	}
	var existCategory bool
	existCategory, err = categoryModel.Dao.NewCategory(tx).Exist(account)
	if existCategory == true {
		return errors.WithStack(errors.New("交易类型已存在"))
	}
	var tmplFatherList []categoryModel.Father
	tmplFatherList, err = categoryModel.Dao.NewFather(tx).GetListByAccount(&tmplAccount)
	if err != nil {
		return err
	}
	for _, tmplFather := range tmplFatherList {
		if err = t.CreateFatherCategory(account, tmplFather, tx); err != nil {
			return err
		}
	}
	return nil
}

func (t *template) CreateFatherCategory(
	account accountModel.Account, tmplFather categoryModel.Father, tx *gorm.DB,
) error {
	father, err := categoryService.GroupApp.CreateOneFather(account, tmplFather.IncomeExpense, tmplFather.Name, tx)
	if err != nil {
		return err
	}

	tmplCategoryList, err := categoryModel.Dao.NewCategory(tx).GetListByFather(&tmplFather)
	if err != nil {
		return err
	}

	categoryDataList := []categoryService.CreateData{}
	for _, tmplCategory := range tmplCategoryList {
		categoryDataList = append(categoryDataList, categoryService.GroupApp.NewCategoryData(tmplCategory))
	}
	if _, err = categoryService.GroupApp.CreateList(father, categoryDataList, tx); err != nil {
		return err
	}
	return nil
}
