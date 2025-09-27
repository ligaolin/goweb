package db

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type Backup struct {
	Db   *sql.DB
	path string
}

// 创建一个数据库备份对象
func NewDbBackup(db *gorm.DB, path string) (*Backup, error) {
	// 连接数据库
	sqlxDb, err := db.DB()
	if err != nil {
		return nil, err
	}
	return &Backup{Db: sqlxDb, path: path}, err
}

func (db *Backup) GetDbName() (string, error) {
	var dbName string
	err := db.Db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		return "", err
	}
	return dbName, nil
}

// 备份数据库
func (db *Backup) Backup() error {
	dir := filepath.Dir(db.path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("无法创建目录 %s：%w", dir, err)
	}

	file, err := os.Create(db.path)
	if err != nil {
		return fmt.Errorf("无法创建备份文件：%w", err)
	}
	defer file.Close()

	dbName, err := db.GetDbName()
	if err != nil {
		return fmt.Errorf("获取数据库名称失败：%w", err)
	}
	file.WriteString("USE `" + dbName + "`;\n\n")

	tables, err := getTables(db.Db)
	if err != nil {
		return fmt.Errorf("获取表列表失败：%w", err)
	}

	for _, table := range tables {
		file.WriteString("DROP TABLE IF EXISTS `" + table + "`;\n")
		createTableSQL, err := getCreateTableSQL(db.Db, table)
		if err != nil {
			return fmt.Errorf("获取表 %s 结构失败：%w", table, err)
		}
		file.WriteString(createTableSQL + ";\n")

		// 使用事务批量插入数据
		tx, err := db.Db.Begin()
		if err != nil {
			return fmt.Errorf("启动事务失败：%w", err)
		}

		rows, err := tx.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("查询表 %s 数据失败：%w", table, err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("获取表 %s 列信息失败：%w", table, err)
		}

		var insertSQL strings.Builder
		for rows.Next() {
			values := make([]any, len(columns))
			pointers := make([]any, len(columns))
			for i := range values {
				pointers[i] = &values[i]
			}

			if err := rows.Scan(pointers...); err != nil {
				tx.Rollback()
				return fmt.Errorf("读取表 %s 数据失败：%w", table, err)
			}

			if insertSQL.Len() == 0 {
				insertSQL.WriteString("INSERT INTO `" + table + "` (`" + strings.Join(columns, "`, `") + "`) VALUES \n\t(")
			} else {
				insertSQL.WriteString(",\n\t(")
			}

			for i, value := range values {
				if i > 0 {
					insertSQL.WriteString(", ")
				}
				switch v := value.(type) {
				case []byte:
					insertSQL.WriteString(fmt.Sprintf("'%s'", string(v)))
				case string:
					insertSQL.WriteString(fmt.Sprintf("'%s'", v))
				case time.Time:
					insertSQL.WriteString(fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05")))
				case nil:
					insertSQL.WriteString("NULL")
				default:
					insertSQL.WriteString(fmt.Sprintf("%v", v))
				}
			}
			insertSQL.WriteString(")")
		}

		if insertSQL.Len() > 0 {
			insertSQL.WriteString(";\n\n")
			file.WriteString(insertSQL.String())
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败：%w", err)
		}
	}

	return nil
}

// 获取数据库中的所有表
func getTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// 获取表的 CREATE TABLE 语句
func getCreateTableSQL(db *sql.DB, table string) (string, error) {
	var tableSQL string
	err := db.QueryRow(fmt.Sprintf("SHOW CREATE TABLE %s", table)).Scan(&table, &tableSQL)
	if err != nil {
		return "", err
	}
	return tableSQL, nil
}

// 恢复数据库
func (db *Backup) Restore() error {
	file, err := os.Open(db.path)
	if err != nil {
		return fmt.Errorf("无法打开备份文件：%w", err)
	}
	defer file.Close()

	tx, err := db.Db.Begin()
	if err != nil {
		return fmt.Errorf("启动事务失败：%w", err)
	}

	scanner := bufio.NewScanner(file)
	var sqlBuffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}

		sqlBuffer.WriteString(line)
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			_, err := tx.Exec(sqlBuffer.String())
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("执行 SQL 语句失败: %w\nSQL：%s", err, sqlBuffer.String())
			}
			sqlBuffer.Reset()
		} else {
			sqlBuffer.WriteString(" ")
		}
	}

	if err := scanner.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("读取备份文件失败：%w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败：%w", err)
	}

	return nil
}
