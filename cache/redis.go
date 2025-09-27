package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Context context.Context
	Client  *redis.Client
}

func (r *Redis) Get(key string, value any) error {
	str, err := r.Client.Get(r.Context, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), value)
}

func (r *Redis) Set(key string, value any, expir time.Duration) error {
	str, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Client.Set(r.Context, key, string(str), expir).Err()
}

func (r *Redis) Delete(key string) error {
	return r.Client.Del(r.Context, key).Err()
}
