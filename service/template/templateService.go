package templateService

import (
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	userModel "KeepAccount/model/user"
	accountService "KeepAccount/service/account"
	categoryService "KeepAccount/service/category"
	"gorm.io/gorm"
)

type template struct{}

func (t *template) GetList() ([]accountModel.Account, error) {
	list := []accountModel.Account{}
	err := global.GvaDb.Where("user_id = ?", tempUser.ID).First(&list).Error
	return list, err
}
func (t *template) CreateAccount(
	user *userModel.User, tmplAccount *accountModel.Account, tx *gorm.DB,
) (*accountModel.Account, error) {
	if tmplAccount.UserId != tempUser.ID {
		return nil, ErrNotBelongTemplate
	}
	account, err := accountService.GroupApp.Base.CreateOne(user, tmplAccount.Name, tx)
	if err != nil {
		return nil, err
	}
	tmplFatherList, err := categoryModel.Dao.NewFather(tx).GetListByAccount(tmplAccount)
	if err != nil {
		return nil, err
	}
	for _, tmplFather := range tmplFatherList {
		if err = t.CreateFatherCategory(account, &tmplFather, tx); err != nil {
			return nil, err
		}
	}
	return account, err
}

func (t *template) CreateFatherCategory(
	account *accountModel.Account, tmplFather *categoryModel.Father, tx *gorm.DB,
) error {
	father, err := categoryService.GroupApp.CreateOneFather(account, tmplFather.IncomeExpense, tmplFather.Name)
	if err != nil {
		return err
	}

	tmplCategoryList, err := categoryModel.Dao.New(tx).GetListByFather(tmplFather)
	if err != nil {
		return err
	}

	categoryDataList := []categoryService.CreateData{}
	for _, tmplCategory := range tmplCategoryList {
		categoryDataList = append(categoryDataList, *categoryService.GroupApp.NewCategoryData(&tmplCategory))
	}
	if _, err = categoryService.GroupApp.CreateList(father, categoryDataList, tx); err != nil {
		return err
	}
	return nil
}
