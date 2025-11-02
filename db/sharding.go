package db

import (
	"fmt"
	"regexp"
	"sort"

	"gorm.io/gorm"
)

type Sharding struct {
	gormDB   *gorm.DB
	baseName string
	num      int32   // 最大分表编号
	maxNum   int32   // 表记录最大数量
	tables   []Table // 分表集合
}

type Table struct {
	Name string
	Num  int32
}

func NewSharding(gormDB *gorm.DB, baseName string) *Sharding {
	return &Sharding{
		baseName: baseName,
		gormDB:   gormDB,
		num:      0,
		maxNum:   1000000, // 默认100万
	}
}

func (s *Sharding) SetMaxNum(maxNum int32) *Sharding {
	s.maxNum = maxNum
	return s
}

func (s *Sharding) getTables() error {
	allTables, err := s.gormDB.Migrator().GetTables()
	if err != nil {
		return err
	}

	pattern := regexp.MustCompile(fmt.Sprintf(`^%s_(\d+)$`, s.baseName))

	for _, table := range allTables {
		if pattern.MatchString(table) {
			matches := pattern.FindStringSubmatch(table)
			if len(matches) > 1 {
				var currentNum int32
				fmt.Sscanf(matches[1], "%d", &currentNum)
				if currentNum > s.num {
					s.num = currentNum
				}
				s.tables = append(s.tables, Table{
					Name: table,
					Num:  currentNum,
				})
			}
		}
	}
	return nil
}

func (s *Sharding) Tables() ([]Table, error) {
	if err := s.getTables(); err != nil {
		return nil, err
	}

	// 排序
	sort.Slice(s.tables, func(i, j int) bool {
		return s.tables[i].Num < s.tables[j].Num
	})
	return s.tables, nil
}

func (s *Sharding) GetTable(model any) (oldTable, newTable string, err error) {
	if err := s.getTables(); err != nil {
		return "", "", err
	}

	if s.num == 0 {
		s.num = 1
		if err := s.migrate(model); err != nil {
			return "", "", err
		}
		oldTable = s.getTableName()
		newTable = oldTable
	} else {
		oldTable = s.getTableName()
		maxID, err := s.GetMaxPrimaryKeyValue()
		if err != nil {
			return "", "", err
		}
		if maxID >= s.maxNum {
			s.num++
			if err := s.migrate(model); err != nil {
				return "", "", err
			}
			newTable = s.getTableName()
		} else {
			newTable = oldTable
		}
	}

	return oldTable, newTable, nil
}

// 查询当前分表的最大ID
func (s *Sharding) GetMaxPrimaryKeyValue() (maxID int32, err error) {
	primaryKey, err := s.getPrimaryKeyColumn()
	if err != nil {
		return 0, err
	}

	err = s.gormDB.Table(s.getTableName()).Select(fmt.Sprintf("COALESCE(MAX(%s), 0)", primaryKey)).Scan(&maxID).Error
	return maxID, err
}

// 自动迁移分表
func (s *Sharding) migrate(model any) (err error) {
	return s.gormDB.Table(s.getTableName()).AutoMigrate(model)
}

// 获取分表的主键列名
func (s *Sharding) getPrimaryKeyColumn() (primaryKey string, err error) {
	err = s.gormDB.Raw(`
        SELECT COLUMN_NAME 
        FROM INFORMATION_SCHEMA.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = ? 
        AND COLUMN_KEY = 'PRI' 
        LIMIT 1`,
		s.getTableName(),
	).Scan(&primaryKey).Error
	return
}

// 获取分表的表名
func (s *Sharding) getTableName() string {
	return fmt.Sprintf("%s_%d", s.baseName, s.num)
}
