package response

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	"KeepAccount/global/db"
	accountModel "KeepAccount/model/account"
	categoryModel "KeepAccount/model/category"
	transactionModel "KeepAccount/model/transaction"
	userModel "KeepAccount/model/user"
	"KeepAccount/util/dataTool"
	"time"
)

// TransactionDetail 交易详情
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
	TradeTime          time.Time
	UpdateTime         time.Time
	CreateTime         time.Time
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
	t.TradeTime = trans.TradeTime
	t.UpdateTime = trans.UpdatedAt
	t.CreateTime = trans.CreatedAt
	return nil
}

// TransactionDetailList 交易详情列表
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
		(*t)[i].TradeTime = transList[i].TradeTime
		(*t)[i].UpdateTime = transList[i].UpdatedAt
		(*t)[i].CreateTime = transList[i].CreatedAt
	}

	// 用户
	var userList []userModel.User
	err := db.Db.Select("username,id").Where("id IN (?)", userIds).Find(&userList).Error
	if err != nil {
		return err
	}
	userMap := make(map[uint]userModel.User)
	for _, item := range userList {
		userMap[item.ID] = item
	}
	// 账本
	var accountList []accountModel.Account
	err = db.Db.Select("name", "id").Where("id IN (?)", accountIds).Find(&accountList).Error
	if err != nil {
		return err
	}
	accountMap := make(map[uint]accountModel.Account)
	for _, item := range accountList {
		accountMap[item.ID] = item
	}
	// 二级交易类型
	var categoryList []categoryModel.Category
	err = db.Db.Select("icon", "name", "father_id", "id").Where(
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
	err = db.Db.Select("name", "id").Where("id IN (?)", fatherIds).Find(&fatherList).Error
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
	global.IEStatistic
}

type TransactionStatistic struct {
	global.IEStatistic
	StartTime time.Time
	EndTime   time.Time
}

type TransactionDayStatistic struct {
	global.AmountCount
	Date time.Time
}

type TransactionCategoryAmountRank struct {
	Category CategoryOne
	global.AmountCount
}

type TransactionTimingConfig struct {
	Id, AccountId, UserId uint
	Type                  transactionModel.TimingType
	OffsetDays            int
	NextTime              time.Time
	Username              string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TransactionTiming struct {
	Trans  TransactionInfo
	Config TransactionTimingConfig
}

func (tt *TransactionTiming) SetData(data transactionModel.Timing) error {
	name, err := userModel.NewDao().PluckNameById(data.UserId)
	if err != nil {
		return err
	}
	err = tt.Trans.SetData(data.TransInfo)
	if err != nil {
		return err
	}
	tt.Config = TransactionTimingConfig{
		Id:         data.ID,
		UserId:     data.UserId,
		AccountId:  data.AccountId,
		Type:       data.Type,
		OffsetDays: data.OffsetDays,
		NextTime:   data.NextTime,
		Username:   name,
		CreatedAt:  data.CreatedAt,
		UpdatedAt:  data.UpdatedAt,
	}
	return nil
}

type TransactionTimingList []TransactionTiming

func (ttl *TransactionTimingList) SetData(list dataTool.Slice[uint, transactionModel.Timing]) error {
	*ttl = make(TransactionTimingList, len(list), len(list))
	if len(*ttl) == 0 {
		return nil
	}
	nameMap, err := getUsernameMap(list.ExtractValues(func(timing transactionModel.Timing) uint { return timing.ID }))
	transList := make(dataTool.Slice[uint, transactionModel.Info], len(list), len(list))
	for i, timing := range list {
		transList[i] = timing.TransInfo
		(*ttl)[i].Config = TransactionTimingConfig{
			Id:         timing.ID,
			UserId:     timing.UserId,
			AccountId:  timing.AccountId,
			Type:       timing.Type,
			OffsetDays: timing.OffsetDays,
			NextTime:   timing.NextTime,
			Username:   nameMap[timing.UserId],
			CreatedAt:  timing.CreatedAt,
			UpdatedAt:  timing.UpdatedAt,
		}
	}
	var infoList TransactionInfoList
	err = infoList.SetData(transList)
	if err != nil {
		return err
	}
	for i, info := range infoList {
		(*ttl)[i].Trans = info
	}
	return nil
}

// 交易详情

type TransactionInfo struct {
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
	TradeTime          time.Time
}

func (ti *TransactionInfo) SetData(data transactionModel.Info) error {
	var (
		username string
		account  accountModel.Account
		category categoryModel.Category
		father   categoryModel.Father
		err      error
	)
	account, err = accountModel.NewDao().SelectById(data.AccountId)
	if err != nil {
		return err
	}
	username, err = userModel.NewDao().PluckNameById(data.UserId)
	if err != nil {
		return err
	}
	category, err = categoryModel.NewDao().SelectById(data.CategoryId)
	if err != nil {
		return err
	}
	if father, err = category.GetFather(); err != nil {
		return err
	}

	ti.UserId = data.UserId
	ti.UserName = username
	ti.AccountId = account.ID
	ti.AccountName = account.Name
	ti.Amount = data.Amount
	ti.CategoryId = data.CategoryId
	ti.CategoryIcon = category.Icon
	ti.CategoryName = category.Name
	ti.CategoryFatherName = father.Name
	ti.IncomeExpense = category.IncomeExpense
	ti.Remark = category.Icon
	ti.TradeTime = data.TradeTime
	return nil
}

type TransactionInfoList []TransactionInfo

func (t *TransactionInfoList) SetData(list dataTool.Slice[uint, transactionModel.Info]) error {
	*t = make([]TransactionInfo, len(list), len(list))
	if len(list) == 0 {
		return nil
	}

	userMap, err := getUsernameMap(list.ExtractValues(func(info transactionModel.Info) uint { return info.UserId }))
	if err != nil {
		return err
	}
	accountMap, err := getAccountNameMap(list.ExtractValues(func(info transactionModel.Info) uint { return info.AccountId }))
	if err != nil {
		return err
	}
	categoryMap, fatherMap, err := t.getCategoryMap(list)
	if err != nil {
		return err
	}
	for i, data := range list {
		category := categoryMap[data.CategoryId]
		(*t)[i] = TransactionInfo{
			Id:                 0,
			UserId:             data.UserId,
			UserName:           userMap[data.UserId],
			AccountId:          data.AccountId,
			AccountName:        accountMap[data.AccountId],
			Amount:             data.Amount,
			CategoryId:         data.CategoryId,
			CategoryIcon:       category.Icon,
			CategoryName:       category.Name,
			CategoryFatherName: fatherMap[category.FatherId].Name,
			IncomeExpense:      data.IncomeExpense,
			Remark:             data.Remark,
			TradeTime:          data.TradeTime,
		}
	}
	return nil
}

func (t *TransactionInfoList) getCategoryMap(list dataTool.Slice[uint, transactionModel.Info]) (
	categoryMap map[uint]categoryModel.Category, fatherMap map[uint]categoryModel.Father, err error,
) {
	var categoryList dataTool.Slice[uint, categoryModel.Category]
	ids := list.ExtractValues(func(info transactionModel.Info) uint { return info.CategoryId })
	err = db.Db.Select("icon", "name", "father_id", "id").Where("id IN (?)", ids).Find(&categoryList).Error
	if err != nil {
		return
	}

	var fatherList dataTool.Slice[uint, categoryModel.Father]
	ids = categoryList.ExtractValues(func(category categoryModel.Category) uint { return category.FatherId })
	err = db.Db.Select("name", "id").Where("id IN (?)", ids).Find(&fatherList).Error
	if err != nil {
		return
	}
	categoryMap = categoryList.ToMap(func(category categoryModel.Category) uint { return category.ID })
	fatherMap = fatherList.ToMap(func(father categoryModel.Father) uint { return father.ID })
	return
}
