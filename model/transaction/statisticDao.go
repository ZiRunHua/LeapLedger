package transactionModel

import (
	"KeepAccount/global"
	"KeepAccount/global/constant"
	accountModel "KeepAccount/model/account"
	"gorm.io/gorm"
	"time"
)

type StatisticDao struct {
	db *gorm.DB
}

func (d *dao) NewStatisticDao(db *gorm.DB) *StatisticDao {
	if db == nil {
		db = global.GvaDb
	}
	return &StatisticDao{db}
}

type DayStatisticCondition struct {
	Account     accountModel.Account
	CategoryIds *[]uint
	StartTime   time.Time
	EndTime     time.Time
}

type DayStatistic struct {
	global.AmountCount
	Date time.Time
}

func (s *StatisticDao) query(ie constant.IncomeExpense, db *gorm.DB) *gorm.DB {
	if ie == constant.Expense {
		return db.Model(&ExpenseStatistic{})
	} else {
		return db.Model(&IncomeStatistic{})
	}
}

func (s *StatisticDao) GetDayStatisticByCondition(
	ie constant.IncomeExpense, condition DayStatisticCondition,
) (result []DayStatistic, err error) {
	query := s.db.Where("account_id = ?", condition.Account.ID)
	query = query.Where("date BETWEEN ? AND ?", condition.StartTime, condition.EndTime)
	if condition.CategoryIds != nil {
		query = query.Where("category_id IN (?)", *condition.CategoryIds)
	}

	query = query.Select("SUM(amount) as Amount,SUM(count) as Count,date").Group("date")
	err = s.query(ie, query).Find(&result).Error
	return result, err
}

// CategoryAmountRankCondition GetCategoryAmountRank查询条件
type CategoryAmountRankCondition struct {
	Account   accountModel.Account
	StartTime time.Time
	EndTime   time.Time
}

// CategoryAmountRankCondition GetCategoryAmountRank查询结果
type CategoryAmountRank struct {
	CategoryId uint
	global.AmountCount
}

func (s *StatisticDao) GetCategoryAmountRank(
	ie constant.IncomeExpense, condition CategoryAmountRankCondition, limit int,
) (result []CategoryAmountRank, err error) {
	query := s.db.Where("account_id = ?", condition.Account.ID)
	query = query.Where("date BETWEEN ? AND ?", condition.StartTime, condition.EndTime)

	query = query.Select("SUM(amount) as Amount,SUM(count) as Count,category_id").Group("category_id")
	err = s.query(ie, query).Order("Amount desc").Limit(limit).Find(&result).Error
	return result, err
}
