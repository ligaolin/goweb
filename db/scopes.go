package db

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// normalizePageParams 统一处理分页参数
func normalizePageParams(page, pageSize *int32, maxVal int32) {
	if maxVal <= 0 {
		maxVal = 100
	}
	if *pageSize > maxVal {
		*pageSize = maxVal
	} else if *pageSize <= 0 {
		*pageSize = 10
	}
	if *page <= 0 {
		*page = 1
	}
}

// Paginate 分页
func Paginate(page, pageSize *int32, maxPageSize ...int32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}
		maxVal := int32(100)
		if len(maxPageSize) > 0 {
			maxVal = maxPageSize[0]
		}
		normalizePageParams(page, pageSize, maxVal)
		db.Offset(int((*page - 1) * *pageSize)).Limit(int(*pageSize))
		return db
	}
}

func Offset(offset, limit *int32, maxLimit ...int32) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}
		maxVal := int32(100)
		if len(maxLimit) > 0 {
			maxVal = maxLimit[0]
		}
		if maxVal <= 0 {
			maxVal = 100
		}
		if *limit > maxVal {
			*limit = maxVal
		} else if *limit <= 0 {
			*limit = 10
		}
		if *offset < 0 {
			*offset = 0
		}
		db.Offset(int(*offset)).Limit(int(*limit))
		return db
	}
}

// Where 条件查询
func Where(where string, value any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil || IsNilOrZero(value) {
			return db
		}
		db.Where(where, value)
		return db
	}
}

// Like 模糊查询
func Like(where string, value any, options ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil || IsNilOrZero(value) {
			return db
		}
		pattern := fmt.Sprintf("%%%v%%", value)
		if len(options) > 0 {
			pattern = fmt.Sprintf(options[0], value)
		}
		db.Where(where, pattern)
		return db
	}
}

// Unique 检查字段是否唯一
func Unique(DB *gorm.DB, primaryField string, primaryKey any, message string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if db.Error != nil {
			return db
		}
		if !IsNilOrZero(primaryKey) {
			DB = DB.Where(primaryField+" != ?", primaryKey)
		}
		var count int64
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

func OraclePaginate(db *gorm.DB, data any, page, pageSize *int32, maxPageSize ...int32) error {
	if db.Error != nil {
		return db.Error
	}
	maxVal := int32(100)
	if len(maxPageSize) > 0 {
		maxVal = maxPageSize[0]
	}
	normalizePageParams(page, pageSize, maxVal)

	startRow := (*page-1)*(*pageSize) + 1
	endRow := (*page) * *pageSize

	return db.Session(&gorm.Session{NewDB: true}).Debug().
		Raw("SELECT * FROM ( SELECT a.*, ROWNUM rn FROM (?) a WHERE ROWNUM <= ?) WHERE rn >= ?",
			db, endRow, startRow).Scan(data).Error
}
