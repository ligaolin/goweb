package db

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Where struct {
	Name     string // 字段名
	Op       string // 操作符
	Value    any    // 条件值
	Nullable bool   // 是否允许为空值
}

// 将 Where 结构体转换为安全的 SQL where 子句和参数
func (m *Model[T]) Where(data []Where) *Model[T] {
	if m.Error != nil {
		return m
	}

	for _, v := range data {
		if (v.Nullable && !IsAnyNil(v.Value)) || (!v.Nullable && !IsZero(v.Value)) {
			upperOp := strings.ToUpper(v.Op)
			// 对值进行安全解引用
			realValue := getRealValue(v.Value)
			switch upperOp {
			case "IN":
				m.Db = m.Db.Where(fmt.Sprintf("%s IN ?", v.Name), realValue)
			case "LIKE", "NOT LIKE":
				m.Db = m.Db.Where(fmt.Sprintf("%s %s ?", v.Name, upperOp), fmt.Sprintf("%%%s%%", realValue))
			case "IS NULL":
				m.Db = m.Db.Where(fmt.Sprintf("%s IS NULL", v.Name))
			case "IS NOT NULL":
				m.Db = m.Db.Where(fmt.Sprintf("%s IS NOT NULL", v.Name))
			case "FIND_IN_SET":
				m.Db = m.Db.Where("FIND_IN_SET(?, ?)", realValue, v.Name) // 注意参数顺序：FIND_IN_SET(值, 字段)
			case "!=", ">", ">=", "<", "<=":
				m.Db = m.Db.Where(fmt.Sprintf("%s %s ?", v.Name, upperOp), realValue)
			case "BETWEEN":
				values, ok := realValue.([]any)
				if !ok || len(values) != 2 {
					m.Error = errors.New("BETWEEN requires a slice of 2 values")
					return m
				}
				if !v.Nullable && (IsZero(values[0]) || IsZero(values[1])) {
					continue
				}
				m.Db = m.Db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", v.Name), values[0], values[1])
			default:
				m.Db = m.Db.Where(fmt.Sprintf("%s = ?", v.Name), realValue)
			}
		}
	}

	return m
}

// 新增：安全解引用函数，获取指针指向的实际值
func getRealValue(value any) any {
	if value == nil {
		return ""
	}

	v := reflect.ValueOf(value)
	// 递归解引用所有指针/接口，直到拿到基础类型
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "" // 空指针返回空字符串
		}
		v = v.Elem()
	}
	return v.Interface()
}

func IsAnyNil(a any) bool {
	if a == nil {
		return true
	}
	v := reflect.ValueOf(a)
	return (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) && v.IsNil()
}

func IsZero(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)

	for {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
			if v.IsNil() {
				return true
			}
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
				continue
			}
		}
		break
	}
	return v.IsZero()
}
