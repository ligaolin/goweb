package db

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type Where struct {
	Name     string
	Op       string
	Value    any
	Nullable bool
}

// 将 Where 结构体转换为安全的 SQL where 子句和参数
func (m *Model) Where(data []Where) *Model {
	if m.Error != nil {
		return m
	}

	for _, v := range data {
		if v.Nullable || (!v.Nullable && !isNilOrEmpty(v.Value)) {
			switch strings.ToUpper(v.Op) {
			case "IN":
				m.Db = m.Db.Where(fmt.Sprintf("%s IN ?", v.Name), v.Value)
			case "LIKE":
				m.Db = m.Db.Where(fmt.Sprintf("%s LIKE ?", v.Name), fmt.Sprintf("%%%s%%", v.Value))
			case "NOT LIKE":
				m.Db = m.Db.Where(fmt.Sprintf("%s NOT LIKE ?", v.Name), fmt.Sprintf("%%%s%%", v.Value))
			case "IS NULL":
				m.Db = m.Db.Where(fmt.Sprintf("%s IS NULL", v.Name))
			case "IS NOT NULL":
				m.Db = m.Db.Where(fmt.Sprintf("%s IS NOT NULL", v.Name))
			case "FIND_IN_SET":
				m.Db.Where("FIND_IN_SET(?, ?)", v.Name, v.Value)
			case "!=", ">", ">=", "<", "<=":
				m.Db = m.Db.Where(fmt.Sprintf("%s %s ?", v.Name, v.Op), v.Value)
			case "BETWEEN":
				if values, ok := v.Value.([]any); ok && len(values) == 2 {
					m.Db = m.Db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", v.Name), values[0], values[1])
				} else {
					m.Error = fmt.Errorf("BETWEEN requires a slice of 2 values")
					return m
				}
			default:
				m.Db = m.Db.Where(fmt.Sprintf("%s = ?", v.Name), v.Value)
			}
		}
	}

	return m
}

// isNilOrEmpty 判断 Value 是否为 nil 或空值
func isNilOrEmpty(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	// 如果是指针，解引用
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Struct:
		if t, ok := value.(time.Time); ok {
			return t.IsZero()
		}
		return false // 其他结构体默认非空
	default:
		return false
	}
}
