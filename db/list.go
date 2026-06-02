package db

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

type ListData struct {
	Data             any   `json:"data"`
	Total            int32 `json:"total"`      // 总数量
	TotalPage        int32 `json:"total_page"` // 总页数
	Page             int32 `json:"page"`
	PageSize         int32 `json:"page_size"`
	CountUseSubQuery bool  `json:"count_use_sub_query"`
}

// 查询列表
func (m *Model[T]) List(data *ListData) *Model[T] {
	if m.Error != nil {
		return m
	}

	// 查询总数
	var total int64
	if data.CountUseSubQuery {
		if err := m.Db.Session(&gorm.Session{NewDB: true}).Table("(?) as t", m.Db.Model(m.Model)).Count(&total).Error; err != nil {
			m.Error = err
			return m
		}
	} else {
		if err := m.Db.Model(m.Model).Count(&total).Error; err != nil {
			m.Error = err
			return m
		}
	}
	data.Total = int32(total)

	// 添加分页
	if data.Page > 0 {
		if data.PageSize <= 0 {
			data.PageSize = 10
		}
		m.Db = m.Db.Offset(int((data.Page - 1) * data.PageSize))
	}
	if data.PageSize > 1000 {
		data.PageSize = 1000
	}
	if data.PageSize > 0 {
		m.Db = m.Db.Limit(int(data.PageSize))
	}

	// 执行查询
	if err := m.Db.Find(&data.Data).Error; err != nil {
		m.Error = err
		return m
	}

	// 计算总页数
	if data.Page > 0 {
		data.TotalPage = (data.Total + data.PageSize - 1) / data.PageSize
	} else {
		data.TotalPage = 1
	}

	return m
}

type ListResult struct {
	Data     any   `json:"data"`
	Total    int64 `json:"total"` // 总数量
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

func Paginate(page, pageSize *int32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if *pageSize > 1000 {
			*pageSize = 1000
		} else if *pageSize <= 0 {
			*pageSize = 10
		}
		if *page > 0 {
			db.Offset(int((*page - 1) * *pageSize))
		}
		if *pageSize > 0 {
			db.Limit(int(*pageSize))
		}
		return db
	}
}

func W(where string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isNilOrZero(value) {
			return db
		}

		db.Where(where, value)
		return db
	}
}

func Like(where string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isNilOrZero(value) {
			return db
		}

		db.Where(where, fmt.Sprintf("%%%v%%", value))
		return db
	}
}

func isNilOrZero(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
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
