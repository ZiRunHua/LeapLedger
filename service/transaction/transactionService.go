package transactionService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Transaction struct{}

func (txnService *Transaction) Create(
	transInfo transactionModel.Info, accountUser accountModel.User, option Option, ctx context.Context,
) (transactionModel.Transaction, error) {
	trans := transactionModel.Transaction{Info: transInfo}
	err := db.Transaction(ctx, func(ctx *cusCtx.TxContext) error {
		tx := ctx.GetDb()
		// check
		if transInfo.AccountId != accountUser.AccountId {
			return global.ErrAccountId
		}
		err := transInfo.Check(tx)
		if err != nil {
			return errors.WithStack(err)
		}
		err = accountUser.CheckTransAddByUserId(transInfo.UserId)
		if err != nil {
			return errors.WithStack(err)
		}
		// handle
		transInfo.UserId = accountUser.UserId
		err = tx.Create(&trans).Error
		if err != nil {
			return errors.WithStack(err)
		}
		// other
		if option.syncUpdateStatistic {
			err = txnService.updateStatistic(transInfo.GetStatisticData(true), tx)
		}
		return nil
	})

	if err != nil {
		return trans, err
	}
	return trans, txnService.onCreateSuccess(trans, option, ctx)
}

func (txnService *Transaction) onCreateSuccess(trans transactionModel.Transaction, option Option, ctx context.Context) error {
	var err error

	if option.transSyncToMappingAccount {
		err = db.AddCommitCallback(ctx, func() {
			err = task.syncToMappingAccount(trans, ctx)
			if err != nil {
				errorLog.Error("onCreateSuccess=>syncToMappingAccount", zap.Error(err))
			}
		})
		if err != nil {
			errorLog.Error("onCreateSuccess=>syncToMappingAccount", zap.Error(err))
		}
	}

	if false == option.syncUpdateStatistic {
		err = db.AddCommitCallback(ctx, func() {
			err = task.updateStatistic(trans.GetStatisticData(true), db.Get(ctx))
			if err != nil {
				errorLog.Error("onCreateSuccess=>updateStatistic", zap.Error(err))
			}
		})
		if err != nil {
			errorLog.Error("onCreateSuccess=>updateStatistic", zap.Error(err))
		}
	}

	return nil
}

// Update only "user_id,income_expense,category_id,amount,remark,trade_time" can be changed
func (txnService *Transaction) Update(
	trans transactionModel.Transaction, accountUser accountModel.User, option Option, ctx context.Context,
) error {
	var oldTrans transactionModel.Transaction
	err := db.Transaction(ctx, func(ctx *cusCtx.TxContext) error {
		tx := ctx.GetDb()
		// check
		err := txnService.checkTransaction(trans, accountUser, tx)
		if err != nil {
			return errors.WithStack(err)
		}
		err = accountUser.CheckTransEditByUserId(trans.UserId)
		if err != nil {
			return errors.WithStack(err)
		}
		// handle
		oldTrans = trans
		if err = oldTrans.ForShare(tx); err != nil {
			return errors.WithStack(err)
		}
		if oldTrans.UpdatedAt.Add(time.Second * 3).After(time.Now()) {
			return errors.WithStack(global.ErrFrequentOperation)
		}
		err = tx.Select("income_expense", "category_id", "amount", "remark", "trade_time").Updates(trans).Error
		if err != nil {
			return errors.WithStack(err)
		}
		// other
		if option.syncUpdateStatistic {
			if err = txnService.updateStatistic(oldTrans.GetStatisticData(false), tx); err != nil {
				return err
			}
			if err = txnService.updateStatistic(trans.GetStatisticData(true), tx); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return txnService.onUpdateSuccess(oldTrans, trans, option, ctx)
}

func (txnService *Transaction) onUpdateSuccess(
	_, trans transactionModel.Transaction, option Option, ctx context.Context,
) error {
	var err error
	if option.transSyncToMappingAccount {
		err = task.syncToMappingAccount(trans, ctx)
		if err != nil {
			errorLog.Error("onUpdateSuccess=>syncToMappingAccount", zap.Error(err))
		}
	}

	if false == option.syncUpdateStatistic {
		err = task.updateStatistic(trans.GetStatisticData(true), db.Get(ctx))
		if err != nil {
			errorLog.Error("onUpdateSuccess=>updateStatistic", zap.Error(err))
		}
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
	return task.updateStatistic(updateStatisticData, tx)
}

func (txnService *Transaction) checkTransaction(trans transactionModel.Transaction, accountUser accountModel.User, tx *gorm.DB) error {
	if trans.AccountId != accountUser.AccountId {
		return global.ErrAccountId
	}
	err := trans.Check(tx)
	if err != nil {
		return err
	}
	return nil
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

func (txnService *Transaction) SyncToMappingAccount(trans transactionModel.Transaction, ctx context.Context) error {
	tx := db.Get(ctx)
	accountType, err := accountModel.NewDao(tx).GetAccountType(trans.AccountId)
	if err != nil {
		return errors.WithMessage(err, "同步交易失败")
	}
	if accountType == accountModel.TypeIndependent {
		err = txnService.syncToShareAccount(trans, ctx)
	} else if accountType == accountModel.TypeShare {
		err = txnService.syncToIndependentAccount(trans, ctx)
	} else {
		return accountModel.ErrAccountType
	}
	return err
}

func (txnService *Transaction) syncToShareAccount(indAccountTrans transactionModel.Transaction, ctx context.Context) error {
	tx := db.Get(ctx)
	accountDao, categoryDao := accountModel.NewDao(tx), categoryModel.NewDao(tx)

	accountMappings, err := accountDao.SelectMultipleMapping(*accountModel.NewMappingCondition().WithRelatedId(indAccountTrans.AccountId))
	if err != nil {
		return err
	}

	categoryMapping, transMapping, syncTrans := categoryModel.Mapping{}, transactionModel.Mapping{}, indAccountTrans.SyncDataClone()

	for _, accountMapping := range accountMappings {
		categoryMapping, err = categoryDao.SelectMapping(accountMapping.MainId, indAccountTrans.CategoryId)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		} else if err != nil {
			return err
		}

		syncTrans.AccountId = categoryMapping.ParentAccountId
		syncTrans.CategoryId = categoryMapping.ParentCategoryId
		transMapping, err = transactionModel.NewDao(tx).SelectMappingByTrans(indAccountTrans, syncTrans)
		if err == nil {
			err = transMapping.ForShare(tx)
			if err != nil {
				return err
			}
			if false == transMapping.CanSyncTrans(indAccountTrans) {
				return nil
			}

			syncTrans.ID = transMapping.RelatedId
			var accountUser accountModel.User
			accountUser, err = accountDao.SelectUser(syncTrans.AccountId, syncTrans.UserId)
			if err != nil {
				return err
			}
			option := txnService.NewDefaultOption()
			option.WithTransSyncToMappingAccount(false)
			err = txnService.Update(syncTrans, accountUser, option, context.WithValue(ctx, cusCtx.Db, tx))
			if err != nil {
				return err
			}
			err = transMapping.OnSyncSuccess(tx, indAccountTrans)
			if err != nil {
				return err
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			_, err = txnService.CreateSyncTrans(indAccountTrans, syncTrans, ctx)
			if err != nil && false == errors.Is(err, gorm.ErrDuplicatedKey) {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func (txnService *Transaction) syncToIndependentAccount(shareAccountTrans transactionModel.Transaction, ctx context.Context) error {
	tx := db.Get(ctx)
	var err error
	accountMapping, err := accountModel.NewDao(tx).SelectMappingByMainAccountAndRelatedUser(shareAccountTrans.AccountId, shareAccountTrans.UserId)
	if err != nil {
		return err
	}
	categoryMapping, err := categoryModel.NewDao(tx).SelectMappingByCAccountIdAndPCategoryId(accountMapping.RelatedId, shareAccountTrans.CategoryId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	syncTrans := shareAccountTrans.SyncDataClone()
	syncTrans.AccountId = categoryMapping.ChildAccountId
	syncTrans.CategoryId = categoryMapping.ChildCategoryId
	transMapping, err := transactionModel.NewDao(tx).SelectMappingByTrans(shareAccountTrans, syncTrans)
	if err == nil {
		err = transMapping.ForShare(tx)
		if err != nil {
			return err
		}
		if false == transMapping.CanSyncTrans(shareAccountTrans) {
			return nil
		}

		syncTrans.ID = transMapping.MainId
		var accountUser accountModel.User
		accountUser, err = accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
		if err != nil {
			return err
		}
		option := txnService.NewDefaultOption()
		option.WithTransSyncToMappingAccount(false)
		err = txnService.Update(syncTrans, accountUser, option, ctx)
		if err != nil {
			return err
		}
		err = transMapping.OnSyncSuccess(tx, shareAccountTrans)
		if err != nil {
			return err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		_, err = txnService.CreateSyncTrans(shareAccountTrans, syncTrans, ctx)
		if err != nil && false == errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
	} else {
		return err
	}

	return nil
}

func (txnService *Transaction) CreateSyncTrans(trans, syncTrans transactionModel.Transaction, ctx context.Context) (mapping transactionModel.Mapping, err error) {
	tx := db.Get(ctx)
	accountUser, err := accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
	if err != nil {
		return
	}
	option := txnService.NewDefaultOption()
	option.WithTransSyncToMappingAccount(false)
	newTrans, err := txnService.Create(syncTrans.Info, accountUser, option, ctx)
	if err != nil {
		return
	}

	mapping, err = txnService.CreateMapping(trans, newTrans, tx)
	if err != nil {
		return
	}
	return
}

func (txnService *Transaction) CreateMapping(trans1, trans2 transactionModel.Transaction, tx *gorm.DB) (mapping transactionModel.Mapping, err error) {
	accountType, err := accountModel.NewDao(tx).GetAccountType(trans1.AccountId)
	if err != nil {
		return
	}
	var mainTrans, relatedTrans transactionModel.Transaction
	switch accountType {
	case accountModel.TypeIndependent:
		mainTrans, relatedTrans = trans1, trans2
	case accountModel.TypeShare:
		mainTrans, relatedTrans = trans2, trans1
	}
	mapping = transactionModel.Mapping{
		MainId:                    mainTrans.ID,
		MainAccountId:             mainTrans.AccountId,
		RelatedId:                 relatedTrans.ID,
		RelatedAccountId:          relatedTrans.AccountId,
		LastSyncedTransUpdateTime: mainTrans.UpdatedAt,
	}
	err = tx.Create(&mapping).Error
	return
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
	if err = tx.Model(&categoryModel.Category{}).Where("account_id = ?", account.ID).Pluck(
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
			if err = txnService.updateStatistic(
				transactionModel.StatisticData{
					AccountId: account.ID, UserId: accountUser.UserId, IncomeExpense: incomeExpense,
					CategoryId: categoryId,
					TradeTime:  tradeTime, Amount: amount, Count: countList[date][categoryId],
				}, tx,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

type Option struct {
	syncUpdateStatistic       bool // syncUpdateStatistic 同步/异步更新统计数据
	transSyncToMappingAccount bool // transSyncToMappingAccount 交易数据至同步关联账本
}

func (txnService *Transaction) NewDefaultOption() Option {
	return Option{syncUpdateStatistic: true, transSyncToMappingAccount: true}
}

func (txnService *Transaction) NewOption() Option {
	return Option{}
}

func (txnService *Transaction) NewOptionFormConfig(trans transactionModel.Info, ctx context.Context) (option Option, err error) {
	userConfig, err := accountModel.NewDao(db.Get(ctx)).SelectUserConfig(trans.AccountId, trans.UserId)
	if err != nil {
		return
	}
	option = txnService.NewDefaultOption()
	option.transSyncToMappingAccount = userConfig.GetFlagStatus(accountModel.Flag_Trans_Sync_Mapping_Account)
	return
}

func (o *Option) WithSyncUpdateStatistic(val bool) *Option {
	o.syncUpdateStatistic = val
	return o
}

func (o *Option) WithTransSyncToMappingAccount(val bool) *Option {
	o.transSyncToMappingAccount = val
	return o
}

type TransTask func(trans transactionModel.Transaction, ctx context.Context) error

// handelTransTasks
func _(taskList []TransTask, trans transactionModel.Transaction, ctx context.Context) error {
	if len(taskList) > 2 {
		errGroup, _ := errgroup.WithContext(ctx)
		handFunc := func(task TransTask, trans transactionModel.Transaction, ctx context.Context) {
			errGroup.Go(func() error { return task(trans, ctx) })
		}
		for _, task := range taskList {
			handFunc(task, trans, ctx)
		}
		if err := errGroup.Wait(); err != nil {
			return err
		}
	} else {
		for _, task := range taskList {
			err := task(trans, ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
