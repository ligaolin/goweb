package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gregjones/httpcache/diskcache"
)

type FileConfig struct {
	Path string
}

type File struct {
	mu     sync.RWMutex
	Client *diskcache.Cache
	Path   string
}

func NewFile(config *FileConfig) *File {
	f := &File{
		Client: diskcache.New(config.Path),
		Path:   config.Path,
	}
	go f.startCleanup()
	return f
}

type Value struct {
	Expire    time.Time
	HasExpire bool
	Value     any
}

func (f *File) Get(key string, value any) error {
	f.mu.RLock()
	str, ok := f.Client.Get(key)
	f.mu.RUnlock()
	if !ok {
		return errors.New("从文件缓存获取数据失败")
	}

	var val Value
	if err := json.Unmarshal(str, &val); err != nil {
		return err
	}

	if val.HasExpire && time.Now().After(val.Expire) {
		f.Delete(key)
		return errors.New("缓存数据已过期")
	}

	data, err := json.Marshal(val.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func (f *File) Set(key string, value any, expire time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	val := Value{
		Value: value,
	}
	if expire > 0 {
		val.HasExpire = true
		val.Expire = time.Now().Add(expire)
	}

	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	f.Client.Set(key, data)
	return nil
}

func (f *File) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Client.Delete(key)
	return nil
}

func (f *File) startCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		f.cleanup()
	}
}

func (f *File) cleanup() {
	entries, err := os.ReadDir(f.Path)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key := filepath.Base(entry.Name())
		var val Value

		f.mu.RLock()
		str, ok := f.Client.Get(key)
		f.mu.RUnlock()
		if !ok {
			continue
		}

		if err := json.Unmarshal(str, &val); err != nil {
			continue
		}

		if val.HasExpire && time.Now().After(val.Expire) {
			f.Delete(key)
		}
	}
}
