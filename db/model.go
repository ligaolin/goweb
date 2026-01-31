package db

import (
	"errors"
	"fmt"
	"slices"

	"github.com/jinzhu/copier"
	"github.com/ligaolin/goweb"
	"gorm.io/gorm"
)

type Model[T any] struct {
	Db     *gorm.DB
	Error  error
	Pk     int32
	PkName string
	Model  T
}

func NewModel[T any](db *gorm.DB, model T) *Model[T] {
	return &Model[T]{
		Db:     db,
		PkName: "id",
		Model:  model,
	}
}

func (m *Model[T]) SetPkName(pkName string) *Model[T] {
	if m.Error != nil {
		return m
	}
	m.PkName = pkName
	return m
}

func (m *Model[T]) SetPk(pk int32) *Model[T] {
	if m.Error != nil {
		return m
	}
	m.Pk = pk
	if m.Pk != 0 {
		if err := m.Db.First(m.Model, m.Pk).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				m.Error = err
				return m
			}
		}
	}
	return m
}

func (m *Model[T]) SetModel(db *gorm.DB) *Model[T] {
	if m.Error != nil {
		return m
	}
	if m.Pk != 0 {
		if err := db.First(m.Model).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				m.Error = err
				return m
			}
		}
	}
	return m
}

func (m *Model[T]) Copy(param any) *Model[T] {
	if m.Error != nil {
		return m
	}

	if err := copier.Copy(m.Model, param); err != nil {
		m.Error = err
	}
	return m
}

type Same struct {
	Db      *gorm.DB
	Message string
}

// 唯一性判断
func (m *Model[T]) NotSame(sames *[]Same) *Model[T] {
	if m.Error != nil {
		return m
	}

	var count int64
	for _, v := range *sames {
		if err := v.Db.Model(m.Model).Where(fmt.Sprintf("%s != ?", m.PkName), m.Pk).Count(&count).Error; err != nil {
			m.Error = err
			return m
		}
		if count > 0 {
			m.Error = errors.New(v.Message)
			return m
		}
	}
	return m
}

func (m *Model[T]) Save() *Model[T] {
	if m.Error != nil {
		return m
	}

	if err := m.Db.Save(m.Model).Error; err != nil {
		m.Error = err
		return m
	}
	return m
}

// 更新
func (m *Model[T]) Update(field string, value any, containsas []string) *Model[T] {
	if m.Error != nil {
		return m
	}

	if field == "" {
		m.Error = errors.New("field必须")
		return m
	}
	if !slices.Contains(containsas, field) {
		m.Error = errors.New("field数据不合法")
		return m
	}
	if err := m.Db.Model(m.Model).Where(m.PkName+" = ?", m.Pk).Update(field, value).Error; err != nil {
		m.Error = err
		return m
	}
	return m
}

// 删除
func (m *Model[T]) Delete(id any) *Model[T] {
	if m.Error != nil {
		return m
	}

	if err := m.Db.Delete(m.Model, id).Error; err != nil {
		m.Error = err
		return m
	}
	return m
}

// 生成唯一随机码
func (m *Model[T]) Code(n int, field string) (string, error) {
	for {
		code := goweb.GenerateRandomAlphanumeric(n)
		var count int64
		m.Db.Model(m.Model).Where(field+" = ?", code).Count(&count)
		if count == 0 {
			return code, nil
		}
	}
}
