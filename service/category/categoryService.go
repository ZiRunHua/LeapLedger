package categoryService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
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

func (catSvc *Category) NewCategoryData(category categoryModel.Category) CreateData {
	return CreateData{
		Name: category.Name,
		Icon: category.Icon,
	}
}

func (catSvc *Category) CreateOne(father categoryModel.Father, data CreateData, tx *gorm.DB) (categoryModel.Category, error) {
	category := categoryModel.Category{
		AccountId:      father.AccountID,
		FatherId:       father.ID,
		IncomeExpense:  father.IncomeExpense,
		Name:           data.Name,
		Icon:           data.Icon,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	err := tx.Create(&category).Error
	return category, errors.Wrap(err, "category.CreateOne()")
}

func (catSvc *Category) CreateList(
	father categoryModel.Father, list []CreateData, tx *gorm.DB,
) ([]categoryModel.Category, error) {
	categoryList := []categoryModel.Category{}
	for _, data := range list {
		categoryList = append(
			categoryList, categoryModel.Category{
				AccountId:      father.AccountID,
				FatherId:       father.ID,
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
	account accountModel.Account, InEx constant.IncomeExpense, name string, tx *gorm.DB,
) (categoryModel.Father, error) {
	father := categoryModel.Father{
		AccountID:      account.ID,
		IncomeExpense:  InEx,
		Name:           name,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	err := tx.Create(&father).Error
	return father, errors.Wrap(err, "father.CreateOne()")
}

func (catSvc *Category) MoveCategory(
	category categoryModel.Category, previous *categoryModel.Category, father categoryModel.Father, tx *gorm.DB,
) error {
	orlPrevious := category.Previous
	if previous != nil && false == previous.IsEmpty() {
		if category.IsEmpty() || previous.AccountId != category.AccountId || previous.IncomeExpense != category.IncomeExpense {
			return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory")
		}
		if false == father.IsEmpty() && (previous.AccountId != father.AccountID || previous.FatherId != father.ID || previous.IncomeExpense != father.IncomeExpense) {
			return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory father")
		}
	}
	err := category.SetPrevious(previous, tx)
	if nil != err {
		return err
	}
	if orlPrevious == 0 {
		// 0作为遍历的起始位置 不能没有
		if _, err = category.GetHead(tx); errors.Is(err, gorm.ErrRecordNotFound) {
			var newHead categoryModel.Category
			if err = newHead.GetOneByPrevious(orlPrevious, tx); err != nil {
				return err
			}
			if err = newHead.SetPrevious(nil, tx); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (catSvc *Category) MoveFather(father categoryModel.Father, previous *categoryModel.Father, tx *gorm.DB) error {
	if previous != nil && false == previous.IsEmpty() {
		if previous.AccountID != father.AccountID || previous.IncomeExpense != father.IncomeExpense {
			panic("Data anomaly")
		}
	}
	return father.SetPrevious(previous, tx)
}

func (catSvc *Category) GetSequenceCategory(
	account accountModel.Account, incomeExpense *constant.IncomeExpense,
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
			if _, ok := tree[category.FatherId]; !ok {
				tree[category.FatherId] = make(map[uint][]categoryModel.Category)
			}
			if _, ok := tree[category.FatherId][category.Previous]; !ok {
				tree[category.FatherId][category.Previous] = []categoryModel.Category{category}
			} else {
				tree[category.FatherId][category.Previous] = append(
					tree[category.FatherId][category.Previous], category,
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
	account accountModel.Account, incomeExpense *constant.IncomeExpense,
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
	category categoryModel.Category, data categoryModel.CategoryUpdateData, tx *gorm.DB,
) error {
	return categoryModel.NewDao(tx).Update(category, data)
}

func (catSvc *Category) UpdateFather(father categoryModel.Father, name string) error {
	if name == "" {
		return global.ErrInvalidParameter
	}
	return global.GvaDb.Model(&father).Update("name", name).Error
}

func (catSvc *Category) Delete(category categoryModel.Category, tx *gorm.DB) error {
	exits, err := catSvc.existTransaction(category)
	if err != nil {
		return err
	}
	if exits {
		return errors.Wrap(ErrExistTransacion, "delete category")
	}
	return tx.Delete(&category).Error
}

func (catSvc *Category) DeleteFather(father categoryModel.Father, tx *gorm.DB) error {
	var categoryList []categoryModel.Category
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

	err = tx.Where("father_id = ?", father.ID).Delete(&categoryModel.Category{}).Error
	if err != nil {
		return errors.Wrap(err, "")
	}

	err = tx.Delete(&father).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (catSvc *Category) existTransaction(categoryList ...categoryModel.Category) (bool, error) {
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

func (catSvc *Category) checkMappingParam(parent, child categoryModel.Category, operator userModel.User, tx *gorm.DB) error {
	if parent.AccountId == child.AccountId {
		return global.ErrAccountId
	}
	if parent.IncomeExpense != child.IncomeExpense {
		return errors.WithStack(global.ErrInvalidParameter)
	}
	accountUser, err := accountModel.NewDao(tx).SelectUser(parent.AccountId, operator.ID)
	if err != nil {
		return err
	}
	if false == accountUser.HavePermission(accountModel.UserPermissionOwnEditor) {
		return global.ErrNoPermission
	}
	return nil
}

func (catSvc *Category) MappingCategory(parent, child categoryModel.Category, operator userModel.User, tx *gorm.DB) (mapping categoryModel.Mapping, err error) {
	err = catSvc.checkMappingParam(parent, child, operator, tx)
	if err != nil {
		return
	}
	mapping, err = categoryModel.NewDao(tx).CreateMapping(parent, child)
	return
}

func (catSvc *Category) DeleteMapping(parent, child categoryModel.Category, operator userModel.User, tx *gorm.DB) error {
	err := catSvc.checkMappingParam(parent, child, operator, tx)
	if err != nil {
		return err
	}
	err = tx.Where(
		"parent_category_id = ? AND child_category_id = ?", parent.ID, child.ID,
	).Delete(&categoryModel.Mapping{}).Error
	return err
}

func (catSvc *Category) MappingAccountCategoryByAI(mainAccount, mappingAccount accountModel.Account, tx *gorm.DB) error {
	categoryDao := categoryModel.NewDao(tx)
	for _, ie := range []constant.IncomeExpense{constant.Income, constant.Expense} {
		categoryDao.GetListByAccount()
	}
	return err
}
