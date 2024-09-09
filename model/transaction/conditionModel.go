package transactionModel

import (
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	"KeepAccount/util/timeTool"
	"gorm.io/gorm"
	"time"
)

// ForeignKeyCondition 交易外键查询条件 用于交易记录和统计的查询
type ForeignKeyCondition struct {
	AccountId   uint
	UserIds     *[]uint
	CategoryIds *[]uint
}

func (f *ForeignKeyCondition) addConditionToQuery(db *gorm.DB) *gorm.DB {
	query := db.Where("account_id = ?", f.AccountId)
	if f.UserIds != nil {
		query = query.Where("user_id IN (?)", *f.UserIds)
	}
	if f.CategoryIds != nil {
		query = query.Where("category_id IN (?)", *f.CategoryIds)
	}
	return query
}

// GetStatisticTableName 根据查询条件返回合适的查询表格
// gorm的Model方法视乎有问题 只会在第一次执行Model时更新查询的表格 故返回表名而不是模型
func (f *ForeignKeyCondition) GetStatisticTableName(ie constant.IncomeExpense) string {
	var model statisticModel
	if ie == constant.Income {
		model = f.getIncomeStatisticModel()
	} else {
		model = f.getExpendStatisticModel()
	}
	if model == nil {
		return "transaction"
	}
	return model.TableName()
}

func (f *ForeignKeyCondition) getIncomeStatisticModel() statisticModel {
	if f.CategoryIds == nil {
		if f.UserIds == nil {
			return &IncomeAccountStatistic{}
		} else {
			return &IncomeAccountUserStatistic{}
		}
	} else if f.UserIds == nil {
		return &IncomeCategoryStatistic{}
	} else {
		return &IncomeAccountUserStatistic{}
	}
}

func (f *ForeignKeyCondition) getExpendStatisticModel() statisticModel {
	if f.CategoryIds == nil {
		if f.UserIds == nil {
			return &ExpenseAccountStatistic{}
		} else {
			return &ExpenseAccountUserStatistic{}
		}
	} else if f.UserIds == nil {
		return &ExpenseCategoryStatistic{}
	} else {
		return &ExpenseAccountUserStatistic{}
	}
}

// Condition 交易记录查询条件 用于交易记录和统计的查询
type Condition struct {
	IncomeExpense *constant.IncomeExpense
	ForeignKeyCondition
	TimeCondition
	ExtensionCondition
}

func (c *Condition) addConditionToQuery(db *gorm.DB) *gorm.DB {
	query := c.ForeignKeyCondition.addConditionToQuery(db)
	query = c.TimeCondition.addConditionToQuery(query)
	query = c.ExtensionCondition.addConditionToQuery(query)
	if c.IncomeExpense != nil {
		query = query.Where("income_expense = ?", *c.IncomeExpense)
	}
	return query
}

// TimeCondition 交易表时间查询条件
type TimeCondition struct {
	TradeStartTime *time.Time
	TradeEndTime   *time.Time
}

func NewTimeCondition() *TimeCondition {
	return &TimeCondition{}
}

func (tc *TimeCondition) SetTradeTimes(startTime, endTime time.Time) {
	tc.TradeStartTime = &startTime
	tc.TradeEndTime = &endTime
}

func (tc *TimeCondition) addConditionToQuery(query *gorm.DB) *gorm.DB {
	if tc.TradeStartTime != nil {
		query = query.Where("trade_time >= ?", *tc.TradeStartTime)
	}
	if tc.TradeEndTime != nil {
		query = query.Where("trade_time <= ?", *tc.TradeEndTime)
	}
	return query
}

// ExtensionCondition 拓展查询条件 多是无索引条件
type ExtensionCondition struct {
	MinAmount, MaxAmount *int
}

func (ec *ExtensionCondition) IsSet() bool {
	return ec != nil && (ec.MinAmount != nil || ec.MaxAmount != nil)
}

func (ec *ExtensionCondition) addConditionToQuery(query *gorm.DB) *gorm.DB {
	if ec.MinAmount != nil {
		query = query.Where("amount >= ?", *ec.MinAmount)
	}
	if ec.MaxAmount != nil {
		query = query.Where("amount <= ?", *ec.MaxAmount)
	}
	return query
}

// StatisticCondition 交易的统计查询条件
type StatisticCondition struct {
	ForeignKeyCondition
	StartTime time.Time
	EndTime   time.Time

	accountId uint
	location  *time.Location
}

func (s *StatisticCondition) getLocation() *time.Location {
	if s.accountId == s.AccountId && s.location != nil {
		return s.location
	}
	var err error
	s.location, err = time.LoadLocation(accountModel.NewDao().GetLocation(s.AccountId))
	if err != nil {
		panic(err)
	}
	s.accountId = s.AccountId
	return s.location
}

// addConditionToQuery 通过条件获取附带查询条件的gorm.db
func (s *StatisticCondition) addConditionToQuery(db *gorm.DB) *gorm.DB {
	query := s.ForeignKeyCondition.addConditionToQuery(db)
	query = query.Where("date BETWEEN ? AND ?", timeTool.ToDay(s.StartTime.In(s.getLocation())),
		timeTool.ToDay(s.EndTime.In(s.getLocation())))
	return query
}

func (s *StatisticCondition) CheckAvailability() bool {
	if s.UserIds != nil && len(*s.UserIds) == 0 {
		return false
	}
	if s.CategoryIds != nil && len(*s.CategoryIds) == 0 {
		return false
	}
	return true
}

// StatisticConditionBuilder 是用于构建 StatisticCondition 的构建器
type StatisticConditionBuilder struct {
	condition StatisticCondition
}

// NewStatisticConditionBuilder 返回一个新的 StatisticConditionBuilder 实例
func NewStatisticConditionBuilder(accountId uint) *StatisticConditionBuilder {
	return &StatisticConditionBuilder{
		condition: StatisticCondition{
			ForeignKeyCondition: ForeignKeyCondition{AccountId: accountId},
		},
	}
}

// WithUserIds 设置用户ids
func (b *StatisticConditionBuilder) WithUserIds(Ids []uint) *StatisticConditionBuilder {
	b.condition.UserIds = &Ids
	return b
}

// WithCategoryIds 设置交易类型ids
func (b *StatisticConditionBuilder) WithCategoryIds(Ids []uint) *StatisticConditionBuilder {
	b.condition.CategoryIds = &Ids
	return b
}

// WithDate 设置时间范围
func (b *StatisticConditionBuilder) WithDate(startTime, endTime time.Time) *StatisticConditionBuilder {
	b.condition.StartTime = startTime
	b.condition.EndTime = endTime
	return b
}

// Build 构建 StatisticCondition 实例
func (b *StatisticConditionBuilder) Build() *StatisticCondition {
	return &b.condition
}
