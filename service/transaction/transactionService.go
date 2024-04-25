package transactionService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Transaction struct{}

type CreateOption struct {
	syncUpdateStatistic bool
	syncShareAccount    bool
}

func (txnService *Transaction) CreateOne(
	trans transactionModel.Transaction, accountUser accountModel.User, option CreateOption, tx *gorm.DB,
) (transactionModel.Transaction, error) {
	// check
	err := txnService.checkTransaction(trans, accountUser, tx)
	if err != nil {
		return trans, errors.WithStack(err)
	}
	err = accountUser.CheckTransAddByUserId(trans.UserId)
	if err != nil {
		return trans, errors.WithStack(err)
	}
	// handle
	trans.UserId = accountUser.UserId
	err = tx.Create(&trans).Error
	if err != nil {
		return trans, errors.WithStack(err)
	}

	if false == option.syncUpdateStatistic {
		err = txnService.asyncUpdateStatistic(trans.GetStatisticData(true), tx)
	} else {
		err = txnService.updateStatistic(trans.GetStatisticData(true), tx)
	}
	if err != nil {
		return trans, err
	}
	if false == option.syncShareAccount {
		err = tx.Transaction(func(tx *gorm.DB) error {
			return txnService.SyncShareAccount(trans, tx)
		})
		if err != nil {
			return trans, err
		}
	}
	return trans, nil
}

func (txnService *Transaction) NewDefaultCreateConfig() CreateOption {
	return CreateOption{syncUpdateStatistic: true, syncShareAccount: true}
}

func (txnService *Transaction) NewCreateConfig() *CreateOption {
	return &CreateOption{}
}

func (co *CreateOption) WithSyncUpdateStatistic(val bool) *CreateOption {
	co.syncUpdateStatistic = val
	return co
}

func (co *CreateOption) WithSyncShareAccount(val bool) *CreateOption {
	co.syncShareAccount = val
	return co
}

func (txnService *Transaction) SyncShareAccount(trans transactionModel.Transaction, tx *gorm.DB) error {
	if err := trans.ForUpdate(tx); err != nil {
		return err
	}
	accountDao, categoryDao := accountModel.NewDao(tx), categoryModel.NewDao(tx)
	var categoryMapping categoryModel.Mapping
	var transMapping transactionModel.Mapping
	syncTrans := trans.SyncDataClone()

	accountUser, err := accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
	if err != nil {
		return err
	}
	accountType, err := accountDao.GetAccountType(trans.AccountId)
	if err != nil {
		return err
	}
	switch accountType {
	case accountModel.TypeIndependent:
		// 独立账本向共享同步
		var accountMappings []accountModel.Mapping
		accountMappings, err = accountDao.SelectMultipleMapping(*accountModel.NewMappingCondition().WithRelatedId(trans.AccountId))
		if err != nil {
			return err
		}

		for _, accountMapping := range accountMappings {
			categoryMapping, err = categoryDao.SelectMapping(accountMapping.MainId, trans.CategoryId)
			if errors.As(err, gorm.ErrRecordNotFound) {
				continue
			} else if err != nil {
				return err
			}

			syncTrans.AccountId = categoryMapping.ParentAccountId
			syncTrans.CategoryId = categoryMapping.ParentCategoryId
			err = global.GvaDb.Transaction(func(tx *gorm.DB) error {
				transMapping, err = transactionModel.NewDao(tx).SelectMappingByTrans(trans, syncTrans)
				if err == nil {
					_, err = txnService.CreateSyncTrans(trans, syncTrans, tx)
				} else if errors.As(err, gorm.ErrRecordNotFound) {
					syncTrans.ID = transMapping.RelatedId
					err = txnService.Update(syncTrans, accountUser, tx)
				}
				return err
			})
			if err != nil && false == errors.As(err, gorm.ErrDuplicatedKey) {
				return err
			}
		}
	case accountModel.TypeShare:
		// 共享同步向独立账本
		var accountMapping accountModel.Mapping
		accountMapping, err = accountDao.SelectMappingByMainAccountAndRelatedUser(trans.AccountId, trans.UserId)
		if err != nil {
			return err
		}

		categoryMapping, err = categoryDao.SelectMappingByCAccountIdAndPCategoryId(accountMapping.RelatedId, trans.CategoryId)
		if errors.As(err, gorm.ErrRecordNotFound) {
			return nil
		} else if err != nil {
			return err
		}
		syncTrans.AccountId = categoryMapping.ChildAccountId
		syncTrans.CategoryId = categoryMapping.ChildCategoryId
		err = global.GvaDb.Transaction(func(tx *gorm.DB) error {
			transMapping, err = transactionModel.NewDao(tx).SelectMappingByTrans(trans, syncTrans)
			if err == nil {
				_, err = txnService.CreateSyncTrans(trans, syncTrans, tx)
			} else if errors.As(err, gorm.ErrRecordNotFound) {
				syncTrans.ID = transMapping.MainId
				err = txnService.Update(syncTrans, accountUser, tx)
			}
			return err
		})
	default:
		panic("error account.type")
	}
	return nil
}

func (txnService *Transaction) CreateSyncTrans(trans, syncTrans transactionModel.Transaction, tx *gorm.DB) (mapping transactionModel.Mapping, err error) {
	accountUser, err := accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
	if err != nil {
		return
	}
	newTrans, err := txnService.CreateOne(syncTrans, accountUser, txnService.NewDefaultCreateConfig(), tx)
	if err != nil {
		return
	}

	accountType, err := accountModel.NewDao(tx).GetAccountType(trans.AccountId)
	if err != nil {
		return
	}
	switch accountType {
	case accountModel.TypeIndependent:
		mapping, err = txnService.CreateMapping(newTrans, trans, tx)
	case accountModel.TypeShare:
		mapping, err = txnService.CreateMapping(trans, newTrans, tx)
	default:
		panic("err account.type")
	}
	if err != nil {
		return
	}
	return
}

func (txnService *Transaction) CreateMapping(mainTrans, relatedTrans transactionModel.Transaction, tx *gorm.DB) (transactionModel.Mapping, error) {
	mapping := transactionModel.Mapping{MainId: mainTrans.ID, RelatedId: relatedTrans.ID}
	err := tx.Create(&mapping).Error
	return mapping, err
}

func (txnService *Transaction) asyncUpdateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	if nats.Publish[transactionModel.StatisticData](nats.TaskStatisticUpdate, data) {
		return nil
	}
	// 添加异步失败直接执行
	return txnService.updateStatistic(data, tx)
}

func (txnService *Transaction) updateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	switch data.IncomeExpense {
	case constant.Income:
		if err := transactionModel.IncomeAccumulate(
			data.TradeTime, data.AccountId, data.UserId, data.CategoryId, data.Amount, data.Count, tx,
		); err != nil {
			return errors.Wrap(err, "transactionModel.IncomeAccumulate")
		}
	case constant.Expense:
		if err := transactionModel.ExpenseAccumulate(
			data.TradeTime, data.AccountId, data.UserId, data.CategoryId, data.Amount, data.Count, tx,
		); err != nil {
			return errors.Wrap(err, "transactionModel.ExpenseAccumulate")
		}
	default:
		panic("income Expense error")
	}
	return nil
}

func (txnService *Transaction) checkTransaction(trans transactionModel.Transaction, accountUser accountModel.User, tx *gorm.DB) error {
	category, err := trans.GetCategory(tx)
	if err != nil {
		return err
	}
	if category.AccountId != trans.AccountId || trans.AccountId != accountUser.AccountId {
		return global.ErrAccountId
	}
	if trans.Amount < 0 {
		return errors.New("error trans.amount")
	}
	return nil
}

// Update only "user_id,income_expense,category_id,amount,remark,trade_time" can be changed
func (txnService *Transaction) Update(
	transaction transactionModel.Transaction, accountUser accountModel.User, tx *gorm.DB,
) error {
	// check
	err := txnService.checkTransaction(transaction, accountUser, tx)
	if err != nil {
		return errors.WithStack(err)
	}
	err = accountUser.CheckTransEditByUserId(transaction.UserId)
	if err != nil {
		return errors.WithStack(err)
	}
	// handle
	var oldTransaction transactionModel.Transaction
	oldTransaction, err = transactionModel.NewDao(tx).SelectById(transaction.ID, true)
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Select("user_id,income_expense,category_id,amount,remark,trade_time").Updates(transaction).Error
	if err != nil {
		return errors.WithStack(err)
	}

	err = txnService.updateStatisticAfterUpdate(oldTransaction, transaction, tx)
	if err != nil {
		return err
	}
	return txnService.SyncShareAccount(transaction, tx)
}

func (txnService *Transaction) updateStatisticAfterUpdate(
	oldTxn transactionModel.Transaction, txn transactionModel.Transaction, tx *gorm.DB,
) error {
	updateStatisticData := oldTxn.GetStatisticData(false)
	if err := txnService.asyncUpdateStatistic(updateStatisticData, tx); err != nil {
		return err
	}
	if err := txnService.asyncUpdateStatistic(txn.GetStatisticData(true), tx); err != nil {
		return err
	}
	return nil
}

func (txnService *Transaction) Delete(
	txn transactionModel.Transaction, accountUser accountModel.User, tx *gorm.DB,
) error {
	err := accountUser.CheckTransEditByUserId(txn.UserId)
	if err != nil {
		return err
	}
	err = txnService.updateStatisticAfterDelete(txn, tx)
	if err != nil {
		return err
	}
	return tx.Delete(&txn).Error
}

func (txnService *Transaction) updateStatisticAfterDelete(txn transactionModel.Transaction, tx *gorm.DB) error {
	updateStatisticData := txn.GetStatisticData(false)
	return txnService.asyncUpdateStatistic(updateStatisticData, tx)
}

func (txnService *Transaction) CreateMultiple(
	accountUser accountModel.User, account accountModel.Account, transactionList []transactionModel.Transaction,
	tx *gorm.DB,
) (failTransList []*transactionModel.Transaction, err error) {
	if account.ID != accountUser.AccountId {
		err = global.ErrAccountId
		return
	}
	err = accountUser.CheckTransAddByUserId(accountUser.UserId)
	if err != nil {
		return
	}

	var categoryIds []uint
	if err = global.GvaDb.Model(&categoryModel.Category{}).Where("account_id = ?", account.ID).Pluck(
		"id", &categoryIds,
	).Error; err != nil {
		return nil, err
	}
	categoryIdMap := make(map[uint]bool)
	for _, id := range categoryIds {
		categoryIdMap[id] = true
	}

	incomeAmount, expenseAmount := make(map[string]map[uint]int), make(map[string]map[uint]int)
	incomeCount, expenseCount := make(map[string]map[uint]int), make(map[string]map[uint]int)

	var incomeTransList, expenseTransList []*transactionModel.Transaction
	var key string
	for index := range transactionList {
		transactionList[index].UserId = accountUser.UserId
		transaction := transactionList[index]
		if !categoryIdMap[transaction.CategoryId] {
			failTransList = append(failTransList, &transaction)
			continue
		}
		if transaction.IncomeExpense == constant.Income {
			incomeTransList = append(incomeTransList, &transaction)
			key = transaction.TradeTime.Format("2006-01-02")
			if incomeAmount[key] == nil {
				incomeAmount[key] = map[uint]int{transaction.CategoryId: transaction.Amount}
				incomeCount[key] = map[uint]int{transaction.CategoryId: 1}
			} else {
				incomeAmount[key][transaction.CategoryId] += transaction.Amount
				incomeCount[key][transaction.CategoryId]++
			}
		} else if transaction.IncomeExpense == constant.Expense {
			expenseTransList = append(expenseTransList, &transaction)
			key = transaction.TradeTime.Format("2006-01-02")
			if expenseAmount[key] == nil {
				expenseAmount[key] = map[uint]int{transaction.CategoryId: transaction.Amount}
				expenseCount[key] = map[uint]int{transaction.CategoryId: 1}
			} else {
				expenseAmount[key][transaction.CategoryId] += transaction.Amount
				expenseCount[key][transaction.CategoryId]++
			}
		} else {
			failTransList = append(failTransList, &transaction)
			continue
		}
	}
	var transaction transactionModel.Transaction
	if len(incomeTransList) > 0 {
		if err = tx.Model(&transaction).Create(incomeTransList).Error; err != nil {
			return nil, err
		}

		if err = txnService.addStatisticAfterCreateMultiple(
			account, accountUser, constant.Income, incomeAmount, incomeCount, tx,
		); err != nil {
			return nil, err
		}
	}
	if len(expenseTransList) > 0 {
		if err = tx.Model(&transaction).Create(expenseTransList).Error; err != nil {
			return nil, err
		}
		if err = txnService.addStatisticAfterCreateMultiple(
			account, accountUser, constant.Expense, expenseAmount, expenseCount, tx,
		); err != nil {
			return nil, err
		}
	}
	return failTransList, err
}

func (txnService *Transaction) addStatisticAfterCreateMultiple(
	account accountModel.Account, accountUser accountModel.User, incomeExpense constant.IncomeExpense,
	amountList map[string]map[uint]int, countList map[string]map[uint]int, tx *gorm.DB,
) error {
	var err error
	var tradeTime time.Time
	for date, categoryList := range amountList {
		if tradeTime, err = time.Parse("2006-01-02", date); err != nil {
			return err
		}
		for categoryId, amount := range categoryList {
			if err = txnService.updateStatistic(transactionModel.StatisticData{
				AccountId: account.ID, UserId: accountUser.UserId, IncomeExpense: incomeExpense, CategoryId: categoryId,
				TradeTime: tradeTime, Amount: amount, Count: countList[date][categoryId],
			}, tx); err != nil {
				return err
			}
		}
	}
	return nil
}
