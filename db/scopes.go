package db

import (
	"errors"
	"fmt"

	"github.com/ligaolin/goweb"
	"gorm.io/gorm"
)

type ListResult struct {
	Data     any   `json:"data"`
	Total    int64 `json:"total"` // 总数量
	Page     int32 `json:"page"`
	PageSize int32 `json:"page_size"`
}

// Paginate 分页
func Paginate(page, pageSize *int32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}

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

// Where 条件查询
func W(where string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}

		if goweb.IsNilOrZero(value) {
			return db
		}

		db.Where(where, value)
		return db
	}
}

// Like 模糊查询
func Like(where string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}

		if goweb.IsNilOrZero(value) {
			return db
		}

		db.Where(where, fmt.Sprintf("%%%v%%", value))
		return db
	}
}

// Unique 检查字段是否唯一
func Unique(DB *gorm.DB, primaryField string, primaryKey any, message string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}

		var count int64
		if !goweb.IsNilOrZero(primaryKey) {
			DB = DB.Where(primaryField+" != ?", primaryKey)
		}
		if err := DB.Count(&count).Error; err != nil {
			db.Error = err
			return db
		}
		if count > 0 {
			db.Error = errors.New(message)
		}
		return db
	}
}
