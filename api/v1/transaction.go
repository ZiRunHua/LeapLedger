package v1

import (
	"KeepAccount/api/request"
	"KeepAccount/api/response"
	"KeepAccount/global"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util"
	"fmt"
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
		fmt.Println("Field not found")
		return nil
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
