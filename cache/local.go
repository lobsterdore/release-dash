package cache

import (
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

type LocalCacheAdapter struct {
	Client            *cache.Cache
	DefaultExpiration time.Duration
}

func NewLocalCacheAdapter(DefaultExpirationSeconds int, CleanupIntervalSeconds int) *LocalCacheAdapter {
	defaultExpiration := time.Duration(DefaultExpirationSeconds) * time.Second
	c := cache.New(defaultExpiration, time.Duration(CleanupIntervalSeconds)*time.Second)

	service := LocalCacheAdapter{
		Client:            c,
		DefaultExpiration: defaultExpiration,
	}

	return &service
}

func (c LocalCacheAdapter) Get(key string) (interface{}, bool) {
	value, found := c.Client.Get(key)
	return value, found
}

func (c LocalCacheAdapter) Set(key string, value interface{}, expireSeconds string) {
	expiration := c.DefaultExpiration
	if expireSeconds != "" {
		if i, err := strconv.Atoi(expireSeconds); err == nil {
			expiration = time.Duration(i) * time.Second
		}
	}
	c.Client.Set(key, value, expiration)
}
