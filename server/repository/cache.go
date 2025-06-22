package repository

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/coocood/freecache"
	"github.com/seatsurfing/seatsurfing/server/config"
	"github.com/valkey-io/valkey-go"
)

type Cache struct {
	Cache        *freecache.Cache
	ValkeyClient valkey.Client
}

var cache *Cache
var cacheOnce sync.Once

func GetCache() *Cache {
	cacheOnce.Do(func() {
		cfg := config.GetConfig()
		switch cfg.CacheType {
		case "valkey":
			client, err := valkey.NewClient(valkey.ClientOption{
				InitAddress: cfg.ValkeyHosts,
				AuthCredentialsFn: func(ctx valkey.AuthCredentialsContext) (valkey.AuthCredentials, error) {
					return valkey.AuthCredentials{
						Username: cfg.ValkeyUsername,
						Password: cfg.ValkeyPassword,
					}, nil
				},
			})
			if err != nil {
				log.Fatalln("Error creating Valkey client:", err)
				return
			}
			cache = &Cache{
				ValkeyClient: client,
			}
		case "default":
			cache = &Cache{
				Cache: freecache.NewCache(10 * 1024 * 1024), // 10 MB cache
			}
		}
	})
	return cache
}

func (c *Cache) Set(key string, value []byte, expireSeconds int) error {
	if config.GetConfig().CacheType == "valkey" {
		ctx := context.Background()
		b := c.ValkeyClient.B().Set().Key(key).Value(string(value)).Ex(time.Duration(expireSeconds) * time.Second).Build()
		if err := c.ValkeyClient.Do(ctx, b).Error(); err != nil {
			return errors.New("failed to set cache")
		}
		return nil
	}

	if err := c.Cache.Set([]byte(key), value, expireSeconds); err != nil {
		return errors.New("failed to set cache")
	}
	return nil
}

func (c *Cache) Get(key string) ([]byte, error) {
	if config.GetConfig().CacheType == "valkey" {
		ctx := context.Background()
		b := c.ValkeyClient.B().Get().Key(key).Build()
		res, err := c.ValkeyClient.Do(ctx, b).ToString()
		if err != nil {
			return nil, errors.New("cache miss")
		}
		return []byte(res), nil
	}

	value, err := c.Cache.Get([]byte(key))
	if err != nil {
		return nil, errors.New("cache miss")
	}
	return value, nil
}

func (c *Cache) Delete(key string) error {
	if config.GetConfig().CacheType == "valkey" {
		ctx := context.Background()
		b := c.ValkeyClient.B().Del().Key(key).Build()
		if err := c.ValkeyClient.Do(ctx, b).Error(); err != nil {
			return errors.New("failed to delete cache")
		}
		return nil
	}

	c.Cache.Del([]byte(key))
	return nil
}
