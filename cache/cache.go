package cache

import (
	"context"
	"time"

	"github.com/gregjones/httpcache/diskcache"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	// 从缓存中获取数据
	Get(key string, value any) error

	// 将数据存入缓存
	Set(key string, value any, expir time.Duration) error

	// 从缓存中删除数据
	Delete(key string) error
}

type Config struct {
	Use   string       `json:"use" toml:"use" yaml:"use"`
	File  *FileConfig  `json:"file" toml:"file" yaml:"file"`
	Redis *RedisConfig `json:"redis" toml:"redis" yaml:"redis"`
}

type FileConfig struct {
	Path string `json:"path" toml:"path" yaml:"path"`
}

type RedisConfig struct {
	Addr     string `json:"addr" toml:"addr" yaml:"addr"`
	Password string `json:"password" toml:"password" yaml:"password"`
}

func New(config *Config) Cache {
	switch config.Use {
	case "redis":
		return &Redis{
			Context: context.Background(),
			Client:  redis.NewClient(&redis.Options{Addr: config.Redis.Addr, Password: config.Redis.Password}),
		}
	default:
		return &File{
			Client: diskcache.New(config.File.Path),
			Path:   config.File.Path,
		}
	}
}
