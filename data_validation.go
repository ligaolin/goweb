package goweb

import "reflect"

func IsNilOrZero(v any) bool {
	if v == nil {
		return true
	}

	return IsZero(v)
}

func IsZero(v any) bool {
	switch val := v.(type) {
	case int, int8, int16, int32, int64:
		return val == 0
	case uint, uint8, uint16, uint32, uint64:
		return val == 0
	case string:
		return val == ""
	case nil:
		return true
	default:
		// 其他类型用反射兜底
		return reflect.ValueOf(v).IsZero() // Go 1.13+ 支持
	}
}
