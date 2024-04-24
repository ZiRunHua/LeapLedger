package categoryModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	"KeepAccount/util"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CategoryDao struct {
	db *gorm.DB
}

func NewDao(db ...*gorm.DB) *CategoryDao {
	if len(db) > 0 {
		return &CategoryDao{db: db[0]}
	}
	return &CategoryDao{global.GvaDb}
}

// Deprecated: 改用 categoryModel.NewDao
func (d *dao) NewCategory(db *gorm.DB) *CategoryDao {
	if db == nil {
		db = global.GvaDb
	}
	return &CategoryDao{db}
}

func (cd *CategoryDao) SelectById(id uint) (category Category, err error) {
	err = cd.db.First(&category, id).Error
	return
}

type CategoryUpdateData struct {
	Name *string
	Icon *string
}

func (cd *CategoryDao) Update(category Category, data CategoryUpdateData) error {
	updateData := &Category{}
	if err := util.Data.CopyNotEmptyStringOptional(data.Name, &updateData.Name); err != nil {
		return err
	}
	if err := util.Data.CopyNotEmptyStringOptional(data.Icon, &updateData.Icon); err != nil {
		return err
	}
	if updateData.Name != "" {
		if err := updateData.CheckName(cd.db); err != nil {
			return err
		}
	}
	return cd.db.Model(&category).Updates(updateData).Error
}

func (cd *CategoryDao) SelectFirstChild(categoryId uint) (Category, error) {
	var result Category
	query := cd.db.Where("previous = ?", categoryId)
	err := cd.setCategoryOrder(query).First(&result).Error
	return result, err
}

func (cd *CategoryDao) SelectFatherFirstChild(fatherId uint) (Father, error) {
	var result Father
	query := cd.db.Where("previous = ?", fatherId)
	err := query.Order("income_expense asc,previous asc,order_updated_at desc").First(&result).Error
	return result, err
}

func (cd *CategoryDao) UpdateChildPrevious(categoryId, newPrevious uint) error {
	return cd.db.Model(&Category{}).Where("previous = ?", categoryId).Update("previous", newPrevious).Error
}

func (cd *CategoryDao) UpdateFatherChildPrevious(categoryId, newPrevious uint) error {
	return cd.db.Model(&Father{}).Where("previous = ?", categoryId).Update("previous", newPrevious).Error
}

func (cd *CategoryDao) GetListByFather(father *Father) ([]Category, error) {
	var list []Category
	err := cd.setCategoryOrder(cd.db.Where("father_id = ?", father.ID)).Find(&list).Error
	return list, err
}

func (cd *CategoryDao) GetListByAccount(account accountModel.Account, ie *constant.IncomeExpense) ([]Category, error) {
	condition := &Condition{account: account, ie: ie}
	var list []Category
	return list, condition.buildWhere(cd.db).Order("income_expense asc,previous asc,order_updated_at desc").Find(&list).Error
}

func (cd *CategoryDao) GetUnmappedList(mainAccount, mappingAccount accountModel.Account, ie *constant.IncomeExpense) (list []Category, err error) {
	childSelect := cd.db.Model(&Mapping{}).Select("child_category_id")
	childSelect.Where("parent_account_id = ? AND child_account_id = ?", mainAccount.ID, mappingAccount.ID)
	err = cd.db.Where("account_id = ? AND income_expense = ? ", mappingAccount.ID, ie).Not("id IN (?)", childSelect).Find(&list).Error
	return
}

func (cd *CategoryDao) GetFatherList(account accountModel.Account, incomeExpense *constant.IncomeExpense) ([]Father, error) {
	condition := &Condition{account: account, ie: incomeExpense}
	var list []Father
	return list, condition.buildWhere(cd.db).Order("income_expense asc,previous asc,order_updated_at desc").Find(&list).Error
}

func (cd *CategoryDao) GetAll(account accountModel.Account, incomeExpense *constant.IncomeExpense) ([]Category, error) {
	condition := &Condition{account: account, ie: incomeExpense}
	var list []Category
	return list, cd.setCategoryOrder(condition.buildWhere(cd.db)).Find(&list).Error
}

func (cd *CategoryDao) Exist(account accountModel.Account) (bool, error) {
	category := &Category{}
	err := cd.db.Where("account_id = ?", account.ID).Take(category).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, errors.WithStack(err)
}

func (cd *CategoryDao) CreateMapping(parent, child Category) (Mapping, error) {
	mapping := Mapping{
		ParentAccountId:  parent.AccountId,
		ChildAccountId:   child.AccountId,
		ParentCategoryId: parent.ID,
		ChildCategoryId:  child.ID,
	}
	err := cd.db.Create(&mapping).Error
	return mapping, err
}

func (cd *CategoryDao) SelectMapping(parentAccountId, childCategoryId uint) (Mapping, error) {
	var result Mapping
	err := cd.db.Where("parent_account_id = ? AND child_category_id = ?", parentAccountId, childCategoryId).First(&result).Error
	return result, err
}

func (cd *CategoryDao) SelectMappingByCAccountIdAndPCategoryId(childAccountId, parentCategoryId uint) (Mapping, error) {
	var result Mapping
	err := cd.db.Where("child_account_id = ? AND parent_category_id = ?", childAccountId, parentCategoryId).First(&result).Error
	return result, err
}

func (cd *CategoryDao) GetMappingByAccountMappingOrderByChildCategoryWeight(parentAccountId, childAccountId uint) (
	[]Mapping, error,
) {
	query := cd.db.Where(
		"category_mapping.parent_account_id = ? AND category_mapping.child_account_id = ?", parentAccountId,
		childAccountId,
	)
	query = query.Joins("LEFT JOIN category ON category_mapping.child_category_id = category.id")
	var list []Mapping
	err := cd.setCategoryOrder(query).Select("category_mapping.*").Find(&list).Error
	return list, err
}

func (cd *CategoryDao) setCategoryOrder(db *gorm.DB) *gorm.DB {
	return db.Order("category.income_expense asc,category.previous asc,category.order_updated_at desc")
}
