package categoryService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataTool"
	"context"
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

func (catSvc *Category) NewCategoryData(name, icon string) CreateData {
	return CreateData{Name: name, Icon: icon}
}

func (catSvc *Category) CreateOne(father categoryModel.Father, data CreateData, ctx context.Context) (categoryModel.Category, error) {
	category := categoryModel.Category{
		AccountId:      father.AccountId,
		FatherId:       father.ID,
		IncomeExpense:  father.IncomeExpense,
		Name:           data.Name,
		Icon:           data.Icon,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	return category, db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := db.Get(ctx)
		if err := category.CheckName(tx); err != nil {
			return err
		}
		err := tx.Create(&category).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			//存在重复名称 则尝试恢复已软删除的交易类型
			err = tx.Where("account_id = ? AND name = ? AND deleted_at IS NOT NULL", category.AccountId, category.Name).First(&category).Error
			if err == nil {
				err = tx.Model(&category).Update("deleted_at", nil).Error
				if err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				return global.ErrCategorySameName
			} else {
				return err
			}
		} else if err != nil {
			return errors.Wrap(err, "category.CreateOne()")
		}
		return db.AddCommitCallback(ctx, func() {
			nats.PublishTaskWithPayload[categoryModel.Category](nats.TaskUpdateCategoryMapping, category)
		})
	})

}

func (catSvc *Category) UpdateCategoryMapping(category categoryModel.Category, ctx context.Context) error {
	if !aiService.IsOpen() {
		return nil
	}
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := db.Get(ctx)
		accountDao, categoryDao := accountModel.NewDao(tx), categoryModel.NewDao(tx)
		accountMapping, err := accountDao.SelectMultipleMapping(*accountModel.NewMappingCondition().WithRelatedId(category.AccountId))
		if err != nil {
			return err
		}
		var mainAccount accountModel.Account
		for _, mapping := range accountMapping {
			mainAccount, err = accountDao.SelectById(mapping.MainId)
			if err != nil {
				return err
			}
			var mainCategoryList dataTool.Slice[string, categoryModel.Category]
			mainCategoryList, err = categoryDao.GetListByAccount(mainAccount, &category.IncomeExpense)
			if err != nil {
				return err
			}
			mainNameList := mainCategoryList.ExtractValues(func(category categoryModel.Category) string {
				return category.Name
			})
			// 匹配交易类型
			var target string
			target, err = aiService.ChineseSimilarityMatching(category.Name, mainNameList, ctx)
			if err != nil {
				return err
			}
			for _, mainCategory := range mainCategoryList {
				if target == mainCategory.Name {
					_, err = categoryDao.CreateMapping(mainCategory, category)
					if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
						return err
					}
					break
				}
			}
		}
		return nil
	})
}

func (catSvc *Category) CreateList(
	father categoryModel.Father, list []CreateData, ctx context.Context,
) ([]categoryModel.Category, error) {
	if len(list) == 0 {
		return nil, nil
	}
	categoryList := make([]categoryModel.Category, len(list), len(list))
	for i, data := range list {
		categoryList[i] = categoryModel.Category{
			AccountId:      father.AccountId,
			FatherId:       father.ID,
			IncomeExpense:  father.IncomeExpense,
			Name:           data.Name,
			Icon:           data.Icon,
			Previous:       0,
			OrderUpdatedAt: time.Now(),
		}

	}
	err := db.Transaction(ctx, func(ctx *cus.TxContext) error {
		return ctx.GetDb().Create(&categoryList).Error
	})
	if err != nil {
		err = errors.Wrap(err, "category.CreateOne()")
	}
	return categoryList, err
}

func (catSvc *Category) CreateOneFather(
	account accountModel.Account, InEx constant.IncomeExpense, name string, ctx context.Context,
) (categoryModel.Father, error) {
	father := categoryModel.Father{
		AccountId:      account.ID,
		IncomeExpense:  InEx,
		Name:           name,
		Previous:       0,
		OrderUpdatedAt: time.Now(),
	}
	return father, db.Transaction(ctx, func(ctx *cus.TxContext) error {
		err := ctx.GetDb().Create(&father).Error
		if err != nil {
			return errors.Wrap(err, "father.CreateOne()")
		}
		return nil
	})
}

func (catSvc *Category) MoveCategory(
	category categoryModel.Category, previous *categoryModel.Category, father *categoryModel.Father,
	operator userModel.User, ctx context.Context,
) error {
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
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
		return err
	})
}

func (catSvc *Category) MoveFather(father categoryModel.Father, previous *categoryModel.Father, ctx context.Context) error {
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
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
	})
}

// GetSequenceCategoryByFather 返回排序后的category
func (catSvc *Category) GetSequenceCategoryByFather(father categoryModel.Father) (categoryList []categoryModel.Category, err error) {
	categoryDao := categoryModel.NewDao()
	categoryList, err = categoryDao.GetListByFather(father)
	if err != nil {
		return categoryList, errors.Wrap(err, "categoryServer.GetSequenceCategory")
	}
	categoryDao.Order(categoryList)
	return
}

func (catSvc *Category) GetSequenceFather(
	account accountModel.Account, incomeExpense *constant.IncomeExpense,
) (list []categoryModel.Father, err error) {
	dao := categoryModel.NewDao()
	list, err = dao.GetFatherList(account, incomeExpense)
	if err != nil {
		return
	}
	dao.OrderFather(list)
	return
}

func (catSvc *Category) Update(categoryId uint, data categoryModel.CategoryUpdateData, ctx context.Context) (category categoryModel.Category, err error) {
	err = db.Transaction(ctx, func(ctx *cus.TxContext) error {
		dao := categoryModel.NewDao(ctx.GetDb())
		err = dao.Update(categoryId, data)
		if err != nil {
			return err
		}
		category, err = dao.SelectById(categoryId)
		return err
	})
	return
}

func (catSvc *Category) UpdateFather(father categoryModel.Father, name string) (categoryModel.Father, error) {
	if name == "" {
		return father, global.ErrCategoryNameEmpty
	}
	err := global.GvaDb.Model(&father).Update("name", name).Error
	return father, err
}

func (catSvc *Category) Delete(category categoryModel.Category, ctx context.Context) error {
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		exits, err := catSvc.existTransaction(category)
		if err != nil {
			return err
		}
		if exits {
			return errors.Wrap(ErrExistTransaction, "已存在交易不可删除")
		}
		return ctx.GetDb().Delete(&category).Error
	})
}

func (catSvc *Category) DeleteFather(father categoryModel.Father, ctx context.Context) error {
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
		var categoryList []categoryModel.Category
		err := global.GvaDb.Select("id").Where("father_id = ?", father.ID).Find(&categoryList).Error
		if err != nil {
			return errors.Wrap(err, "")
		}
		exits, err := catSvc.existTransaction(categoryList...)
		if err != nil {
			return err
		} else if exits {
			return errors.Wrap(ErrExistTransaction, "delete category")
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
	})
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

func (catSvc *Category) MappingCategory(parent, child categoryModel.Category, operator userModel.User, ctx context.Context) (mapping categoryModel.Mapping, err error) {
	err = db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
		err = catSvc.checkMappingParam(parent, child, operator, tx)
		if err != nil {
			return err
		}
		mapping, err = categoryModel.NewDao(tx).CreateMapping(parent, child)
		return err
	})
	return
}

func (catSvc *Category) DeleteMapping(parent, child categoryModel.Category, operator userModel.User, ctx context.Context) error {
	return db.Transaction(ctx, func(ctx *cus.TxContext) error {
		tx := ctx.GetDb()
		err := catSvc.checkMappingParam(parent, child, operator, tx)
		if err != nil {
			return err
		}
		return tx.Where(
			"parent_category_id = ? AND child_category_id = ?", parent.ID, child.ID,
		).Delete(&categoryModel.Mapping{}).Error
	})
}

func (catSvc *Category) MappingCategoryToAccountMapping(mappingAccount accountModel.Mapping, ctx context.Context) error {
	tx := db.Get(ctx)
	main, err := mappingAccount.GetMainAccount(tx)
	if err != nil {
		return err
	}
	related, err := mappingAccount.GetRelatedAccount(tx)
	if err != nil {
		return err
	}
	return catSvc.mappingAccountCategoryByAI(main, related, ctx)
}

func (catSvc *Category) mappingAccountCategoryByAI(mainAccount, mappingAccount accountModel.Account, ctx context.Context) error {
	tx := db.Get(ctx)
	if false == aiService.IsOpen() {
		return nil
	}
	var mainCategoryList, mappingCategoryList dataTool.Slice[string, categoryModel.Category]
	var err error
	var matchingResult map[string]string
	categoryDao := categoryModel.NewDao(tx)
	for _, ie := range []constant.IncomeExpense{constant.Income, constant.Expense} {
		//查询交易类型
		mainCategoryList, err = categoryDao.GetListByAccount(mainAccount, &ie)
		if err != nil {
			return err
		}
		mappingCategoryList, err = categoryDao.GetUnmappedList(mainAccount, mappingAccount, &ie)
		if err != nil {
			return err
		}
		if len(mainCategoryList) == 0 || len(mappingCategoryList) == 0 {
			continue
		}
		//转数据格式
		mainNameList := mainCategoryList.ExtractValues(func(category categoryModel.Category) string {
			return category.Name
		})
		mappingNameList := mainCategoryList.ExtractValues(func(category categoryModel.Category) string {
			return category.Name
		})
		//获得相似度匹配
		matchingResult, err = aiService.BatchChineseSimilarityMatching(mappingNameList, mainNameList, ctx)
		if err != nil {
			return err
		}
		mainNameMap := mainCategoryList.ToMap(func(category categoryModel.Category) string {
			return category.Name
		})
		mappingNameMap := mappingCategoryList.ToMap(func(category categoryModel.Category) string {
			return category.Name
		})
		for mappingName, mainName := range matchingResult {
			_, err = categoryDao.CreateMapping(mainNameMap[mainName], mappingNameMap[mappingName])
			if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
				return err
			}
		}
	}
	return err
}
