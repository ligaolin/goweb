package config

import (
	"fmt"
	"os"

	"sigs.k8s.io/yaml"
)

func NewYAML[T any](path string) (*T, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	var cfg T
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("解析yaml类型配置文件失败: %w", err)
	}
	return &cfg, nil
}

func LoadYAML[T any](path string) (*T, error) {
	return load(path, NewYAML[T])
}