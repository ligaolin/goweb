package db

import (
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
func (m *Model) Where(data []Where) *Model {
	if m.Error != nil {
		return m
	}

	for _, v := range data {
		if v.Nullable || (!v.Nullable && !IsZero(v.Value)) {
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
					if !v.Nullable && (IsZero(values[0]) || IsZero(values[1])) {
						continue
					}
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

// isNilOrEmpty 判断 Value 是否为 nil 或空值（使用 IsZero 优化）
func IsZero(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)

	// 处理指针类型
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}

	return v.IsZero()
}
