package cache

import (
	"time"

	"github.com/google/uuid"
)

type Client struct {
	Cache Cache
}

func NewClient(cache Cache) *Client {
	return &Client{
		Cache: cache,
	}
}

func (c *Client) Set(key string, v any, expir time.Duration) (string, error) {
	uuid := uuid.New().String()
	err := c.Cache.Set("client-"+key+uuid, v, expir)
	return uuid, err
}

func (c *Client) Get(uuid string, key string, value any, clear bool) error {
	if err := c.Cache.Get("client-"+key+uuid, value); err != nil {
		return err
	}
	if clear {
		c.Cache.Delete("client-" + key + uuid)
	}
	return nil
}
