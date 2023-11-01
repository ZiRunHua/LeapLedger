package util

import "time"

func IsSameDay(t1 time.Time, t2 time.Time) bool {
	return t1.Truncate(24 * time.Hour).Equal(t2.Truncate(24 * time.Hour))
}

func GetLastMonthMidnight() time.Time {
	currentTime := time.Now()
	lastMonth := currentTime.AddDate(0, -1, 0)
	return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
}
