package util

import "time"

type _time struct{}

var Time = &_time{}

func (t *_time) IsSameDay(t1 time.Time, t2 time.Time) bool {
	return t1.Truncate(24 * time.Hour).Equal(t2.Truncate(24 * time.Hour))
}

func (t *_time) GetLastMonthMidnight() time.Time {
	currentTime := time.Now()
	lastMonth := currentTime.AddDate(0, -1, 0)
	return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
}

func (t *_time) SplitMonths(startDate, endDate time.Time) []time.Time {
	var months []time.Time

	current := startDate
	for !current.After(endDate) {
		months = append(months, current)

		current = current.AddDate(0, 1, 0)
		current = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())

		if current.After(endDate) || current.Equal(endDate) {
			break
		}
	}

	return months
}

func (t *_time) SplitDays(startDate, endDate time.Time) []time.Time {
	duration := endDate.Sub(startDate)
	length := int(duration.Hours()/24) + 1
	days := make([]time.Time, length, length)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	current := startDate
	for i := 0; i < len(days); i++ {
		days[i] = current
		current = time.Date(current.Year(), current.Month(), current.Day()+1, 0, 0, 0, 0, current.Location())
	}
	return days
}

func (t *_time) GetFirstSecondOfMonth(date time.Time) time.Time {
	year, month, _ := date.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, date.Location())
}

func (t *_time) GetLastSecondOfMonth(date time.Time) time.Time {
	year, month, _ := date.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, date.Location())
	lastSecond := nextMonth.Add(-time.Second)
	return lastSecond
}

// 获取本周一的第一秒
func (t *_time) GetFirstSecondOfMonday(currentTime time.Time) time.Time {
	weekday := currentTime.Weekday()
	// 计算需要减去的天数，使得得到本周一的日期
	daysToMonday := time.Duration(0)
	if weekday != time.Monday {
		daysToMonday = time.Duration(weekday - time.Monday)
		if weekday < time.Monday {
			daysToMonday += 7 // 如果当前是周日，则需要向前推算7天
		}
	}

	monday := currentTime.Add(-daysToMonday * 24 * time.Hour)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, currentTime.Location())
	return monday
}

// 获取今年的第一秒
func (t *_time) GetFirstSecondOfYear(currentTime time.Time) time.Time {
	year := currentTime.Year()
	firstSecondOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, currentTime.Location())
	return firstSecondOfYear
}
