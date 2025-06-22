package repository

import (
	"errors"
	"sync"

	"github.com/coocood/freecache"
)

type Cache struct {
	Cache *freecache.Cache
}

var cache *Cache
var cacheOnce sync.Once

func GetCache() *Cache {
	cacheOnce.Do(func() {
		cache = &Cache{
			Cache: freecache.NewCache(10 * 1024 * 1024), // 10 MB cache
		}
	})
	return cache
}

func (c *Cache) Set(key string, value []byte, expireSeconds int) error {
	if err := c.Cache.Set([]byte(key), value, expireSeconds); err != nil {
		return errors.New("failed to set cache")
	}
	return nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	value, err := c.Cache.Get([]byte(key))
	if err != nil {
		return nil, errors.New("cache miss")
	}
	return value, nil
}

func (c *Cache) Delete(key string) error {
	c.Cache.Del([]byte(key))
	return nil
}
