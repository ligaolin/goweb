package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func NewJSON[T any](path string) (*T, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg T
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("解析json类型配置文件失败: %w", err)
	}
	return &cfg, nil
}

func LoadJSON[T any](path string) (*T, error) {
	return load(path, NewJSON[T])
}