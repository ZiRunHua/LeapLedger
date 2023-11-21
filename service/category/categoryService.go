package categoryService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Category struct {
}

type CreateData struct {
	Name string
	Icon string
}

func (catSvc *Category) NewCategoryData(category *categoryModel.Category) *CreateData {
	return &CreateData{
		Name: category.Name,
		Icon: category.Icon,
	}
}

func (catSvc *Category) CreateOne(father *categoryModel.Father, data *CreateData) (*categoryModel.Category, error) {
	category := &categoryModel.Category{
		AccountID:      father.AccountID,
		FatherID:       father.ID,
		IncomeExpense:  father.IncomeExpense,
		Name:           data.Name,
		Icon:           data.Icon,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	err := category.CreateOne()
	return category, errors.Wrap(err, "category.CreateOne()")
}

func (catSvc *Category) CreateList(
	father *categoryModel.Father, list []CreateData, tx *gorm.DB,
) ([]categoryModel.Category, error) {
	categoryList := []categoryModel.Category{}
	for _, data := range list {
		categoryList = append(
			categoryList, categoryModel.Category{
				AccountID:      father.AccountID,
				FatherID:       father.ID,
				IncomeExpense:  father.IncomeExpense,
				Name:           data.Name,
				Previous:       0,
				OrderUpdatedAt: time.Now(),
			},
		)
	}
	var err error
	if len(categoryList) > 0 {
		err = tx.Create(&categoryList).Error
	}
	return categoryList, errors.Wrap(err, "category.CreateOne()")
}

func (catSvc *Category) CreateOneFather(
	account *accountModel.Account, InEx constant.IncomeExpense, name string,
) (*categoryModel.Father, error) {
	father := &categoryModel.Father{
		AccountID:      account.ID,
		IncomeExpense:  InEx,
		Name:           name,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	err := father.CreateOne()
	return father, errors.Wrap(err, "father.CreateOne()")
}

func (catSvc *Category) MoveCategory(
	category *categoryModel.Category, previous *categoryModel.Category, father *categoryModel.Father,
) error {
	orlPrevious := category.Previous
	if previous != nil && false == previous.IsEmpty() {
		if category.IsEmpty() || previous.AccountID != category.AccountID || previous.IncomeExpense != category.IncomeExpense {
			return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory")
		}
		if false == father.IsEmpty() && (previous.AccountID != father.AccountID || previous.FatherID != father.ID || previous.IncomeExpense != father.IncomeExpense) {
			return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory father")
		}
	}
	err := category.SetPrevious(previous)
	if nil != err {
		return err
	}
	if orlPrevious == 0 {
		// 0作为遍历的起始位置 不能没有
		if _, err = category.GetHead(); errors.Is(err, gorm.ErrRecordNotFound) {
			var newHead categoryModel.Category
			if err = newHead.GetOneByPrevious(orlPrevious); err != nil {
				return err
			}
			if err = newHead.SetPrevious(nil); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (catSvc *Category) MoveFather(father *categoryModel.Father, previous *categoryModel.Father) error {
	if previous != nil && false == previous.IsEmpty() {
		if previous.AccountID != father.AccountID || previous.IncomeExpense != father.IncomeExpense {
			panic("Data anomaly")
		}
	}
	return father.SetPrevious(previous)
}

func (catSvc *Category) SetFather(
	category categoryModel.Category, father categoryModel.Father, previous *categoryModel.Category,
) error {
	if category.FatherID == father.ID || previous.FatherID != father.ID || category.IncomeExpense != father.IncomeExpense || category.AccountID != father.AccountID {
		return errors.Wrap(global.ErrInvalidParameter, "categoryService.SetFather")
	}
	if previous != nil {
		if father.AccountID != previous.AccountID || father.IncomeExpense != previous.IncomeExpense {
			return errors.Wrap(global.ErrInvalidParameter, "categoryService.SetFather")
		}
		category.Previous = previous.ID
	}
	category.FatherID = father.ID
	return category.SetFather()
}

func (catSvc *Category) GetSequenceCategory(
	account *accountModel.Account, incomeExpense *constant.IncomeExpense,
) (map[uint]*[]categoryModel.Category, error) {
	rows, err := new(categoryModel.Category).GetAll(account, incomeExpense)
	if err != nil {
		return map[uint]*[]categoryModel.Category{}, errors.Wrap(err, "")
	}
	var category categoryModel.Category
	tree := make(map[uint]map[uint][]categoryModel.Category)
	heads := make(map[uint]uint)
	for rows.Next() {
		if err = global.GvaDb.ScanRows(rows, &category); err == nil {
			if _, ok := tree[category.FatherID]; !ok {
				tree[category.FatherID] = make(map[uint][]categoryModel.Category)
			}
			if _, ok := tree[category.FatherID][category.Previous]; !ok {
				tree[category.FatherID][category.Previous] = []categoryModel.Category{category}
			} else {
				tree[category.FatherID][category.Previous] = append(
					tree[category.FatherID][category.Previous], category,
				)
			}
		}
	}
	var result = make(map[uint]*[]categoryModel.Category)
	for fatherId, fatherTree := range tree {
		result[fatherId] = &[]categoryModel.Category{}
		catSvc.makeSequenceOfCategory(result[fatherId], fatherTree, heads[fatherId])
	}
	return result, nil
}

func (catSvc *Category) makeSequenceOfCategory(
	queue *[]categoryModel.Category, tree map[uint][]categoryModel.Category, treeKey uint,
) {
	if _, exist := tree[treeKey]; exist {
		for _, child := range tree[treeKey] {
			*queue = append(*queue, child)
			catSvc.makeSequenceOfCategory(queue, tree, child.ID)
		}
	}
}

func (catSvc *Category) GetSequenceFather(
	account *accountModel.Account, incomeExpense *constant.IncomeExpense,
) ([]categoryModel.Father, error) {
	var model categoryModel.Father
	rows, err := model.GetAll(account, incomeExpense)
	if err != nil {
		return []categoryModel.Father{}, err
	}
	var category categoryModel.Father
	var tree = make(map[uint][]categoryModel.Father)
	var head uint = 0
	for rows.Next() {
		if err = global.GvaDb.ScanRows(rows, &category); err == nil {
			tree[category.Previous] = append(tree[category.Previous], category)
		}
	}
	var result = []categoryModel.Father{}
	catSvc.makeSequenceOfFather(&result, tree, head)
	return result, nil
}

func (catSvc *Category) makeSequenceOfFather(
	queue *[]categoryModel.Father, tree map[uint][]categoryModel.Father, treeKey uint,
) {
	if _, exist := tree[treeKey]; exist {
		for _, child := range tree[treeKey] {
			*queue = append(*queue, child)
			catSvc.makeSequenceOfFather(queue, tree, child.ID)
		}
	}
}

func (catSvc *Category) Update(
	category *categoryModel.Category, data categoryModel.CategoryUpdateData, tx *gorm.DB,
) error {
	return categoryModel.Dao.NewCategory(tx).Update(category, &data)
}

func (catSvc *Category) UpdateFather(father *categoryModel.Father, name string) error {
	if name == "" {
		return global.ErrInvalidParameter
	}
	return global.GvaDb.Model(father).Update("name", name).Error
}

func (catSvc *Category) Delete(category *categoryModel.Category) error {
	exits, err := catSvc.existTransaction(category)
	if err != nil {
		return err
	}
	if exits {
		return errors.Wrap(ErrExistTransacion, "delete category")
	}
	return category.GetDb().Delete(category).Error
}

func (catSvc *Category) DeleteFather(father *categoryModel.Father) error {
	if false == father.InTx() {
		return errors.Wrap(global.ErrNotInTransaction, "")
	}

	categoryList := []*categoryModel.Category{}
	err := global.GvaDb.Select("id").Where("father_id = ?", father.ID).Find(&categoryList).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	exits, err := catSvc.existTransaction(categoryList...)
	if err != nil {
		return err
	} else if exits {
		return errors.Wrap(ErrExistTransacion, "delete category")
	}

	err = father.GetDb().Where("father_id = ?", father.ID).Delete(&categoryModel.Category{}).Error
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = father.GetDb().Delete(&father).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (catSvc *Category) existTransaction(categoryList ...*categoryModel.Category) (bool, error) {
	if len(categoryList) < 1 {
		return false, nil
	}
	ids := make([]uint, len(categoryList))
	for key, category := range categoryList {
		ids[key] = category.ID
	}
	var transaction transactionModel.Transaction
	return transaction.Exits("category_id IN (?)", ids)
}
