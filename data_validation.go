package goweb

import "reflect"

func IsNilOrZero(v any) bool {
	if v == nil {
		return true
	}

	return IsZero(v)
}

func IsZero(v any) bool {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Pointer, reflect.Interface:
		return val.IsNil()
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
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	default:
		return false
	}
}
