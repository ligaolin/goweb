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
	id := uuid.New().String()
	err := c.Cache.Set(c.getKey(id, key), v, expir)
	return id, err
}

func (c *Client) Get(id string, key string, value any) error {
	return c.Cache.Get(c.getKey(id, key), value)
}

func (c *Client) GetAndDelete(id string, key string, value any) error {
	fullKey := c.getKey(id, key)
	if err := c.Cache.Get(fullKey, value); err != nil {
		return err
	}
	return c.Cache.Delete(fullKey)
}

func (c *Client) Delete(id string, key string) error {
	return c.Cache.Delete(c.getKey(id, key))
}

func (c *Client) getKey(id string, key string) string {
	return "client_" + key + "_" + id
}
