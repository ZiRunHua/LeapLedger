package transactionService

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/util"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

type Transaction struct{}

func (txnService *Transaction) CreateOne(transaction *transactionModel.Transaction) error {
	if err := txnService.checkTransaction(transaction); err != nil {
		return err
	}
	err := transaction.GetDb().Create(&transaction).Error
	if err != nil {
		return errors.Wrap(err, "")
	}
	return txnService.updateStatistic(
		transaction.IncomeExpense, transaction.TradeTime, transaction.CategoryID, transaction.AccountID,
		transaction.Amount,
	)
}

func (txnService *Transaction) updateStatistic(
	incomeExpense constant.IncomeExpense, tradeTime time.Time, categoryId uint, accountId uint, amount int,
) error {
	switch incomeExpense {
	case constant.Income:
		var incomeStatistic transactionModel.IncomeStatistic
		if err := incomeStatistic.Accumulate(
			tradeTime, categoryId, accountId, amount,
		); err != nil {
			return errors.Wrap(err, "incomeStatistic.Accumulate")
		}
	case constant.Expense:
		var expenseStatistic transactionModel.ExpenseStatistic
		if err := expenseStatistic.Accumulate(
			tradeTime, categoryId, accountId, amount,
		); err != nil {
			return errors.Wrap(err, "expenseStatistic.Accumulate")
		}
	default:
		panic("income Expense error")
	}
	return nil
}

func (txnService *Transaction) checkTransaction(transaction *transactionModel.Transaction) error {
	var category categoryModel.Category
	err := category.SelectById(transaction.CategoryID, false)
	if err != nil {
		return errors.Wrap(err, "")
	}
	if category.AccountID != transaction.AccountID || transaction.Amount < 0 {
		return errors.Wrap(global.ErrInvalidParameter, "")
	}
	return nil
}

func (txnService *Transaction) Update(transaction *transactionModel.Transaction) error {
	var err error
	if err = txnService.checkTransaction(transaction); err != nil {
		return err
	}
	oldTransaction := transactionModel.NewTransaction(transaction.GetTransaction())
	if err = oldTransaction.SelectById(transaction.ID, true); err != nil {
		return errors.Wrap(err, "transactionModel SelectById")
	}
	if err = transaction.Update(); err != nil {
		return errors.Wrap(err, "transactionModel Update")
	}
	return nil
}

func (txnService *Transaction) updateStatisticAfterUpdate(
	oldTxn *transactionModel.Transaction, txn transactionModel.Transaction,
) error {
	var err error
	if oldTxn.IncomeExpense == txn.IncomeExpense && oldTxn.CategoryID == txn.CategoryID && util.IsSameDay(
		oldTxn.TradeTime, txn.TradeTime,
	) { //同表同一条记录特殊处理
		err = txnService.updateStatistic(
			txn.IncomeExpense, txn.TradeTime, txn.CategoryID, txn.AccountID, txn.Amount-oldTxn.Amount,
		)
		if err != nil {
			return err
		}
	} else {
		if err = txnService.updateStatistic(
			oldTxn.IncomeExpense, oldTxn.TradeTime, oldTxn.CategoryID, oldTxn.AccountID, -oldTxn.Amount,
		); err != nil {
			return err
		}

		if err = txnService.updateStatistic(
			txn.IncomeExpense, txn.TradeTime, txn.CategoryID, txn.AccountID, txn.Amount,
		); err != nil {
			return err
		}
	}
	return nil
}

func (txnService *Transaction) CreateMultiple(
	account *accountModel.Account, transactionList []transactionModel.Transaction, tx *gorm.DB,
) ([]*transactionModel.Transaction, error) {
	var categoryIds []uint
	var err error
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
	failTransList, incomeTransList, expenseTransList := []*transactionModel.Transaction{}, []*transactionModel.Transaction{},
		[]*transactionModel.Transaction{}

	var key string
	for index, _ := range transactionList {
		transaction := transactionList[index]
		if !categoryIdMap[transaction.CategoryID] {
			failTransList = append(failTransList, &transaction)
			continue
		}
		if transaction.IncomeExpense == constant.Income {
			incomeTransList = append(incomeTransList, &transaction)
			key = transaction.TradeTime.Format("2006-01-02")
			if incomeAmount[key] == nil {
				incomeAmount[key] = map[uint]int{transaction.CategoryID: transaction.Amount}
			} else {
				incomeAmount[key][transaction.CategoryID] += transaction.Amount
			}
		} else if transaction.IncomeExpense == constant.Expense {
			expenseTransList = append(expenseTransList, &transaction)
			key = transaction.TradeTime.Format("2006-01-02")
			if expenseAmount[key] == nil {
				expenseAmount[key] = map[uint]int{transaction.CategoryID: transaction.Amount}
			} else {
				expenseAmount[key][transaction.CategoryID] += transaction.Amount
			}
		} else {
			failTransList = append(failTransList, &transaction)
			continue
		}
	}
	fmt.Println(incomeTransList)
	for _, t := range expenseTransList {
		fmt.Println(t)
	}
	var transaction transactionModel.Transaction
	if len(incomeTransList) > 0 {
		if err = tx.Model(&transaction).Create(incomeTransList).Error; err != nil {
			return nil, err
		}

		if err = txnService.addStatisticAfterCreateMultiple(account, constant.Income, incomeAmount); err != nil {
			return nil, err
		}
	}
	if len(expenseTransList) > 0 {
		if err = tx.Model(&transaction).Create(expenseTransList).Error; err != nil {
			return nil, err
		}
		if err = txnService.addStatisticAfterCreateMultiple(account, constant.Expense, expenseAmount); err != nil {
			return nil, err
		}
	}
	return failTransList, err
}

func (txnService *Transaction) addStatisticAfterCreateMultiple(
	account *accountModel.Account, incomeExpense constant.IncomeExpense, amountList map[string]map[uint]int,
) error {
	var err error
	var tradeTime time.Time
	for date, categoryList := range amountList {
		if tradeTime, err = time.Parse("2006-01-02", date); err != nil {
			return err
		}
		for categoryId, amount := range categoryList {
			if err = txnService.updateStatistic(incomeExpense, tradeTime, categoryId, account.ID, amount); err != nil {
				return err
			}
		}
	}
	return nil
}
