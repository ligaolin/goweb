package config

import "sync"

var instances sync.Map

func load[T any](path string, fn func(string) (*T, error)) (*T, error) {
	if v, ok := instances.Load(path); ok {
		return v.(*T), nil
	}
	cfg, err := fn(path)
	if err != nil {
		return nil, err
	}
	instances.Store(path, cfg)
	return cfg, nil
}