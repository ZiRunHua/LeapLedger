package transactionService

import (
	"context"
	"time"

	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/cus"
	"KeepAccount/global/db"
	"KeepAccount/global/nats"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Transaction struct{}

func (txnService *Transaction) Create(
	transInfo transactionModel.Info,
	accountUser accountModel.User,
	recordType transactionModel.RecordType,
	option Option,
	ctx context.Context,
) (trans transactionModel.Transaction, err error) {
	return trans, db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			tx := ctx.GetDb()
			// check
			transInfo.AccountId = accountUser.AccountId
			err = transInfo.Check(tx)
			if err != nil {
				return errors.WithStack(err)
			}
			err = accountUser.CheckTransAddByUserId(transInfo.UserId)
			if err != nil {
				return errors.WithStack(err)
			}
			// handle
			trans, err = transactionModel.NewDao(tx).Create(transInfo, recordType)
			if err != nil {
				return errors.WithStack(err)
			}
			if option.isSyncTrans {
				err = nats.PublishTaskToOutboxWithPayload(ctx, nats.TaskTransactionSync, trans)
				if err != nil {
					return err
				}
			}
			return nats.PublishEventToOutboxWithPayload(ctx, nats.EventTransactionCreate, trans)
		},
	)
}

// Update only "user_id,income_expense,category_id,amount,remark,trade_time" can be changed
func (txnService *Transaction) Update(
	trans transactionModel.Transaction, accountUser accountModel.User, option Option, ctx context.Context,
) error {
	return db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			tx := ctx.GetDb()
			// check
			err := trans.Check(tx)
			if err != nil {
				return errors.WithStack(err)
			}
			err = accountUser.CheckTransEditByUserId(trans.UserId)
			if err != nil {
				return errors.WithStack(err)
			}
			// handle
			oldTrans := trans
			if err = oldTrans.ForShare(tx); err != nil {
				return errors.WithStack(err)
			}
			if oldTrans.UpdatedAt.Add(time.Second * 3).After(time.Now()) {
				return errors.WithStack(global.ErrFrequentOperation)
			}
			err = tx.Model(&trans).Select(
				"income_expense", "category_id", "amount", "remark", "trade_time",
			).Updates(trans).Error
			if err != nil {
				return errors.WithStack(err)
			}
			if option.isSyncTrans {
				err = nats.PublishTaskToOutboxWithPayload(ctx, nats.TaskTransactionSync, trans)
				if err != nil {
					return err
				}
			}
			return nats.PublishEventToOutboxWithPayload(
				ctx, nats.EventTransactionUpdate,
				nats.EventTransactionUpdatePayload{OldTrans: oldTrans, NewTrans: trans},
			)
		},
	)
}

func (txnService *Transaction) Delete(
	txn transactionModel.Transaction, accountUser accountModel.User, ctx context.Context,
) error {
	err := db.Transaction(
		ctx, func(ctx *cus.TxContext) error {
			tx := ctx.GetDb()
			err := accountUser.CheckTransEditByUserId(txn.UserId)
			if err != nil {
				return err
			}
			err = tx.Delete(&txn).Error
			if err != nil {
				return err
			}
			return nats.PublishEventToOutboxWithPayload(ctx, nats.EventTransactionDelete, txn)
		},
	)
	return err
}

func (txnService *Transaction) updateStatistic(data transactionModel.StatisticData, ctx context.Context) (err error) {
	l, err := time.LoadLocation(data.Location)
	if err != nil {
		return err
	}
	data.TradeTime = data.TradeTime.In(l)

	switch data.IncomeExpense {
	case constant.Income:
		err = db.Transaction(
			ctx, func(ctx *cus.TxContext) error {
				return transactionModel.IncomeAccumulate(
					data.TradeTime, data.AccountId, data.UserId, data.CategoryId, data.Amount, data.Count, ctx.GetDb(),
				)
			},
		)
	case constant.Expense:
		err = db.Transaction(
			ctx, func(ctx *cus.TxContext) error {
				return transactionModel.ExpenseAccumulate(
					data.TradeTime, data.AccountId, data.UserId, data.CategoryId, data.Amount, data.Count, ctx.GetDb(),
				)
			},
		)
	default:
		panic("income expense error")
	}
	return err
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

func (txnService *Transaction) syncToShareAccount(
	indAccountTrans transactionModel.Transaction, ctx context.Context,
) error {
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
			syncTrans.ID = transMapping.RelatedId
			err = txnService.syncUpdate(transMapping, indAccountTrans, syncTrans, ctx)
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

func (txnService *Transaction) syncToIndependentAccount(
	shareAccountTrans transactionModel.Transaction, ctx context.Context,
) error {
	tx := db.Get(ctx)
	var err error
	accountMapping, err := accountModel.NewDao(tx).SelectMappingByMainAccountAndRelatedUser(
		shareAccountTrans.AccountId, shareAccountTrans.UserId,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
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
		err = txnService.syncUpdate(transMapping, shareAccountTrans, syncTrans, ctx)
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

func (txnService *Transaction) syncUpdate(
	mapping transactionModel.Mapping, origin, target transactionModel.Transaction, ctx context.Context,
) error {
	tx := db.Get(ctx)
	err := mapping.ForShare(tx)
	if err != nil {
		return err
	}
	if false == mapping.CanSyncTrans(origin) {
		return nil
	}

	var accountUser accountModel.User
	accountUser, err = accountModel.NewDao(tx).SelectUser(target.AccountId, target.UserId)
	if err != nil {
		return err
	}
	option := txnService.NewDefaultOption()
	err = txnService.Update(target, accountUser, option, ctx)
	if err != nil {
		return err
	}
	err = mapping.OnSyncSuccess(tx, origin)
	if err != nil {
		return err
	}
	return nil
}

func (txnService *Transaction) CreateSyncTrans(
	trans, syncTrans transactionModel.Transaction, ctx context.Context,
) (mapping transactionModel.Mapping, err error) {
	tx := db.Get(ctx)
	accountUser, err := accountModel.NewDao(tx).SelectUser(syncTrans.AccountId, syncTrans.UserId)
	if err != nil {
		return
	}
	newTrans, err := txnService.Create(
		syncTrans.Info, accountUser, transactionModel.RecordTypeOfSync, txnService.NewDefaultOption(), ctx,
	)
	if err != nil {
		return
	}

	mapping, err = txnService.CreateMapping(trans, newTrans, ctx)
	if err != nil {
		return
	}
	return
}

func (txnService *Transaction) CreateMapping(
	trans1, trans2 transactionModel.Transaction, ctx context.Context,
) (mapping transactionModel.Mapping, err error) {
	tx := db.Get(ctx)
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
	ctx := context.WithValue(context.Background(), cus.Tx, tx)
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
				}, ctx,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

type Option struct {
	// isSyncTrans 交易数据至同步关联账本
	isSyncTrans bool
}

func (txnService *Transaction) NewDefaultOption() Option {
	return Option{isSyncTrans: false}
}

func (txnService *Transaction) NewOption() Option {
	return Option{}
}

func (txnService *Transaction) NewOptionFormConfig(trans transactionModel.Info, ctx context.Context) (
	option Option, err error,
) {
	userConfig, err := accountModel.NewDao(db.Get(ctx)).SelectUserConfig(trans.AccountId, trans.UserId)
	if err != nil {
		return
	}
	option = txnService.NewDefaultOption()
	option.isSyncTrans = userConfig.GetFlagStatus(accountModel.Flag_Trans_Sync_Mapping_Account)
	return
}

func (o *Option) IsSyncTrans() *Option {
	o.isSyncTrans = true
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
