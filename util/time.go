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

func (t *_time) GetLastSecondOfMonth(date time.Time) time.Time {
	year, month, _ := date.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, date.Location())
	lastSecond := nextMonth.Add(-time.Second)
	return lastSecond
}
