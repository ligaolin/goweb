package goweb

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

/**
 * @description：生成随机字符串
 */
func GenerateRandomAlphanumeric(length int) string {
	const (
		letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits  = "0123456789"
	)
	rand.NewSource(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		if rand.Intn(2) == 0 { // 随机选择字母或数字
			result[i] = letters[rand.Intn(len(letters))]
		} else {
			result[i] = digits[rand.Intn(len(digits))]
		}
	}
	return string(result)
}

/**
 * @description：字符串转切片，例如1,2,3转成[]uint{1,2,3}
 */
func ToSliceUint(s any, split string) ([]uint, error) {
	// 分割字符串
	parts := strings.Split(fmt.Sprintf("%v", s), split)
	var result []uint

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue // 忽略空字符串
		}
		// 将子字符串转换为 uint64
		value, err := strconv.ParseUint(trimmedPart, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("error converting string to uint64：%v", err)
		}
		result = append(result, uint(value))
	}

	return result, nil
}

/**
 * @description：字符串转字符串，例如1,2,3转成1','2','3
 */
func StringToString(s string, split0 string, split1 string) string {
	return strings.Join(strings.Split(s, split0), split1)
}

/**
 * @description：判断字符串是否以prefix开头
 */
func StringPreIs(s string, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

/**
 * @description：从数组中随机获取count个元素（不重复）组成新数组
 */
func GetRandomUniqueStrings(arr []string, count int) ([]string, error) {
	if len(arr) < count {
		return nil, fmt.Errorf("数组长度不足")
	}

	// 生成随机索引排列
	randIndices := rand.Perm(len(arr))[:count]

	result := make([]string, count)
	for i, idx := range randIndices {
		result[i] = arr[idx]
	}

	return result, nil
}
