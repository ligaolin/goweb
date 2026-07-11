package data

import (
	"math"
	"math/rand"
	"strconv"
)

// FormatFloat 格式化浮点数，保留decimal位小数
func FormatFloat(num float64, decimal int) float64 {
	pow := math.Pow10(decimal)
	return math.Round(num*pow) / pow
}

// TruncateToTwoDecimal 保留两位浮点数小数（截断不四舍五入）
func TruncateToTwoDecimal(num float64) float64 {
	return math.Trunc(num*100) / 100
}

// ParseInt 处理字符串，返回其整数部分，非数字则返回0
func ParseInt(s string) int64 {
	floatVal, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(floatVal)
}

// GenerateRandomIndexes 生成不重复的随机索引
func GenerateRandomIndexes(max, count int) []int {
	if count > max {
		count = max
	}
	return rand.Perm(max)[:count]
}

// EarthRadiusMeters 地球平均半径（米）
const EarthRadiusMeters = 6371000.0

// DistanceInMeters 计算两个经纬度之间的距离（米）
func DistanceInMeters(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadiusMeters * c
}
