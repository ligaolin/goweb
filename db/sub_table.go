package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ShardingMeta 分表元数据模型：记录基础表对应的所有分表
type ShardingMeta struct {
	ID          int32     `gorm:"primaryKey;autoIncrement"`         // 自增主键
	BaseTable   string    `gorm:"type:varchar(50);not null;index"`  // 基础表名（如 access_log）
	ActualTable string    `gorm:"type:varchar(50);not null;unique"` // 实际分表名（如 access_log_1）
	Num         int64     `gorm:"not null"`                         // 分表序号（1、2、3...）
	CreatedAt   time.Time `gorm:"type:datetime;"`                   // 分表创建时间
}

func (m *ShardingMeta) TableName() string {
	return "sharding_meta"
}

type ShardingService struct {
	baseName string   // 基础表名
	gormDB   *gorm.DB // GORM 数据库实例
	MaxNum   int64
	Table    string
	Num      int64
	Error    error
}

func NewShardingService(gormDB *gorm.DB, baseName string, maxNum int64) *ShardingService {
	gormDB.AutoMigrate(&ShardingMeta{})
	s := ShardingService{
		baseName: baseName,
		gormDB:   gormDB,
		MaxNum:   maxNum,
	}
	if baseName == "" {
		s.Error = errors.New("baseName 不能为空")
	}

	return &s
}

func (s *ShardingService) SetTable(model any) *ShardingService {
	s.GetTable(model)
	if s.Error != nil {
		return s
	}

	var recordCount int64
	if s.Error = s.gormDB.Table(s.Table).Model(model).Count(&recordCount).Error; s.Error != nil {
		return s
	}
	// 如果当前表记录数 >= 最大限制，创建新表
	if recordCount >= s.MaxNum {
		s.Error = s.createTable(model).Error
	}

	return s
}

func (s *ShardingService) GetTable(model any) *ShardingService {
	if s.Error != nil {
		return s
	}

	// 查询最新分表，没有记录，创建新表
	var latestMeta ShardingMeta
	if err := s.gormDB.Where("base_table = ?", s.baseName).Order("num desc").First(&latestMeta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.Error = s.createTable(model).Error
		} else {
			s.Error = fmt.Errorf("查询最新分表失败: %w", err)
		}
		if s.Error != nil {
			return s
		}
	} else {
		s.Table = latestMeta.ActualTable
		// 分表迁移
		if err := s.gormDB.Table(s.Table).AutoMigrate(model); err != nil {
			s.Error = fmt.Errorf("分表迁移失败: %w", err)
			return s
		}
	}

	return s
}

func (s *ShardingService) createTable(model any) *ShardingService {
	if s.Error != nil {
		return s
	}
	s.Error = s.gormDB.Transaction(func(tx *gorm.DB) error {
		s.Num++
		s.Table = fmt.Sprintf("%s_%d", s.baseName, s.Num)

		// 创建分表
		if err := tx.Table(s.Table).AutoMigrate(model); err != nil {
			return fmt.Errorf("创建分表 %s 失败: %w", s.Table, err)
		}

		// 写入分表元数据
		return tx.Create(&ShardingMeta{
			BaseTable:   s.baseName,
			ActualTable: s.Table,
			Num:         s.Num,
		}).Error
	})

	return s
}
