package db

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// 控制数据库连接超时
func OpenDBWithTimeout(dialector gorm.Dialector, timeout time.Duration) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var db *gorm.DB
	var err error

	// 用 channel 等待结果，超时直接退出
	done := make(chan struct{})
	go func() {
		defer close(done)
		db, err = gorm.Open(dialector, &gorm.Config{})
	}()

	select {
	case <-done:
		return db, err
	case <-ctx.Done():
		return nil, ctx.Err() // 强制 1s 超时
	}
}
