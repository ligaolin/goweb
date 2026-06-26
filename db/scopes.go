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
func Paginate(page, pageSize *int32, maxPageSize ...int32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}
		var max int32 = 100
		if len(maxPageSize) > 0 {
			max = maxPageSize[0]
		}
		if max <= 0 {
			max = 100
		}
		if *pageSize > max {
			*pageSize = max
		} else if *pageSize <= 0 {
			*pageSize = 10
		}
		if *page > 0 {
			db.Offset(int((*page - 1) * *pageSize))
		} else if *page <= 0 {
			*page = 1
		}
		db.Limit(int(*pageSize))
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
func Like(where string, value any, options ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}

		if goweb.IsNilOrZero(value) {
			return db
		}

		str := fmt.Sprintf("%%%v%%", value)
		if len(options) > 0 {
			str = fmt.Sprintf(options[0], value)
		}

		db.Where(where, str)
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

func OcaclePaginate(db *gorm.DB, data any, page, pageSize *int32, maxPageSize ...int32) error {
	if db.Error != nil {
		return db.Error
	}
	var max int32 = 100
	if len(maxPageSize) > 0 {
		max = maxPageSize[0]
	}
	if max <= 0 {
		max = 100
	}
	if *pageSize > max {
		*pageSize = max
	} else if *pageSize <= 0 {
		*pageSize = 10
	}
	if *page <= 0 {
		*page = 1
	}

	startRow := (*page-1)*(*pageSize) + 1
	endRow := (*page) * *pageSize

	return db.Session(&gorm.Session{NewDB: true}).Debug().Raw("SELECT * FROM ( SELECT a.*, ROWNUM rn FROM (?) a WHERE ROWNUM <= ?) WHERE rn >= ?", db, endRow, startRow).Scan(data).Error
}
