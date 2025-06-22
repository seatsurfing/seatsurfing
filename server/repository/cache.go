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
	Cache *freecache.Cache
}

var cache *Cache
var cacheOnce sync.Once

func GetCache() *Cache {
	cacheOnce.Do(func() {
		if config.GetConfig().CacheType == "default" {
			cache = &Cache{
				Cache: freecache.NewCache(10 * 1024 * 1024), // 10 MB cache
			}
		}
	})
	return cache
}

func (c *Cache) Set(key string, value []byte, expireSeconds int) error {
	if config.GetConfig().CacheType == "valkey" {
		client, err := c.getValkeyClient()
		if err != nil {
			return errors.New("failed to get valkey client")
		}
		defer client.Close()
		ctx := context.Background()
		b := client.B().Set().Key(key).Value(string(value)).Ex(time.Duration(expireSeconds) * time.Second).Build()
		if err := client.Do(ctx, b).Error(); err != nil {
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
		client, err := c.getValkeyClient()
		if err != nil {
			return nil, errors.New("failed to get valkey client")
		}
		defer client.Close()
		ctx := context.Background()
		b := client.B().Get().Key(key).Build()
		res, err := client.Do(ctx, b).ToString()
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
		client, err := c.getValkeyClient()
		if err != nil {
			return errors.New("failed to get valkey client")
		}
		defer client.Close()
		ctx := context.Background()
		b := client.B().Del().Key(key).Build()
		if err := client.Do(ctx, b).Error(); err != nil {
			return errors.New("failed to delete cache")
		}
		return nil
	}

	c.Cache.Del([]byte(key))
	return nil
}

func (c *Cache) getValkeyClient() (valkey.Client, error) {
	cfg := config.GetConfig()
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.ValkeyHost},
		AuthCredentialsFn: func(ctx valkey.AuthCredentialsContext) (valkey.AuthCredentials, error) {
			return valkey.AuthCredentials{
				Username: cfg.ValkeyUsername,
				Password: cfg.ValkeyPassword,
			}, nil
		},
	})
	if err != nil {
		log.Println("Error creating Valkey client:", err)
		return nil, err
	}
	return client, nil
}
