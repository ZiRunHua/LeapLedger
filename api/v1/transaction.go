package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	"KeepAccount/util"
	"KeepAccount/util/dataType"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type TransactionApi struct {
}

func (t *TransactionApi) transactionApi() {}
func (t *TransactionApi) GetOne(ctx *gin.Context) {
	trans, ok := contextFunc.GetTransByParam(ctx)
	if false == ok {
		return
	}
	response.OkWithData(response.TransactionModelToResponse(trans), ctx)
}

func (t *TransactionApi) CreateOne(ctx *gin.Context) {
	var requestData request.TransactionCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	account, accountUser, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if false == pass {
		return
	}

	transaction := transactionModel.Transaction{
		AccountId:     requestData.AccountId,
		CategoryId:    requestData.CategoryId,
		IncomeExpense: requestData.IncomeExpense,
		Amount:        requestData.Amount,
		Remark:        requestData.Remark,
		TradeTime:     time.Unix(int64(requestData.TradeTime), 0),
	}
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			userClient, err := contextFunc.GetUserCurrentClientInfo(ctx)
			if err != nil {
				return err
			}
			// 新交易非客户端当前使用账本 则异步更新统计数据 以加快接口响应
			asyncUpdateStatistic := transaction.AccountId != userClient.CurrentAccountId
			transaction, err = transactionService.CreateOne(transaction, accountUser, asyncUpdateStatistic, tx)
			return err
		},
	)
	if responseError(err, ctx) {
		return
	}

	var responseData response.TransactionDetail
	if err = responseData.SetData(transaction, &account); responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (t *TransactionApi) Update(ctx *gin.Context) {
	var requestData request.TransactionUpdateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	id, ok := contextFunc.GetParamId(ctx)
	if false == ok {
		return
	}
	account, accountUser, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if false == pass {
		return
	}
	transaction := transactionModel.Transaction{
		UserId:        requestData.UserId,
		AccountId:     requestData.AccountId,
		CategoryId:    requestData.CategoryId,
		IncomeExpense: requestData.IncomeExpense,
		Amount:        requestData.Amount,
		Remark:        requestData.Remark,
		TradeTime:     time.Unix(int64(requestData.TradeTime), 0),
	}
	transaction.ID = id
	err := global.GvaDb.Transaction(
		func(tx *gorm.DB) error {
			return transactionService.Update(transaction, accountUser, tx)
		},
	)
	if responseError(err, ctx) {
		return
	}

	var responseData response.TransactionDetail
	if err = responseData.SetData(transaction, &account); responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}

func (t *TransactionApi) Delete(ctx *gin.Context) {
	trans, pass := contextFunc.GetTransByParam(ctx)
	if false == pass {
		return
	}
	accountUser, err := accountModel.NewDao().SelectUser(trans.AccountId, contextFunc.GetUserId(ctx))
	if responseError(err, ctx) {
		return
	}
	err = global.GvaDb.Transaction(
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

	if pass := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}

	// 查询并获取结果
	condition := requestData.GetCondition()
	var transactionList []transactionModel.Transaction
	transactionList, err = transactionModel.NewDao().GetListByCondition(
		condition, requestData.Limit, requestData.Offset,
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
	if pass := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}
	// 查询条件
	condition := requestData.GetStatisticCondition()
	extCond := requestData.GetExtensionCondition()
	// 查询并处理响应
	total, err := transactionModel.NewDao().GetIeStatisticByCondition(
		requestData.IncomeExpense, condition, &extCond,
	)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.TransactionTotal{IncomeExpenseStatistic: total}, ctx)
}

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
	// 设置查询条件
	condition := requestData.GetStatisticCondition()
	months := util.Time.SplitMonths(condition.StartTime, condition.EndTime)
	// 查询并处理响应
	responseList := make([]response.TransactionStatistic, len(months), len(months))
	dao := transactionModel.NewDao()
	for i := 0; i < len(months); i++ {
		condition.StartTime = months[len(months)-i-1]
		condition.EndTime = util.Time.GetLastSecondOfMonth(condition.StartTime)

		monthStatistic, err := dao.GetIeStatisticByCondition(requestData.IncomeExpense, condition, nil)
		if responseError(err, ctx) {
			return
		}
		responseList[i] = response.TransactionStatistic{
			IncomeExpenseStatistic: monthStatistic,
			StartTime:              condition.StartTime.Unix(),
			EndTime:                condition.EndTime.Unix(),
		}
	}
	response.OkWithData(response.TransactionMonthStatistic{List: responseList}, ctx)
}

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
	account, _, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if pass == false {
		return
	}
	// 处理请求
	var startTime, endTime = requestData.FormatDayTime()
	days := util.Time.SplitDays(startTime, endTime)
	dayMap := make(map[time.Time]*response.TransactionDayStatistic, len(days))
	condition := transactionModel.StatisticCondition{
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId:   account.ID,
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
		responseData[i] = response.TransactionDayStatistic{Date: day.Unix()}
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

func (t *TransactionApi) GetCategoryAmountRank(ctx *gin.Context) {
	var requestData request.TransactionCategoryAmountRank
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	if err := requestData.CheckTimeFrame(); responseError(err, ctx) {
		return
	}
	account, _, pass := checkFunc.AccountBelongAndGet(requestData.AccountId, ctx)
	if pass == false {
		return
	}
	// 处理查询
	var startTime, endTime = requestData.FormatDayTime()
	condition := transactionModel.CategoryAmountRankCondition{
		Account:   account,
		StartTime: startTime,
		EndTime:   endTime,
	}
	var err error
	var rankingList dataType.Slice[uint, transactionModel.CategoryAmountRank]
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
	// 获取category
	var categoryList dataType.Slice[uint, categoryModel.Category]
	err = global.GvaDb.Where("id IN (?)", categoryIds).Find(&categoryList).Error
	if responseError(err, ctx) {
		return
	}
	categoryMap := categoryList.ToMap(
		func(category categoryModel.Category) uint {
			return category.ID
		},
	)
	// 处理响应
	responseData := make([]response.TransactionCategoryAmountRank, len(rankingList), requestData.Limit)
	for i, rank := range rankingList {
		responseData[i].Amount = rank.Amount
		responseData[i].Count = rank.Count
		err = responseData[i].Category.SetData(categoryMap[rank.CategoryId])
		if responseError(err, ctx) {
			return
		}
	}
	//数量不足时补足响应数量
	if len(rankingList) < requestData.Limit {
		categoryList = []categoryModel.Category{}
		limit := requestData.Limit - len(rankingList)
		db := global.GvaDb.Where("account_id = ?", account.ID)
		db = db.Where("income_expense = ?", requestData.IncomeExpense)
		err = db.Where("id NOT IN (?)", categoryIds).Limit(limit).Find(&categoryList).Error
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
