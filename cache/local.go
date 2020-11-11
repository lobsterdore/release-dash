package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type LocalCacheAdapter struct {
	Client            *cache.Cache
	DefaultExpiration time.Duration
}

func NewLocalCacheAdapter(defaultExpirationMinutes int, cleanupIntervalMinutes int) *LocalCacheAdapter {
	defaultExpiration := time.Duration(defaultExpirationMinutes) * time.Minute
	c := cache.New(defaultExpiration, time.Duration(cleanupIntervalMinutes)*time.Minute)

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

func (c LocalCacheAdapter) Set(key string, value interface{}) {
	c.Client.Set(key, value, c.DefaultExpiration)
}
