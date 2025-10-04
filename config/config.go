package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"sigs.k8s.io/yaml"
)

// LoadConfig 从指定路径加载配置文件，并绑定到cfg上
func LoadConfig(path string, cfg any) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// 根据文件扩展名选择解析器
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		loadJSON(path, cfg)
	case ".toml":
		loadTOML(path, cfg)
	case ".yaml", ".yml":
		loadYAML(path, cfg)
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}
	return nil
}

// loadJSON 加载 JSON 配置文件
func loadJSON(path string, cfg any) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 JSON 文件失败: %w", err)
	}

	if err := json.Unmarshal(file, cfg); err != nil {
		return fmt.Errorf("解析 JSON 文件失败: %w", err)
	}

	return nil
}

// loadTOML 加载 TOML 配置文件
func loadTOML(path string, cfg any) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 TOML 文件失败: %w", err)
	}

	if _, err := toml.Decode(string(file), cfg); err != nil {
		return fmt.Errorf("解析 TOML 文件失败: %w", err)
	}

	return nil
}

// loadYAML 加载 YAML 配置文件
func loadYAML(path string, cfg any) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 YAML 文件失败: %w", err)
	}

	if err := yaml.Unmarshal(file, cfg); err != nil {
		return fmt.Errorf("解析 YAML 文件失败: %w", err)
	}

	return nil
}
