package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func NewTOML[T any](path string) (*Config[T], error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg T
	if err := toml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("解析toml类型配置文件失败: %w", err)
	}
	return &Config[T]{data: &cfg}, nil
}
