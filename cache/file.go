package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gregjones/httpcache/diskcache"
)

type File struct {
	Client *diskcache.Cache
	Path   string
}

type Value struct {
	Expir   time.Time
	IsExpir bool // 是否会过期，true会过期，false不会过期
	Value   any
}

func (f *File) Get(key string, value any) error {
	// 获取文件缓存的数据
	str, ok := f.Client.Get(key)
	if !ok {
		return errors.New("从文件缓存获取数据失败")
	}

	// 反序列化文件缓存数据
	var val Value
	if err := json.Unmarshal([]byte(string(str)), &val); err != nil {
		return err
	}

	// 判断是否过期
	if val.IsExpir {
		if time.Now().After(val.Expir) {
			f.Delete(key)
			return errors.New("缓存数据已过期")
		}
	}

	// 反序列化存储数据
	if err := json.Unmarshal([]byte(val.Value.(string)), value); err != nil {
		return err
	}

	// 清理过期文件缓存
	go f.Clean()
	return nil
}

func (f *File) Set(key string, value any, expir time.Duration) error {
	valStr, err := json.Marshal(value)
	if err != nil {
		return err
	}

	isExpir := true
	if expir <= -1 {
		isExpir = false
	}
	v, err := json.Marshal(Value{Value: string(valStr), Expir: time.Now().Add(expir), IsExpir: isExpir})
	if err != nil {
		return err
	}
	f.Client.Set(key, v)
	return nil
}

func (f *File) Delete(key string) error {
	f.Client.Delete(key)
	return nil
}

// 清理过期文件缓存
func (f *File) Clean() error {
	files, err := os.ReadDir(f.Path)
	if err != nil {
		return err
	}

	var val Value
	for _, file := range files {
		path := filepath.Join(f.Path, file.Name())
		bytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if err = json.Unmarshal(bytes, &val); err != nil {
			continue
		}

		if val.IsExpir {
			if time.Now().After(val.Expir) {
				os.Remove(path)
			}
		}
	}
	return nil
}
