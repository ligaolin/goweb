package gin_lin

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

/**
 * @description: 生成指定位数随机整数
 * @param {int32} n 位数
 */
func Random(n int32) int32 {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(int32(math.Pow10(int(n))))
}

/**
 * @description: 生成随机数，指定范围
 * @param {int} min 最小值，包含
 * @param {int} max 最大值，不包含
 */
func RandomInt(min int, max int) int {
	// 生成[min, max)范围的随机数
	return rand.Intn(max-min) + min
}

/**
 * @description: 生成随机数，指定范围
 * @param {int} min 最小值，包含
 * @param {int} max 最大值，不包含
 */
func RandomFloat(start float64, end float64) float64 {
	return start + rand.Float64()*(end-start)
}

/**
 * @description: 保留两位浮点数小数
 */
func TruncateToTwoDecimal(num float64) float64 {
	return math.Trunc(num*100) / 100
}

// ParseInt 处理字符串，返回其整数部分，非数字则返回0
func ParseInt(s string) int64 {
	// 尝试解析为浮点数
	floatVal, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// 解析失败，返回0
		return 0
	}

	return int64(floatVal) // 直接截断小数部分，取整数部分
}

// 生成不重复的随机索引
func GenerateRandomIndexes(max, count int) []int {
	if count > max {
		count = max
	}

	// 创建索引池
	indexPool := make([]int, max)
	for i := 0; i < max; i++ {
		indexPool[i] = i
	}

	// 随机打乱索引
	rand.Shuffle(max, func(i, j int) {
		indexPool[i], indexPool[j] = indexPool[j], indexPool[i]
	})

	// 返回前count个索引
	return indexPool[:count]
}
