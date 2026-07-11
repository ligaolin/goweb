package time

import "time"

// GetRemainingSecondsToday 获取今天剩余时间
func GetRemainingSecondsToday() time.Duration {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(),
		23, 59, 59, 999999999, now.Location())
	return endOfDay.Sub(now)
}

// GetAgeFromStringDate 从字符串日期直接计算年龄
func GetAgeFromStringDate(dateStr string) (int, error) {
	birthdate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, err
	}
	return CalculateAge(birthdate), nil
}

// CalculateAge 计算年龄
func CalculateAge(birthdate time.Time) int {
	today := time.Now()
	age := today.Year() - birthdate.Year()
	if birthdate.Month() > today.Month() ||
		(birthdate.Month() == today.Month() && birthdate.Day() > today.Day()) {
		age--
	}
	return age
}
