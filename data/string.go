package data

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// GenerateRandomAlphanumeric 生成随机字母数字字符串
func GenerateRandomAlphanumeric(length int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// ToSlice 字符串转切片，例如1,2,3转成[]T{1,2,3}
func ToSlice[T any](s any, split string) ([]T, error) {
	parts := strings.Split(fmt.Sprintf("%v", s), split)
	var result []T

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}

		// 根据类型进行不同的转换
		var value T
		switch any(value).(type) {
		case string:
			value = any(trimmedPart).(T)
		case int, int8, int16, int32, int64:
			v, err := strconv.ParseInt(trimmedPart, 10, 64)
			if err != nil {
				return nil, err
			}
			value = any(v).(T)
		case uint, uint8, uint16, uint32, uint64:
			v, err := strconv.ParseUint(trimmedPart, 10, 64)
			if err != nil {
				return nil, err
			}
			value = any(v).(T)
		case float32, float64:
			v, err := strconv.ParseFloat(trimmedPart, 64)
			if err != nil {
				return nil, err
			}
			value = any(v).(T)
		case bool:
			v, err := strconv.ParseBool(trimmedPart)
			if err != nil {
				return nil, err
			}
			value = any(v).(T)
		default:
			return nil, fmt.Errorf("unsupported type: %T", value)
		}
		result = append(result, value)
	}
	return result, nil
}

// GetRandomUnique 从切片中随机获取count个元素（不重复）
func GetRandomUnique[T any](arr []T, count int) ([]T, error) {
	if len(arr) < count {
		return nil, fmt.Errorf("数组长度不足")
	}
	randIndices := rand.Perm(len(arr))[:count]
	result := make([]T, count)
	for i, idx := range randIndices {
		result[i] = arr[idx]
	}
	return result, nil
}
