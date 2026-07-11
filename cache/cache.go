package cache

import (
	"time"
)

type Cache interface {
	Get(key string, value any) error                       // 从缓存中获取数据
	Set(key string, value any, expire time.Duration) error // 将数据存入缓存
	Delete(key string) error                               // 从缓存中删除数据
}

type Config struct {
	Use   string
	File  *FileConfig
	Redis *RedisConfig
}

func New(config *Config) Cache {
	switch config.Use {
	case "redis":
		return NewRedis(config.Redis)
	default:
		return NewFile(config.File)
	}
}
