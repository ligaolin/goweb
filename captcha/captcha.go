package captcha

import (
	"fmt"
	"time"

	"github.com/ligaolin/goweb/cache"
)

type Captcha[T any] struct {
	Client *cache.Client
}

func NewCaptcha[T any](c *cache.Client) *Captcha[T] {
	return &Captcha[T]{Client: c}
}

func (c *Captcha[T]) Generate(key string, expir time.Duration) (string, error) {
	code := generateCode()
	v := value{
		Code: code,
	}
	uuid, err := c.Client.Set(key, v, expir)
	if err != nil {
		return "", fmt.Errorf("存储验证码失败: %w", err)
	}
	return uuid, nil
}

func (c *Captcha[T]) Verify(key string, uuid string, code string) error {
	var val value
	if err := c.Client.GetAndDelete(uuid, key, &val); err != nil {
		return fmt.Errorf("验证码不存在或已过期: %w", err)
	}
	if val.Code != code {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

func (c *Captcha[T]) Delete(key string, uuid string) error {
	return c.Client.Delete(uuid, key)
}
