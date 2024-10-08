package timeTool

import "time"

func SplitMonths(startDate, endDate time.Time) [][2]time.Time {
	var months [][2]time.Time
	current := startDate
	for !current.Equal(endDate) {
		current = GetLastSecondOfMonth(startDate)
		if current.After(endDate) {
			current = endDate
		}
		months = append(months, [2]time.Time{startDate, current})
		startDate = current.Add(time.Second)
	}
	return months
}

func SplitDays(startDate, endDate time.Time) []time.Time {
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	var days []time.Time
	current := startDate
	for !current.After(endDate) {
		days = append(days, current)
		current = current.AddDate(0, 0, 1)
	}
	return days
}
func GetFirstSecondOfDay(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func GetFirstSecondOfMonth(date time.Time) time.Time {
	year, month, _ := date.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, date.Location())
}

func GetLastSecondOfMonth(date time.Time) time.Time {
	year, month, _ := date.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, date.Location())
	lastSecond := nextMonth.Add(-time.Second)
	return lastSecond
}

// 获取本周一的第一秒
func GetFirstSecondOfWeek(currentTime time.Time) time.Time {
	weekday := currentTime.Weekday()
	daysToMonday := time.Duration(0)
	if weekday != time.Monday {
		daysToMonday = time.Duration(weekday - time.Monday)
		if weekday < time.Monday {
			daysToMonday += 7
		}
	}

	monday := currentTime.Add(-daysToMonday * 24 * time.Hour)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, currentTime.Location())
	return monday
}

func GetLastSecondOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	daysToSunday := (time.Sunday - weekday) % 7
	if daysToSunday < 0 {
		daysToSunday += 7
	}
	sunday := t.AddDate(0, 0, int(daysToSunday))
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())
	return sunday
}

// 获取今年的第一秒
func GetFirstSecondOfYear(currentTime time.Time) time.Time {
	return time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, currentTime.Location())
}

func GetLastSecondOfYear(currentTime time.Time) time.Time {
	return time.Date(currentTime.Year(), time.December, 31, 23, 59, 59, 0, currentTime.Location())
}

func ToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
