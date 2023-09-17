package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	transactionModel "KeepAccount/model/transaction"
	"database/sql"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type TransactionApi struct {
}

func (a *TransactionApi) CreateOne(ctx *gin.Context) {
	var requestData request.TransactionCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
	}
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); false == pass {
		return
	}

	transaction := &transactionModel.Transaction{
		AccountID:     requestData.AccountId,
		CategoryID:    requestData.CategoryId,
		IncomeExpense: requestData.IncomeExpense,
		Amount:        requestData.Amount,
		Remark:        requestData.Remark,
		TradeTime:     time.Unix(int64(requestData.TradeTime), 0),
	}
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			transaction.SetTx(tx)
			err := transactionService.CreateOne(transaction)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
	}
	response.Ok(ctx)
}

func (a *TransactionApi) Update(ctx *gin.Context) {
	var requestData request.TransactionUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
	}
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); false == pass {
		return
	}

	transaction := &transactionModel.Transaction{
		AccountID:     requestData.AccountId,
		CategoryID:    requestData.CategoryId,
		IncomeExpense: requestData.IncomeExpense,
		Amount:        requestData.Amount,
		Remark:        requestData.Remark,
		TradeTime:     time.Unix(int64(requestData.TradeTime), 0),
	}
	err := transactionService.CreateOne(transaction)
	if err != nil {
		response.FailToError(ctx, err)
	}
	response.Ok(ctx)
}

func (a *TransactionApi) Delete(ctx *gin.Context) {
	var requestData request.TransactionUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
	}
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); false == pass {
		return
	}

	transaction := &transactionModel.Transaction{
		AccountID:     requestData.AccountId,
		CategoryID:    requestData.CategoryId,
		IncomeExpense: requestData.IncomeExpense,
		Amount:        requestData.Amount,
		Remark:        requestData.Remark,
		TradeTime:     time.Unix(int64(requestData.TradeTime), 0),
	}
	err := transactionService.CreateOne(transaction)
	if err != nil {
		response.FailToError(ctx, err)
	}
	response.Ok(ctx)
}

func (a *TransactionApi) GetList(ctx *gin.Context) {
	var requestData request.TransactionGetList
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
	}

	var transaction transactionModel.Transaction
	where := map[string]interface{}{
		"category_id": requestData.CategoryId, "income_expense": requestData.IncomeExpense,
	}
	var rows *sql.Rows
	var err error
	if requestData.StartTime != 0 && requestData.EndTime != 0 {
		rows, err = global.GvaDb.Model(&transaction).Where(where).Where("trans_time Between ? And ?").Rows()
		if handelError(err, ctx) {
			return
		}
	} else {
		rows, err = global.GvaDb.Model(&transaction).Where(where).Rows()
		if handelError(err, ctx) {
			return
		}
	}

	var responseData response.TransactionGetList
	for rows.Next() {
		err = global.GvaDb.ScanRows(rows, transaction)
		if handelError(err, ctx) {
			return
		}
		responseData.List = append(
			responseData.List, response.TransactionOne{
				AccountId:     transaction.AccountID,
				Amount:        transaction.Amount,
				CategoryId:    transaction.CategoryID,
				IncomeExpense: transaction.IncomeExpense,
				Remark:        transaction.Remark,
				TradeTime:     transaction.TradeTime.Unix(),
			},
		)
	}
	response.Ok(ctx)
}
