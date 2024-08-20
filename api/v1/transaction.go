package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global/constant"
	"KeepAccount/global/cusCtx"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/util/dataTool"
	"KeepAccount/util/timeTool"
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type TransactionApi struct {
}

// GetOne
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Param		id			path		int	true	"Transaction ID"
//	@Success	200			{object}	response.Data{Data=response.TransactionDetail}
//	@Router		/account/{accountId}/transaction/{id} [get]
func (t *TransactionApi) GetOne(ctx *gin.Context) {
	trans, ok := contextFunc.GetTransByParam(ctx)
	if false == ok {
		return
	}
	var data response.TransactionDetail
	err := data.SetData(trans, nil)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(data, ctx)
}

// CreateOne
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		id			path		int								true	"Transaction ID"
//	@Param		body		body		request.TransactionCreateOne	true	"transaction data"
//	@Success	200			{object}	response.Data{Data=response.TransactionDetail}
//	@Router		/account/{accountId}/transaction/{id} [post]
func (t *TransactionApi) CreateOne(ctx *gin.Context) {
	var requestData request.TransactionCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}

	accountUser := contextFunc.GetAccountUser(ctx)
	transaction := transactionModel.Transaction{
		Info: transactionModel.Info{
			AccountId:     accountUser.AccountId,
			UserId:        accountUser.UserId,
			CategoryId:    requestData.CategoryId,
			IncomeExpense: requestData.IncomeExpense,
			Amount:        requestData.Amount,
			Remark:        requestData.Remark,
			TradeTime:     requestData.TradeTime,
		},
	}
	err := db.Transaction(
		ctx, func(ctx *cusCtx.TxContext) error {
			createOption, err := transactionService.NewOptionFormConfig(transaction.Info, ctx)
			if err != nil {
				return err
			}
			createOption.WithSyncUpdateStatistic(false)
			transaction, err = transactionService.Create(transaction.Info, accountUser, createOption, ctx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}

	var responseData response.TransactionDetail
	if err = responseData.SetData(transaction, nil); responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// Update
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		id			path		int								true	"Transaction ID"
//	@Param		body		body		request.TransactionCreateOne	true	"Transaction data"
//	@Success	200			{object}	response.Data{Data=response.TransactionDetail}
//	@Router		/account/{accountId}/transaction/{id} [put]
func (t *TransactionApi) Update(ctx *gin.Context) {
	var requestData request.TransactionUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	oldTrans, ok := contextFunc.GetTransByParam(ctx)
	if false == ok {
		return
	}
	transaction := transactionModel.Transaction{
		Info: transactionModel.Info{
			UserId:        oldTrans.UserId,
			AccountId:     requestData.AccountId,
			CategoryId:    requestData.CategoryId,
			IncomeExpense: requestData.IncomeExpense,
			Amount:        requestData.Amount,
			Remark:        requestData.Remark,
			TradeTime:     requestData.TradeTime,
		},
	}
	transaction.ID = oldTrans.ID

	txCtx := context.WithValue(ctx, cusCtx.Db, db.Db)
	option, err := transactionService.NewOptionFormConfig(transaction.Info, txCtx)
	if responseError(err, ctx) {
		return
	}
	option.WithSyncUpdateStatistic(true)
	err = transactionService.Update(transaction, contextFunc.GetAccountUser(ctx), option, txCtx)
	if responseError(err, ctx) {
		return
	}

	var responseData response.TransactionDetail
	if err = responseData.SetData(transaction, nil); responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// Delete
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Param		id			path		int	true	"Transaction ID"
//	@Success	200			{object}	response.NoContent
//	@Router		/account/{accountId}/transaction/{id} [delete]
func (t *TransactionApi) Delete(ctx *gin.Context) {
	trans, pass := contextFunc.GetTransByParam(ctx)
	if false == pass {
		return
	}
	accountUser, err := accountModel.NewDao().SelectUser(trans.AccountId, contextFunc.GetUserId(ctx))
	if responseError(err, ctx) {
		return
	}
	err = db.Db.Transaction(
		func(tx *gorm.DB) error {
			return transactionService.Delete(trans, accountUser, tx)
		},
	)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	response.Ok(ctx)
}

// GetList
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int	true	"Account ID"
//	@Success	200			{object}	response.Data{Data=response.List[response.TransactionDetail]{}}
//	@Router		/account/{accountId}/transaction/list [get]
func (t *TransactionApi) GetList(ctx *gin.Context) {
	var requestData request.TransactionGetList
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	_, err := contextFunc.GetUser(ctx)
	if responseError(err, ctx) {
		return
	}

	// 查询并获取结果
	condition := requestData.GetCondition()
	condition.AccountId = contextFunc.GetAccountId(ctx)
	var transactionList []transactionModel.Transaction
	transactionList, err = transactionModel.NewDao().GetListByCondition(
		condition, requestData.Offset, requestData.Limit,
	)
	if responseError(err, ctx) {
		return
	}
	responseData := response.TransactionGetList{List: response.TransactionDetailList{}}
	err = responseData.List.SetData(transactionList)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetTotal
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		data		body		request.TransactionTotal	true	"Transaction total data"
//	@Success	200			{object}	response.Data{Data=response.TransactionTotal{}}
//	@Router		/account/{accountId}/transaction/total [get]
func (t *TransactionApi) GetTotal(ctx *gin.Context) {
	var requestData request.TransactionTotal
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	// 查询条件
	condition := requestData.GetStatisticCondition()
	condition.AccountId = contextFunc.GetAccountId(ctx)
	extCond := requestData.GetExtensionCondition()
	// 查询并处理响应
	total, err := transactionModel.NewDao().GetIeStatisticByCondition(
		requestData.IncomeExpense, condition, &extCond,
	)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.TransactionTotal{IEStatistic: total}, ctx)
}

// GetMonthStatistic
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int									true	"Account ID"
//	@Param		data		body		request.TransactionMonthStatistic	true	"condition"
//	@Success	200			{object}	response.Data{Data=response.List[response.TransactionStatistic]{}}
//	@Router		/account/{accountId}/transaction/month/statistic [get]
func (t *TransactionApi) GetMonthStatistic(ctx *gin.Context) {
	var requestData request.TransactionMonthStatistic
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if pass := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}
	requestData.AccountId = contextFunc.GetAccountId(ctx)
	// 设置查询条件
	statisticCondition, extCond := requestData.GetStatisticCondition(), requestData.GetExtensionCondition()
	condition := statisticCondition
	months := timeTool.SplitMonths(statisticCondition.StartTime, statisticCondition.EndTime)
	// 查询并处理响应
	responseList := make([]response.TransactionStatistic, len(months), len(months))
	dao := transactionModel.NewDao()
	for i := len(months) - 1; i >= 0; i-- {
		condition.StartTime = months[i][0]
		condition.EndTime = months[i][1]

		monthStatistic, err := dao.GetIeStatisticByCondition(requestData.IncomeExpense, condition, &extCond)
		if responseError(err, ctx) {
			return
		}
		responseList[i] = response.TransactionStatistic{
			IEStatistic: monthStatistic,
			StartTime:   condition.StartTime,
			EndTime:     condition.EndTime,
		}
	}
	response.OkWithData(response.List[response.TransactionStatistic]{List: responseList}, ctx)
}

// GetDayStatistic
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		data		body		request.TransactionDayStatistic	true	"condition"
//	@Success	200			{object}	response.Data{Data=response.List[response.TransactionDayStatistic]{}}
//	@Router		/account/{accountId}/transaction/day/statistic [get]
func (t *TransactionApi) GetDayStatistic(ctx *gin.Context) {
	var requestData request.TransactionDayStatistic
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	requestData.AccountId = contextFunc.GetAccountId(ctx)
	// 处理请求
	var startTime, endTime = requestData.FormatDayTime()
	days := timeTool.SplitDays(startTime, endTime)
	dayMap := make(map[time.Time]*response.TransactionDayStatistic, len(days))
	condition := transactionModel.StatisticCondition{
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId:   requestData.AccountId,
			CategoryIds: requestData.CategoryIds,
		},
		StartTime: startTime,
		EndTime:   endTime,
	}
	handleFunc := func(ie constant.IncomeExpense) error {
		statistics, err := transactionModel.NewStatisticDao().GetDayStatisticByCondition(ie, condition)
		if err != nil {
			return err
		}
		for _, item := range statistics {
			dayMap[item.Date].Amount += item.Amount
			dayMap[item.Date].Count += item.Count
		}
		return nil
	}
	// 处理响应
	var err error
	responseData := make([]response.TransactionDayStatistic, len(days), len(days))
	for i, day := range days {
		responseData[i] = response.TransactionDayStatistic{Date: day}
		dayMap[day] = &responseData[i]
	}
	if requestData.IncomeExpense != nil {
		err = handleFunc(*requestData.IncomeExpense)
		if responseError(err, ctx) {
			return
		}
	} else {
		if err = handleFunc(constant.Income); responseError(err, ctx) {
			return
		}
		if err = handleFunc(constant.Expense); responseError(err, ctx) {
			return
		}
	}
	response.OkWithData(
		response.List[response.TransactionDayStatistic]{List: responseData}, ctx,
	)
}

// GetCategoryAmountRank
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int										true	"Account ID"
//	@Param		data		body		request.TransactionCategoryAmountRank	true	"condition"
//	@Success	200			{object}	response.Data{Data=response.List[response.TransactionCategoryAmountRank]{}}
//	@Router		/account/{accountId}/transaction/category/amount/rank [get]
func (t *TransactionApi) GetCategoryAmountRank(ctx *gin.Context) {
	var requestData request.TransactionCategoryAmountRank
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); responseError(err, ctx) {
		return
	}
	requestData.AccountId = contextFunc.GetAccountId(ctx)
	account := contextFunc.GetAccount(ctx)
	// fetch ranking List
	var startTime, endTime = requestData.FormatDayTime()
	condition := transactionModel.CategoryAmountRankCondition{
		Account:   account,
		StartTime: startTime,
		EndTime:   endTime,
	}
	var err error
	var rankingList dataTool.Slice[uint, transactionModel.CategoryAmountRank]
	rankingList, err = transactionModel.NewStatisticDao().GetCategoryAmountRank(
		requestData.IncomeExpense, condition, requestData.Limit,
	)

	if responseError(err, ctx) {
		return
	}
	categoryIds := rankingList.ExtractValues(
		func(rank transactionModel.CategoryAmountRank) uint {
			return rank.CategoryId
		},
	)
	// fetch category
	var categoryList dataTool.Slice[uint, categoryModel.Category]
	err = db.Db.Where("id IN (?)", categoryIds).Find(&categoryList).Error
	if responseError(err, ctx) {
		return
	}
	categoryMap := categoryList.ToMap(
		func(category categoryModel.Category) uint {
			return category.ID
		},
	)
	// response
	responseData := make([]response.TransactionCategoryAmountRank, len(rankingList), len(rankingList))
	for i, rank := range rankingList {
		responseData[i].Amount = rank.Amount
		responseData[i].Count = rank.Count
		err = responseData[i].Category.SetData(categoryMap[rank.CategoryId])
		if responseError(err, ctx) {
			return
		}
	}
	// 数量不足时补足响应数量
	if requestData.Limit != nil && len(rankingList) < *requestData.Limit {
		categoryList = []categoryModel.Category{}
		limit := *requestData.Limit - len(rankingList)
		query := db.Db.Where("account_id = ?", account.ID)
		query = query.Where("income_expense = ?", requestData.IncomeExpense)
		err = query.Where("id NOT IN (?)", categoryIds).Limit(limit).Find(&categoryList).Error
		if responseError(err, ctx) {
			return
		}
		for _, category := range categoryList {
			responseCategory := response.TransactionCategoryAmountRank{}
			err = responseCategory.Category.SetData(category)
			if responseError(err, ctx) {
				return
			}
			responseData = append(responseData, responseCategory)
		}
	}
	response.OkWithData(response.List[response.TransactionCategoryAmountRank]{List: responseData}, ctx)
}

// GetAmountRank
//
//	@Tags		Transaction
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int								true	"Account ID"
//	@Param		data		body		request.TransactionAmountRank	true	"condition"
//	@Success	200			{object}	response.Data{Data=response.List[response.TransactionDetailList]{}}
//	@Router		/account/{accountId}/transaction/amount/rank [get]
func (t *TransactionApi) GetAmountRank(ctx *gin.Context) {
	var requestData request.TransactionAmountRank
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	requestData.AccountId = contextFunc.GetAccountId(ctx)
	// fetch
	timeCond := transactionModel.NewTimeCondition()
	timeCond.SetTradeTimes(requestData.StartTime, requestData.EndTime)
	rankingList, err := transactionModel.NewDao().GetAmountRank(
		requestData.AccountId, requestData.IncomeExpense, *timeCond,
	)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseList response.TransactionDetailList
	err = responseList.SetData(rankingList)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.TransactionDetail]{List: responseList}, ctx)
}

// CreateTiming
//
//	@Tags		Transaction/Timing
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		data		body		request.TransactionTiming	true	"timing config"
//	@Success	200			{object}	response.Data{Data=response.TransactionTiming}
//	@Router		/account/{accountId}/transaction/timing [post]
func (t *TransactionApi) CreateTiming(ctx *gin.Context) {
	var requestData request.TransactionTiming
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	requestData.Trans.AccountId = contextFunc.GetAccountId(ctx)
	timing := requestData.GetTimingModel()
	if timing.UserId != contextFunc.GetUserId(ctx) || timing.TransInfo.UserId != contextFunc.GetUserId(ctx) {
		response.Forbidden(ctx)
		return
	}
	// handle
	var err error
	err = db.Transaction(
		ctx, func(ctx *cusCtx.TxContext) error {
			timing, err = transactionService.Timing.CreateTiming(timing, ctx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.TransactionTiming
	err = responseData.SetData(timing)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// UpdateTiming
//
//	@Tags		Transaction/Timing
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int							true	"Account ID"
//	@Param		id			path		int							true	"Transaction ID"
//	@Param		data		body		request.TransactionTiming	true	"timing config"
//	@Success	200			{object}	response.Data{Data=response.TransactionTiming}
//	@Router		/account/{accountId}/transaction/timing/{id} [put]
func (t *TransactionApi) UpdateTiming(ctx *gin.Context) {
	var requestData request.TransactionTiming
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	requestData.Trans.AccountId = contextFunc.GetAccountId(ctx)
	timing := requestData.GetTimingModel()
	timing.ID = contextFunc.GetId(ctx)
	// handle
	var err error
	err = db.Transaction(
		ctx, func(ctx *cusCtx.TxContext) error {
			tx := ctx.GetDb()
			timing, err = transactionService.Timing.UpdateTiming(timing, context.WithValue(ctx, cusCtx.Db, tx))
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.TransactionTiming
	err = responseData.SetData(timing)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

// GetTimingList
//
//	@Tags		Transaction/Timing
//	@Accept		json
//	@Produce	json
//	@Param		accountId	path		int					true	"Account ID"
//	@Param		data		body		request.PageData	true	"Page data"
//	@Success	200			{object}	response.Data{Data=response.List[response.TransactionTiming]}
//	@Router		/account/{accountId}/transaction/timing/list [get]
func (t *TransactionApi) GetTimingList(ctx *gin.Context) {
	var requestData request.PageData
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	list, err := transactionModel.NewDao().SelectTimingListByUserId(contextFunc.GetAccountId(ctx), requestData.Offset, requestData.Limit)
	if responseError(err, ctx) {
		return
	}
	// response
	var responseData response.TransactionTimingList
	err = responseData.SetData(list)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.List[response.TransactionTiming]{List: responseData}, ctx)
}
