package db

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

func (m *Model) FindChildrenID(ids *[]int32, pidName string) *Model {
	if m.Error != nil {
		return m
	}

	var cids []int32
	if err := m.Db.Model(m.Model).Where(pidName+" IN ?", *ids).Pluck(m.PkName, &cids).Error; err != nil {
		m.Error = err
		return m
	}
	if len(cids) > 0 {
		if err := m.FindChildrenID(&cids, pidName).Error; err != nil {
			m.Error = err
			return m
		}
		*ids = append(*ids, cids...)
	}
	return m
}
func (m *Model) FindChildren(pid any, idName, pidName, childrenName, order string) *Model {
	if m.Error != nil {
		return m
	}

	// 设置默认字段名
	if pidName == "" {
		pidName = "pid" // 更常见的默认列名
	}
	if idName == "" {
		idName = "ID"
	}
	if childrenName == "" {
		childrenName = "Children"
	}

	// 检查模型是否为切片指针
	sliceVal := reflect.ValueOf(m.Model)
	if sliceVal.Kind() != reflect.Ptr || sliceVal.Elem().Kind() != reflect.Slice {
		m.Error = fmt.Errorf("target must be a pointer to slice")
		return m
	}

	// 获取切片元素类型
	elemType := sliceVal.Elem().Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// 关键修复：每次递归创建新的 DB 会话
	db := m.Db.Session(&gorm.Session{NewDB: true})

	// 查询当前层级数据
	if err := db.Where(fmt.Sprintf("%s = ?", pidName), pid).Order(order).Find(m.Model).Error; err != nil {
		m.Error = err
		return m
	}

	// 处理子节点
	slice := sliceVal.Elem()
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		// 获取当前节点ID
		idField := item.FieldByName(idName)
		if !idField.IsValid() {
			m.Error = fmt.Errorf("field %s not found", idName)
			return m
		}

		// 准备子节点容器
		childrenField := item.FieldByName(childrenName)
		if !childrenField.IsValid() {
			m.Error = fmt.Errorf("field %s not found", childrenName)
			return m
		}

		childrenSlice := reflect.New(reflect.SliceOf(elemType)).Interface()
		childModel := &Model{
			Db:    db, // 使用新的会话
			Model: childrenSlice,
		}

		// 递归查询
		if err := childModel.FindChildren(idField.Interface(), idName, pidName, childrenName, order).Error; err != nil {
			m.Error = err
			return m
		}

		// 赋值子节点
		childrenField.Set(reflect.ValueOf(childrenSlice).Elem())
	}

	return m
}
