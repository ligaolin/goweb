package config

type Config[T any] struct {
	data *T
}

func (c *Config[T]) Get() *T {
	return c.data
}

func (c *Config[T]) Clone() *T {
	if c.data == nil {
		return nil
	}
	clone := *c.data
	return &clone
}
