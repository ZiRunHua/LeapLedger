package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	"KeepAccount/model/common/query"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"reflect"
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

func (t *TransactionApi) CreateOne(ctx *gin.Context) {
	var requestData request.TransactionCreateOne
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		response.FailToParameter(ctx, err)
		return
	}
	pass, account := checkFunc.AccountBelong(requestData.AccountId, ctx)
	if false == pass {
		return
	}
	user, err := contextFunc.GetUser(ctx)
	if err != nil {
		response.FailToError(ctx, err)
		return
	}
	transaction := &transactionModel.Transaction{
		AccountId:     requestData.AccountId,
		CategoryId:    requestData.CategoryId,
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
	if responseError(err, ctx) {
		return
	}

	var responseData *response.TransactionDetail
	if responseData, err = t.getResponseDetail(*transaction, account); responseError(err, ctx) {
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
	pass, account := checkFunc.AccountBelong(requestData.AccountId, ctx)
	if false == pass {
		return
	}
	transaction := &transactionModel.Transaction{
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
			transaction.SetTx(tx)
			return transactionService.Update(transaction)
		},
	)
	if responseError(err, ctx) {
		return
	}

	var responseData *response.TransactionDetail
	if responseData, err = t.getResponseDetail(*transaction, account); responseError(err, ctx) {
		return
	}
	response.OkWithData(responseData, ctx)
}
func (t *TransactionApi) getResponseDetail(
	trans transactionModel.Transaction, account *accountModel.Account,
) (*response.TransactionDetail, error) {
	var (
		user     *userModel.User
		category *categoryModel.Category
		father   *categoryModel.Father
		err      error
	)
	if account == nil {
		if account, err = trans.GetAccount(); err != nil {
			return nil, err
		}
	}
	if user, err = trans.GetUser(); err != nil {
		return nil, err
	}
	if category, err = trans.GetCategory(); err != nil {
		return nil, err
	}
	if father, err = category.GetFather(); err != nil {
		return nil, err
	}
	return &response.TransactionDetail{
		Id:                 trans.ID,
		UserId:             user.ID,
		UserName:           user.Username,
		AccountId:          account.ID,
		AccountName:        account.Name,
		Amount:             trans.Amount,
		CategoryId:         trans.CategoryId,
		CategoryIcon:       category.Icon,
		CategoryName:       category.Name,
		CategoryFatherName: father.Name,
		IncomeExpense:      category.IncomeExpense,
		Remark:             category.Icon,
		TradeTime:          trans.TradeTime.Unix(),
		UpdateTime:         trans.UpdatedAt.Unix(),
		CreateTime:         trans.CreatedAt.Unix(),
	}, nil
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

	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}

	// 设置查询条件
	condition := &transactionModel.TransactionCondition{
		// 交易外键条件
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId:   &requestData.AccountId,
			UserIds:     requestData.UserIds,
			CategoryIds: requestData.CategoryIds,
		},
		// 交易时间条件
		TradeTimeCondition: transactionModel.TradeTimeCondition{
			TradeStartTime: request.GetTimeByTimestamp(&requestData.StartTime),
			TradeEndTime:   request.GetTimeByTimestamp(&requestData.EndTime),
		},
		IncomeExpense: requestData.IncomeExpense,
		MinimumAmount: requestData.MinimumAmount,
		MaximumAmount: requestData.MaximumAmount,
	}
	// 查询并获取结果
	transactionList, err := transactionModel.Dao.NewTransaction(nil).GetListByCondition(
		condition,
		requestData.Limit,
		requestData.Offset,
	)
	if responseError(err, ctx) {
		return
	}
	var responseData response.TransactionGetList
	if len(transactionList) > 0 {
		responseData = response.TransactionGetList{List: t.getResponseDetailList(transactionList)}
	} else {
		responseData = response.TransactionGetList{List: make([]response.TransactionDetail, 0)}
	}
	response.OkWithData(responseData, ctx)
}

func (t *TransactionApi) getResponseDetailList(transList []transactionModel.Transaction) []response.TransactionDetail {
	// 用户
	Ids := t.getFieldValues(transList, "UserId")
	var userList []userModel.User
	global.GvaDb.Select("username,id").Where("id IN (?)", Ids).Find(&userList)
	userMap := make(map[uint]userModel.User)
	for _, item := range userList {
		userMap[item.ID] = item
	}
	// 账本
	Ids = t.getFieldValues(transList, "AccountId")
	var accountList []accountModel.Account
	global.GvaDb.Select("name", "id").Where("id IN (?)", Ids).Find(&accountList)
	accountMap := make(map[uint]accountModel.Account)
	for _, item := range accountList {
		accountMap[item.ID] = item
	}
	// 二级交易类型
	Ids = t.getFieldValues(transList, "CategoryId")
	var categoryList []categoryModel.Category
	global.GvaDb.Select("icon", "name", "father_id", "id").Where("id IN (?)", Ids).Find(&categoryList)
	categoryMap := make(map[uint]categoryModel.Category)
	fatherIds := []uint{}
	for _, item := range categoryList {
		categoryMap[item.ID] = item
		fatherIds = append(fatherIds, item.FatherId)
	}
	// 一级交易类型
	var fatherList []categoryModel.Father
	global.GvaDb.Select("name", "id").Where("id IN (?)", fatherIds).Find(&fatherList)
	fatherMap := make(map[uint]categoryModel.Father)
	for _, item := range fatherList {
		fatherMap[item.ID] = item
	}

	result := make([]response.TransactionDetail, len(transList), len(transList))
	for i, trans := range transList {
		category := categoryMap[trans.CategoryId]
		result[i] = response.TransactionDetail{
			Id:                 trans.ID,
			UserId:             trans.UserId,
			UserName:           userMap[trans.UserId].Username,
			AccountId:          trans.AccountId,
			AccountName:        accountMap[trans.AccountId].Name,
			Amount:             trans.Amount,
			CategoryId:         trans.CategoryId,
			CategoryIcon:       category.Icon,
			CategoryName:       category.Name,
			CategoryFatherName: fatherMap[category.FatherId].Name,
			IncomeExpense:      trans.IncomeExpense,
			Remark:             trans.Remark,
			TradeTime:          trans.TradeTime.Unix(),
			UpdateTime:         trans.UpdatedAt.Unix(),
			CreateTime:         trans.CreatedAt.Unix(),
		}

	}
	return result
}
func (t *TransactionApi) getFieldValues(transList []transactionModel.Transaction, fieldName string) []interface{} {
	var fieldValues []interface{}

	structType := reflect.TypeOf(transList[0])
	field, found := structType.FieldByName(fieldName)
	if !found {
		panic("TransactionApi.getFieldValues:Field not found")
	}

	fieldIndex := field.Index[0]

	for _, p := range transList {
		val := reflect.ValueOf(p)
		fieldValue := val.Field(fieldIndex).Interface()
		fieldValues = append(fieldValues, fieldValue)
	}

	return fieldValues
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
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}
	// 设置查询条件
	condition := &transactionModel.StatisticCondition{
		// 交易外键条件
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId:   &requestData.AccountId,
			UserIds:     requestData.UserIds,
			CategoryIds: requestData.CategoryIds,
		},
		IncomeExpense: requestData.IncomeExpense,
		MinimumAmount: requestData.MinimumAmount,
		MaximumAmount: requestData.MaximumAmount,
	}
	// 处理查询时间
	var startTime, endTime time.Time
	startTime = *request.GetTimeByTimestamp(&requestData.StartTime)
	endTime = *request.GetTimeByTimestamp(&requestData.EndTime)
	// 查询并处理响应
	total, err := transactionModel.Dao.NewTransaction(nil).GetStatisticByCondition(condition, startTime, endTime)
	if responseError(err, ctx) {
		return
	}
	response.OkWithData(response.TransactionTotal{*total}, ctx)
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
	if pass, _ := checkFunc.AccountBelong(requestData.AccountId, ctx); pass == false {
		return
	}

	// 设置查询条件
	condition := &transactionModel.StatisticCondition{
		// 交易外键条件
		ForeignKeyCondition: transactionModel.ForeignKeyCondition{
			AccountId:   &requestData.AccountId,
			UserIds:     requestData.UserIds,
			CategoryIds: requestData.CategoryIds,
		},
		IncomeExpense: requestData.IncomeExpense,
		MinimumAmount: requestData.MinimumAmount,
		MaximumAmount: requestData.MaximumAmount,
	}
	// 处理查询时间
	var startTime, endTime time.Time
	startTime = *request.GetTimeByTimestamp(&requestData.StartTime)
	endTime = *request.GetTimeByTimestamp(&requestData.EndTime)
	months := util.Time.SplitMonths(startTime, endTime)
	// 查询并处理响应
	responseList := make([]response.TransactionStatistic, 0, len(months))
	dao := transactionModel.Dao.NewTransaction(nil)
	for i := len(months) - 1; i >= 0; i-- {
		monthStartTime := months[i]
		monthEndTime := util.Time.GetLastSecondOfMonth(monthStartTime)

		monthStatistic, err := dao.GetStatisticByCondition(condition, monthStartTime, monthEndTime)

		if responseError(err, ctx) {
			return
		}
		responseList = append(
			responseList, response.TransactionStatistic{
				IncomeExpenseStatistic: *monthStatistic,
				StartTime:              monthStartTime.Unix(),
				EndTime:                monthEndTime.Unix(),
			},
		)
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
	pass, account := checkFunc.AccountBelong(requestData.AccountId, ctx)
	if pass == false {
		return
	}
	// 处理请求
	startTime, endTime := requestData.FormatDayTime()
	days := util.Time.SplitDays(startTime, endTime)
	dayMap := make(map[time.Time]*response.TransactionDayStatistic, len(days))
	condition := transactionModel.DayStatisticCondition{
		Account:     *account,
		CategoryIds: requestData.CategoryIds,
		StartTime:   startTime,
		EndTime:     endTime,
	}
	handleFunc := func(ie constant.IncomeExpense) error {
		statistics, err := transactionModel.Dao.NewStatisticDao(nil).GetDayStatisticByCondition(ie, condition)
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
	pass, account := checkFunc.AccountBelong(requestData.AccountId, ctx)
	if pass == false {
		return
	}
	// 处理查询
	startTime, endTime := requestData.FormatDayTime()
	condition := transactionModel.CategoryAmountRankCondition{
		Account:   *account,
		StartTime: startTime,
		EndTime:   endTime,
	}
	rankingList, err := transactionModel.Dao.NewStatisticDao(nil).GetCategoryAmountRank(
		requestData.IncomeExpense, condition, requestData.Limit,
	)
	if responseError(err, ctx) {
		return
	}
	// 处理响应
	var category *categoryModel.Category
	responseData := make([]response.TransactionCategoryAmountRank, len(rankingList), requestData.Limit)
	categoryIds := []uint{}
	for i, rank := range rankingList {
		responseData[i].Amount = rank.Amount
		responseData[i].Count = rank.Count
		category, err = query.FirstByPrimaryKey[*categoryModel.Category](rank.CategoryId)
		categoryIds = append(categoryIds, rank.CategoryId)
		if responseError(err, ctx) {
			return
		}
		responseData[i].Category = *response.CategoryModelToResponse(category)
	}
	//数量不足时补足响应数量
	if len(rankingList) < requestData.Limit {
		categoryList := []categoryModel.Category{}
		limit := requestData.Limit - len(rankingList)
		query := global.GvaDb
		query = query.Where("account_id = ?", account.ID)
		query = query.Where("income_expense = ?", requestData.IncomeExpense)
		err = query.Where("id NOT IN (?)", categoryIds).Limit(limit).Find(&categoryList).Error
		if responseError(err, ctx) {
			return
		}
		for _, c := range categoryList {
			responseData = append(
				responseData, response.TransactionCategoryAmountRank{
					Category: *response.CategoryModelToResponse(&c),
				},
			)
		}
	}
	response.OkWithData(
		response.List[response.TransactionCategoryAmountRank]{List: responseData}, ctx,
	)
}
