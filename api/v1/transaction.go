package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	transactionModel "KeepAccount/model/transaction"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type _transactionApi interface {
	transactionApi()
	GetOne(ctx *gin.Context)
	CreateOne(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	GetList(ctx *gin.Context)
}

type TransactionApi struct {
}

func (a *TransactionApi) transactionApi() {}
func (a *TransactionApi) GetOne(ctx *gin.Context) {
	trans, ok := contextFunc.GetTransByParam(ctx)
	if false == ok {
		return
	}
	response.OkWithData(response.TransactionModelToResponse(trans), ctx)
}

func (a *TransactionApi) CreateOne(ctx *gin.Context) {
	var requestData request.TransactionCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); false == pass {
		return
	}
	user, err := contextFunc.GetUser(ctx)
	if err != nil {
		response.FailToError(ctx, err)
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
	err = global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			transaction.SetTx(tx)
			return transactionService.CreateOne(transaction, user)
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	responseData := response.Id{
		Id: transaction.ID,
	}
	response.OkWithData(responseData, ctx)
}

func (a *TransactionApi) Update(ctx *gin.Context) {
	var requestData request.TransactionUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	id, ok := contextFunc.GetParamId(ctx)
	if false == ok {
		return
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
	transaction.ID = id
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			transaction.SetTx(tx)
			return transactionService.Update(transaction)
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (a *TransactionApi) Delete(ctx *gin.Context) {
	trans, ok := contextFunc.GetTransByParam(ctx)
	if false == ok {
		return
	}
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			trans.SetTx(tx)
			return transactionService.Delete(trans)
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

func (a *TransactionApi) GetList(ctx *gin.Context) {
	var requestData request.TransactionGetList
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	user, err := contextFunc.GetUser(ctx)
	if err != err {
		response.FailToError(ctx, err)
		return
	}
	if requestData.UserId != nil && user.ID != *requestData.UserId {
		response.FailToError(ctx, errors.New("UserId数据异常"))
		return
	}
	if requestData.AccountId != nil {
		if pass, _ := checkFunc.AccountBelong(*requestData.AccountId, ctx); pass == false {
			return
		}
	}

	var startTime, endTime *time.Time
	startTime = request.GetTimeByTimestamp(requestData.StartTime)
	endTime = request.GetTimeByTimestamp(requestData.EndTime)
	transactionList, err := transactionModel.NewTransactionDao(nil).GetListByCondition(
		&transactionModel.TransactionCondition{
			UserID: requestData.UserId, AccountID: requestData.AccountId,
			CategoryID: requestData.CategoryId, IncomeExpense: requestData.IncomeExpense, TradeStartTime: startTime,
			TradeEndTime: endTime,
		},
		requestData.Limit,
		requestData.Offset,
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	responseData := response.TransactionGetList{List: []response.TransactionOne{}}
	for _, transaction := range *transactionList {
		responseData.List = append(
			responseData.List, *response.TransactionModelToResponse(&transaction),
		)
	}
	response.OkWithData(responseData, ctx)
}
