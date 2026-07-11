package db

import "reflect"

// IsNilOrZero 判断值是否为 nil 或零值
func IsNilOrZero(v any) bool {
	if v == nil {
		return true
	}
	return IsZero(v)
}

// IsZero 判断值是否为零值
func IsZero(v any) bool {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Array:
		return val.IsNil() || val.Len() == 0
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	default:
		return false
	}
}
