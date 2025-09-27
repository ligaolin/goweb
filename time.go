package goweb

import "time"

// 获取今天剩余时间
func GetRemainingSecondsToday() time.Duration {
	now := time.Now()

	// 获取今天的最后一秒（23:59:59.999999999）
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(),
		23, 59, 59, 999999999, now.Location())

	// 计算剩余时间并返回Duration
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

	// 如果今年的生日还没到，年龄减1
	if birthdate.Month() > today.Month() || (birthdate.Month() == today.Month() && birthdate.Day() > today.Day()) {
		age--
	}

	return age
}
