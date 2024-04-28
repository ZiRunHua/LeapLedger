package transactionService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/contextKey"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Transaction struct{}

func (txnService *Transaction) Create(
	trans transactionModel.Transaction, accountUser accountModel.User, option Option, ctx context.Context,
) (transactionModel.Transaction, error) {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
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
	// other
	if false == option.syncUpdateStatistic {
		err = txnService.asyncUpdateStatistic(trans.GetStatisticData(true), tx)
	} else {
		err = txnService.updateStatistic(trans.GetStatisticData(true), tx)
	}
	if err != nil {
		return trans, err
	}
	return trans, txnService.onCreateSuccess(trans, option, ctx)
}

func (txnService *Transaction) onCreateSuccess(
	trans transactionModel.Transaction, option Option, ctx context.Context,
) error {
	var taskList []func(tx *gorm.DB) error
	if option.transSyncToMappingAccount {
		taskList = append(
			taskList, func(tx *gorm.DB) error {
				accountType, err := accountModel.NewDao(tx).GetAccountType(trans.AccountId)
				if err != nil {
					return errors.WithMessage(err, "同步交易失败")
				}
				if accountType == accountModel.TypeIndependent {
					err = txnService.SyncToShareAccount(trans, context.WithValue(ctx, contextKey.Tx, tx))
				} else {
					err = txnService.SyncToIndependentAccount(trans, context.WithValue(ctx, contextKey.Tx, tx))
				}
				if err != nil {
					return errors.WithMessage(err, "同步交易失败")
				}
				return nil
			},
		)
	}
	err := handelTaskList(taskList, ctx)
	if err != nil {
		errorLog.Error("onCreateSuccess", zap.Error(err))
	}
	return nil
}

// Update only "user_id,income_expense,category_id,amount,remark,trade_time" can be changed
func (txnService *Transaction) Update(
	trans transactionModel.Transaction, accountUser accountModel.User, option Option, ctx context.Context,
) error {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
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
	oldTrans := trans
	if err = oldTrans.ForUpdate(tx); err != nil {
		return errors.WithStack(err)
	}
	err = tx.Select("income_expense,category_id,amount,remark,trade_time").Updates(trans).Error
	if err != nil {
		return errors.WithStack(err)
	}
	// other
	if option.syncUpdateStatistic {
		if err = txnService.asyncUpdateStatistic(oldTrans.GetStatisticData(false), tx); err != nil {
			return err
		}
		if err = txnService.asyncUpdateStatistic(trans.GetStatisticData(true), tx); err != nil {
			return err
		}
	} else {
		if err = txnService.updateStatistic(oldTrans.GetStatisticData(false), tx); err != nil {
			return err
		}
		if err = txnService.updateStatistic(trans.GetStatisticData(true), tx); err != nil {
			return err
		}
	}

	return txnService.onUpdateSuccess(oldTrans, trans, option, ctx)
}

func (txnService *Transaction) onUpdateSuccess(
	oldTrans transactionModel.Transaction, trans transactionModel.Transaction, option Option, ctx context.Context,
) error {
	var taskList []func(tx *gorm.DB) error
	if option.transSyncToMappingAccount {
		taskList = append(
			taskList, func(tx *gorm.DB) error {
				accountType, err := accountModel.NewDao(tx).GetAccountType(trans.AccountId)
				if err != nil {
					return err
				}
				if accountType == accountModel.TypeIndependent {
					return txnService.SyncToShareAccount(trans, context.WithValue(ctx, contextKey.Tx, tx))
				} else {
					return txnService.SyncToIndependentAccount(trans, context.WithValue(ctx, contextKey.Tx, tx))
				}
			},
		)
	}
	err := handelTaskList(taskList, ctx)
	if err != nil {
		errorLog.Error("onUpdateSuccess", zap.Error(err))
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

func (txnService *Transaction) checkTransaction(
	trans transactionModel.Transaction, accountUser accountModel.User, tx *gorm.DB,
) error {
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

func (txnService *Transaction) asyncUpdateStatistic(data transactionModel.StatisticData, tx *gorm.DB) error {
	if nats.Publish[transactionModel.StatisticData](nats.TaskStatisticUpdate, data) {
		return nil
	}
	// 添加异步失败直接执行
	return txnService.updateStatistic(data, tx)
}

func (txnService *Transaction) SyncToShareAccount(
	indAccountTrans transactionModel.Transaction, ctx context.Context,
) error {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
	if err := indAccountTrans.ForUpdate(tx); err != nil {
		return err
	}
	accountDao, categoryDao := accountModel.NewDao(tx), categoryModel.NewDao(tx)

	accountMappings, err := accountDao.SelectMultipleMapping(*accountModel.NewMappingCondition().WithRelatedId(indAccountTrans.AccountId))
	if err != nil {
		return err
	}

	categoryMapping, transMapping, syncTrans := categoryModel.Mapping{}, transactionModel.Mapping{}, indAccountTrans.SyncDataClone()

	for _, accountMapping := range accountMappings {
		categoryMapping, err = categoryDao.SelectMapping(accountMapping.MainId, indAccountTrans.CategoryId)
		if errors.As(err, gorm.ErrRecordNotFound) {
			continue
		} else if err != nil {
			return err
		}

		syncTrans.AccountId = categoryMapping.ParentAccountId
		syncTrans.CategoryId = categoryMapping.ParentCategoryId
		transMapping, err = transactionModel.NewDao(tx).SelectMappingByTrans(indAccountTrans, syncTrans)
		if err == nil {
			syncTrans.ID = transMapping.RelatedId
			var accountUser accountModel.User
			accountUser, err = accountDao.SelectUser(syncTrans.AccountId, syncTrans.UserId)
			if err != nil {
				return err
			}
			option := txnService.NewDefaultOption()
			err = txnService.Update(
				syncTrans, accountUser, *option.WithTransSyncToMappingAccount(false),
				context.WithValue(ctx, contextKey.Tx, tx),
			)
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			_, err = txnService.CreateSyncTrans(indAccountTrans, syncTrans, ctx)
		}
		if err != nil && false == errors.Is(err, gorm.ErrDuplicatedKey) {
			return err
		}
	}
	return nil
}

func (txnService *Transaction) SyncToIndependentAccount(
	shareAccountTrans transactionModel.Transaction, ctx context.Context,
) error {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
	if err := shareAccountTrans.ForUpdate(tx); err != nil {
		return err
	}
	var err error
	accountMapping, err := accountModel.NewDao(tx).SelectMappingByMainAccountAndRelatedUser(
		shareAccountTrans.AccountId, shareAccountTrans.UserId,
	)
	if err != nil {
		return err
	}
	categoryMapping, err := categoryModel.NewDao(tx).SelectMappingByCAccountIdAndPCategoryId(
		accountMapping.RelatedId, shareAccountTrans.CategoryId,
	)
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
		syncTrans.ID = transMapping.MainId
		var accountUser accountModel.User
		accountUser, err = accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
		if err != nil {
			return err
		}
		option := txnService.NewDefaultOption()
		err = txnService.Update(syncTrans, accountUser, *option.WithTransSyncToMappingAccount(false), ctx)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		_, err = txnService.CreateSyncTrans(shareAccountTrans, syncTrans, ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func (txnService *Transaction) CreateSyncTrans(
	trans, syncTrans transactionModel.Transaction, ctx context.Context,
) (mapping transactionModel.Mapping, err error) {
	tx := ctx.Value(contextKey.Tx).(*gorm.DB)
	accountUser, err := accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
	if err != nil {
		return
	}
	option := txnService.NewDefaultOption()
	newTrans, err := txnService.Create(syncTrans, accountUser, *option.WithTransSyncToMappingAccount(false), ctx)
	if err != nil {
		return
	}

	accountType, err := accountModel.NewDao(tx).GetAccountType(trans.AccountId)
	if err != nil {
		return
	}
	switch accountType {
	case accountModel.TypeIndependent:
		mapping, err = txnService.CreateMapping(trans, newTrans, tx)
	case accountModel.TypeShare:
		mapping, err = txnService.CreateMapping(newTrans, trans, tx)
	default:
		panic("err account.type")
	}
	if err != nil {
		return
	}
	return
}

func (txnService *Transaction) CreateMapping(
	mainTrans, relatedTrans transactionModel.Transaction, tx *gorm.DB,
) (transactionModel.Mapping, error) {
	mapping := transactionModel.Mapping{MainId: mainTrans.ID, RelatedId: relatedTrans.ID}
	err := tx.Create(&mapping).Error
	return mapping, err
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

// Option
type Option struct {
	syncUpdateStatistic       bool // syncUpdateStatistic 同步/异步更新统计数据
	transSyncToMappingAccount bool // transSyncToMappingAccount 交易数据至同步关联账本
}

func (txnService *Transaction) NewDefaultOption() Option {
	return Option{syncUpdateStatistic: false, transSyncToMappingAccount: true}
}

func (txnService *Transaction) NewOption() *Option {
	return &Option{}
}

func (co *Option) WithSyncUpdateStatistic(val bool) *Option {
	co.syncUpdateStatistic = val
	return co
}

func (co *Option) WithTransSyncToMappingAccount(val bool) *Option {
	co.transSyncToMappingAccount = val
	return co
}

func handelTaskList(taskList []func(tx *gorm.DB) error, ctx context.Context) error {
	if len(taskList) > 2 {
		errGroup, errGroupCtx := errgroup.WithContext(ctx)
		for i := range taskList {
			tx, task := errGroupCtx.Value(contextKey.Tx).(*gorm.DB), taskList[i]
			errGroup.Go(func() error { return tx.Transaction(task) })
		}
		if err := errGroup.Wait(); err != nil {
			return err
		}
	} else {
		tx := ctx.Value(contextKey.Tx).(*gorm.DB)
		for _, task := range taskList {
			err := tx.Transaction(task)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
