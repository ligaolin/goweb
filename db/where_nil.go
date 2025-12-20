package db

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type WhereNil struct {
	Name  string // 字段名
	Op    string // 操作符
	Value any
}

func (m *Model) WhereNil(data *[]WhereNil) *Model {
	if m.Error != nil {
		return m
	}

	if data == nil {
		return m
	}

	for _, v := range *data {
		if v.Value != nil {
			upperOp := strings.ToUpper(v.Op)
			actualValue := getActualValue(v.Value)
			switch upperOp {
			case "IN":
				if IsZero(actualValue) {
					continue
				}
				m.Db = m.Db.Where(fmt.Sprintf("%s IN ?", v.Name), actualValue)
			case "LIKE", "NOT LIKE":
				if IsZero(actualValue) {
					continue
				}
				likeValue := fmt.Sprintf("%%%s%%", actualValue)
				m.Db = m.Db.Where(fmt.Sprintf("%s %s ?", v.Name, upperOp), likeValue)
			case "IS NULL":
				m.Db = m.Db.Where(fmt.Sprintf("%s IS NULL", v.Name))
			case "IS NOT NULL":
				m.Db = m.Db.Where(fmt.Sprintf("%s IS NOT NULL", v.Name))
			case "FIND_IN_SET":
				if IsZero(actualValue) {
					continue
				}
				m.Db = m.Db.Where("FIND_IN_SET(?, ?)", actualValue, v.Name)
			case "!=", ">", ">=", "<", "<=":
				m.Db = m.Db.Where(fmt.Sprintf("%s %s ?", v.Name, upperOp), actualValue)
			case "BETWEEN":
				values, ok := getActualValue(actualValue).([]any)
				if !ok || len(values) != 2 {
					m.Error = errors.New("BETWEEN requires a slice of 2 values")
					return m
				}
				if IsZero(values[0]) || IsZero(values[1]) {
					continue
				}
				m.Db = m.Db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", v.Name), values[0], values[1])
			default:
				if reflect.ValueOf(v.Value).Kind() != reflect.Pointer {
					m.Db = m.Db.Where(fmt.Sprintf("%s = ?", v.Name), actualValue)
				}
			}
		}
	}

	return m
}

// 安全解引用指针，获取实际值
func getActualValue(value any) any {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	// 循环解引用所有层级的指针
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil // 指针nil，返回nil
		}
		v = v.Elem()
	}
	return v.Interface()
}
