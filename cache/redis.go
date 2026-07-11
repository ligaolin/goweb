package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Redis struct {
	Client *redis.Client
}

func NewRedis(config *RedisConfig) *Redis {
	return &Redis{
		Client: redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
		}),
	}
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *Redis) Get(key string, value any) error {
	return r.GetWithContext(context.Background(), key, value)
}

func (r *Redis) GetWithContext(ctx context.Context, key string, value any) error {
	str, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), value)
}

func (r *Redis) Set(key string, value any, expir time.Duration) error {
	return r.SetWithContext(context.Background(), key, value, expir)
}

func (r *Redis) SetWithContext(ctx context.Context, key string, value any, expir time.Duration) error {
	str, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(ctx, key, string(str), expir).Err()
}

func (r *Redis) Delete(key string) error {
	return r.DeleteWithContext(context.Background(), key)
}

func (r *Redis) DeleteWithContext(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
