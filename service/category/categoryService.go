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
		AccountId:      father.AccountId,
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
				AccountId:      father.AccountId,
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
		AccountId:      account.ID,
		IncomeExpense:  InEx,
		Name:           name,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	err := tx.Create(&father).Error
	return father, errors.Wrap(err, "father.CreateOne()")
}

func (catSvc *Category) MoveCategory(
	category categoryModel.Category, previous *categoryModel.Category, father *categoryModel.Father,
	operator userModel.User, tx *gorm.DB,
) error {
	// 数据校验
	if previous != nil && (category.ID == previous.ID || previous.AccountId != category.AccountId || previous.IncomeExpense != category.IncomeExpense) {
		return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory")
	}
	if father != nil && (father.AccountId != category.AccountId || father.IncomeExpense != category.IncomeExpense) {
		return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory")
	}
	if previous != nil && father != nil && previous.FatherId != father.ID || previous == nil && father == nil {
		return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveCategory")
	}
	accountUser, err := accountModel.NewDao(tx).SelectUser(category.AccountId, operator.ID)
	if err != nil {
		return err
	} else if false == accountUser.HavePermission(accountModel.UserPermissionCreator) {
		return global.ErrNoPermission
	}

	// 处理
	categoryDao := categoryModel.NewDao(tx)
	firstChild, err := categoryDao.SelectFirstChild(category.ID)
	if err == nil {
		// 将排头的子更新为当前所有子的父
		err = tx.Model(&firstChild).Select("previous", "order_updated_at").Updates(category).Error
		if err != nil {
			return err
		}
		err = categoryDao.UpdateChildPrevious(category.ID, firstChild.ID)
		if err != nil {
			return err
		}
	} else if false == errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 最后更新当前交易类型的位置
	if previous != nil {
		err = tx.Model(&category).Select("previous", "father_id", "order_updated_at").Updates(
			categoryModel.Category{
				Previous:       previous.ID,
				FatherId:       previous.FatherId,
				OrderUpdatedAt: time.Now(),
			},
		).Error
	} else {
		err = tx.Model(&category).Select("previous", "father_id", "order_updated_at").Updates(
			categoryModel.Category{
				Previous:       0,
				FatherId:       father.ID,
				OrderUpdatedAt: time.Now(),
			},
		).Error
	}
	if err != nil {
		return err
	}
	return nil
}

func (catSvc *Category) MoveFather(father categoryModel.Father, previous *categoryModel.Father, tx *gorm.DB) error {
	if previous != nil && (previous.AccountId != father.AccountId || previous.IncomeExpense != father.IncomeExpense || father.ID == previous.ID) {
		return errors.Wrap(global.ErrInvalidParameter, "categoryService.MoveFather")
	}

	categoryDao := categoryModel.NewDao(tx)
	firstChild, err := categoryDao.SelectFatherFirstChild(father.ID)
	if err == nil {
		// 将排头的子更新为当前所有子的父
		err = tx.Model(&firstChild).Select("previous", "order_updated_at").Updates(father).Error
		if err != nil {
			return err
		}
		err = categoryDao.UpdateFatherChildPrevious(father.ID, firstChild.ID)
		if err != nil {
			return err
		}
	} else if false == errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 最后更新father的位置
	var previousId uint
	if previous != nil {
		previousId = previous.ID
	} else {
		// 未传入previous则移动到头部
		previousId = 0
	}
	return tx.Model(&father).Select("previous", "order_updated_at").Updates(
		categoryModel.Father{
			Previous:       previousId,
			OrderUpdatedAt: time.Now(),
		},
	).Error
}

// GetSequenceCategoryByFather 返回排序后的category
func (catSvc *Category) GetSequenceCategoryByFather(father categoryModel.Father) (sequenceCategory []categoryModel.Category, err error) {
	categoryDao := categoryModel.NewDao()
	categoryList, err := categoryDao.GetListByFather(&father)
	if err != nil {
		return sequenceCategory, errors.Wrap(err, "categoryServer.GetSequenceCategory")
	}
	if len(categoryList) == 0 {
		return []categoryModel.Category{}, nil
	}
	tree := make(map[uint][]categoryModel.Category)
	for _, category := range categoryList {
		if _, ok := tree[category.Previous]; !ok {
			tree[category.Previous] = []categoryModel.Category{category}
		} else {
			tree[category.Previous] = append(tree[category.Previous], category)
		}
	}

	sequenceCategory = make([]categoryModel.Category, 0, len(categoryList))
	catSvc.makeSequenceOfCategory(&sequenceCategory, tree, 0)
	return sequenceCategory, nil
}

func (catSvc *Category) makeSequenceOfCategory(
	queue *[]categoryModel.Category, tree map[uint][]categoryModel.Category, id uint,
) {
	if categoryList, exist := tree[id]; exist {
		for _, child := range categoryList {
			*queue = append(*queue, child)
			catSvc.makeSequenceOfCategory(queue, tree, child.ID)
		}
	}
}

func (catSvc *Category) GetSequenceFather(
	account accountModel.Account, incomeExpense *constant.IncomeExpense,
) ([]categoryModel.Father, error) {
	rows, err := categoryModel.NewDao().GetFatherList(account, incomeExpense)
	if err != nil {
		return []categoryModel.Father{}, err
	}
	var tree = make(map[uint][]categoryModel.Father, len(rows))
	for _, father := range rows {
		tree[father.Previous] = append(tree[father.Previous], father)
	}
	var result = []categoryModel.Father{}
	catSvc.makeSequenceOfFather(&result, tree, 0)
	return result, nil
}

func (catSvc *Category) makeSequenceOfFather(
	queue *[]categoryModel.Father, tree map[uint][]categoryModel.Father, treeKey uint,
) {
	if children, exist := tree[treeKey]; exist {
		for _, child := range children {
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
