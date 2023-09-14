package response

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
)

func TransactionModelToResponse(trans transactionModel.Transaction) TransactionOne {
	return TransactionOne{
		Id:            trans.ID,
		UserId:        trans.UserId,
		AccountId:     trans.AccountId,
		Amount:        trans.Amount,
		CategoryId:    trans.CategoryId,
		IncomeExpense: trans.IncomeExpense,
		Remark:        trans.Remark,
		TradeTime:     trans.TradeTime.Unix(),
		UpdateTime:    trans.UpdatedAt.Unix(),
		CreateTime:    trans.CreatedAt.Unix(),
	}
}

type TransactionOne struct {
	Id            uint
	UserId        uint
	AccountId     uint
	Amount        int
	CategoryId    uint
	IncomeExpense constant.IncomeExpense
	Remark        string
	TradeTime     int64
	UpdateTime    int64
	CreateTime    int64
}

// 交易详情
type TransactionDetail struct {
	Id                 uint
	UserId             uint
	UserName           string
	AccountId          uint
	AccountName        string
	Amount             int
	CategoryId         uint
	CategoryIcon       string
	CategoryName       string
	CategoryFatherName string
	IncomeExpense      constant.IncomeExpense
	Remark             string
	TradeTime          int64
	UpdateTime         int64
	CreateTime         int64
}

func (t *TransactionDetail) SetData(
	trans transactionModel.Transaction, account *accountModel.Account,
) error {
	var (
		user     userModel.User
		category categoryModel.Category
		father   categoryModel.Father
		err      error
	)
	if account == nil {
		account = &accountModel.Account{}
		if *account, err = trans.GetAccount(); err != nil {
			return err
		}
	}
	if user, err = trans.GetUser("username", "id"); err != nil {
		return err
	}
	if category, err = trans.GetCategory(); err != nil {
		return err
	}
	if father, err = category.GetFather(); err != nil {
		return err
	}
	t.Id = trans.ID
	t.UserId = user.ID
	t.UserName = user.Username
	t.AccountId = account.ID
	t.AccountName = account.Name
	t.Amount = trans.Amount
	t.CategoryId = trans.CategoryId
	t.CategoryIcon = category.Icon
	t.CategoryName = category.Name
	t.CategoryFatherName = father.Name
	t.IncomeExpense = category.IncomeExpense
	t.Remark = category.Icon
	t.TradeTime = trans.TradeTime.Unix()
	t.UpdateTime = trans.UpdatedAt.Unix()
	t.CreateTime = trans.CreatedAt.Unix()
	return nil
}

// 交易详情列表
type TransactionDetailList []TransactionDetail

func (t *TransactionDetailList) SetData(transList []transactionModel.Transaction) error {
	*t = make([]TransactionDetail, len(transList), len(transList))
	if len(transList) == 0 {
		return nil
	}
	userIds := make([]uint, len(transList), len(transList))
	accountIds := make([]uint, len(transList), len(transList))
	categoryIds := make([]uint, len(transList), len(transList))
	for i := 0; i < len(transList); i++ {
		userIds[i] = transList[i].UserId
		accountIds[i] = transList[i].AccountId
		categoryIds[i] = transList[i].CategoryId

		(*t)[i].Id = transList[i].ID
		(*t)[i].UserId = transList[i].UserId
		(*t)[i].AccountId = transList[i].AccountId
		(*t)[i].Amount = transList[i].Amount
		(*t)[i].CategoryId = transList[i].CategoryId
		(*t)[i].IncomeExpense = transList[i].IncomeExpense
		(*t)[i].Remark = transList[i].Remark
		(*t)[i].TradeTime = transList[i].TradeTime.Unix()
		(*t)[i].UpdateTime = transList[i].UpdatedAt.Unix()
		(*t)[i].CreateTime = transList[i].CreatedAt.Unix()
	}

	// 用户
	var userList []userModel.User
	err := global.GvaDb.Select("username,id").Where("id IN (?)", userIds).Find(&userList).Error
	if err != nil {
		return err
	}
	userMap := make(map[uint]userModel.User)
	for _, item := range userList {
		userMap[item.ID] = item
	}
	// 账本
	var accountList []accountModel.Account
	err = global.GvaDb.Select("name", "id").Where("id IN (?)", accountIds).Find(&accountList).Error
	if err != nil {
		return err
	}
	accountMap := make(map[uint]accountModel.Account)
	for _, item := range accountList {
		accountMap[item.ID] = item
	}
	// 二级交易类型
	var categoryList []categoryModel.Category
	err = global.GvaDb.Select("icon", "name", "father_id", "id").Where(
		"id IN (?)", categoryIds,
	).Find(&categoryList).Error
	if err != nil {
		return err
	}
	categoryMap := make(map[uint]categoryModel.Category)
	fatherIds := make([]uint, len(categoryList), len(categoryList))
	for _, item := range categoryList {
		categoryMap[item.ID] = item
		fatherIds = append(fatherIds, item.FatherId)
	}
	// 一级交易类型
	var fatherList []categoryModel.Father
	err = global.GvaDb.Select("name", "id").Where("id IN (?)", fatherIds).Find(&fatherList).Error
	if err != nil {
		return err
	}
	fatherMap := make(map[uint]categoryModel.Father)
	for _, item := range fatherList {
		fatherMap[item.ID] = item
	}

	for i, trans := range transList {
		category := categoryMap[trans.CategoryId]
		(*t)[i].UserName = userMap[trans.UserId].Username
		(*t)[i].AccountName = accountMap[transList[i].AccountId].Name
		(*t)[i].CategoryIcon = category.Icon
		(*t)[i].CategoryName = category.Name
		(*t)[i].CategoryFatherName = fatherMap[category.FatherId].Name
	}
	return nil
}

type TransactionGetList struct {
	List TransactionDetailList
	PageData
}
type TransactionTotal struct {
	global.IncomeExpenseStatistic
}

type TransactionStatistic struct {
	global.IncomeExpenseStatistic
	StartTime int64
	EndTime   int64
}

type TransactionMonthStatistic struct {
	List []TransactionStatistic
}

type TransactionDayStatistic struct {
	global.AmountCount
	Date int64
}

type TransactionCategoryAmountRank struct {
	Category CategoryOne
	global.AmountCount
}
