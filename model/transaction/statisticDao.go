package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	"errors"
	"gorm.io/gorm"
	"time"
)

type StatisticDao struct {
	db *gorm.DB
}

func NewStatisticDao(db ...*gorm.DB) *StatisticDao {
	if len(db) > 0 && db[0] != nil {
		return &StatisticDao{db: db[0]}
	}
	return &StatisticDao{db: global.GvaDb}
}

func (s *StatisticDao) query(ie constant.IncomeExpense, db *gorm.DB) *gorm.DB {
	if ie == constant.Expense {
		return db.Model(&ExpenseCategoryStatistic{})
	} else {
		return db.Model(&IncomeCategoryStatistic{})
	}
}

// DayStatistic StatisticDao.GetDayStatisticByCondition 方法返回
type DayStatistic struct {
	global.AmountCount
	Date time.Time
}

func (s *StatisticDao) GetDayStatisticByCondition(
	ie constant.IncomeExpense, condition StatisticCondition,
) (result []DayStatistic, err error) {
	if false == condition.CheckAvailability() {
		return
	}
	query := condition.addConditionToQuery(s.db)
	query.Select("SUM(amount) as Amount,SUM(count) as Count,date").Group("date")
	err = query.Table(condition.GetStatisticTableName(ie)).Find(&result).Error
	return result, err
}

func (s *StatisticDao) GetTotalByCondition(
	ie constant.IncomeExpense, condition StatisticCondition,
) (result global.AmountCount, err error) {
	if false == condition.CheckAvailability() {
		return
	}
	query := condition.addConditionToQuery(s.db)
	query.Select("SUM(amount) as Amount,SUM(count) as Count")
	err = query.Table(condition.GetStatisticTableName(ie)).Find(&result).Error
	return result, err
}

// CategoryAmountRankCondition StatisticDao.GetCategoryAmountRank查询条件
type CategoryAmountRankCondition struct {
	Account   accountModel.Account
	StartTime time.Time
	EndTime   time.Time
}

func (c *CategoryAmountRankCondition) Local() {
	location := accountModel.NewDao().GetTimeLocation(c.Account.ID)
	c.StartTime = c.StartTime.In(location)
	c.EndTime = c.EndTime.In(location)
}

// CategoryAmountRank  StatisticDao.GetCategoryAmountRank查询结果
type CategoryAmountRank struct {
	CategoryId uint
	global.AmountCount
}

func (s *StatisticDao) GetCategoryAmountRank(
	ie constant.IncomeExpense, condition CategoryAmountRankCondition, limit *int,
) (result []CategoryAmountRank, err error) {
	condition.Local()
	query := s.db.Where("account_id = ?", condition.Account.ID)
	query = query.Where("date BETWEEN ? AND ?", condition.StartTime, condition.EndTime)

	query = query.Select("SUM(amount) as Amount,SUM(count) as Count,category_id").Group("category_id")
	if limit != nil {
		query = query.Limit(*limit)
	}
	err = s.query(ie, query).Order("Amount desc").Find(&result).Error
	return result, err
}

// GetIeStatisticByCondition 查询收支统计 返回 global.IEStatistic
func (s *StatisticDao) GetIeStatisticByCondition(ie *constant.IncomeExpense, condition StatisticCondition) (
	result global.IEStatistic, err error,
) {
	if false == condition.CheckAvailability() {
		return result, errors.New("查询条件错误")
	}
	query := condition.addConditionToQuery(s.db)
	if ie.QueryIncome() {
		err = query.Table(condition.GetStatisticTableName(constant.Income)).Select("SUM(amount) as amount,SUM(count) as count").Scan(&result.Income).Error
		if err != nil {
			return
		}
	}
	if ie.QueryExpense() {
		err = query.Table(condition.GetStatisticTableName(constant.Expense)).Select("SUM(amount) as amount,SUM(count) as count").Scan(&result.Expense).Error
		if err != nil {
			return
		}
	}
	return result, err
}

func (s *StatisticDao) GetAmountCountByCondition(condition StatisticCondition, ie constant.IncomeExpense) (
	result global.AmountCount, err error,
) {
	if false == condition.CheckAvailability() {
		return
	}
	query := condition.addConditionToQuery(s.db).Table(condition.GetStatisticTableName(ie))
	err = query.Select("SUM(amount) as amount,SUM(count) as count").Scan(&result).Error
	return result, err
}
