package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type LocalCacheAdaptor struct {
	Client            *cache.Cache
	DefaultExpiration time.Duration
}

func NewLocalCacheAdaptor(defaultExpirationMinutes int, cleanupIntervalMinutes int) CacheAdaptor {
	defaultExpiration := time.Duration(defaultExpirationMinutes) * time.Minute
	c := cache.New(defaultExpiration, time.Duration(cleanupIntervalMinutes)*time.Minute)

	service := LocalCacheAdaptor{
		Client:            c,
		DefaultExpiration: defaultExpiration,
	}

	return &service
}

func (c LocalCacheAdaptor) Get(key string) (interface{}, bool) {
	value, found := c.Client.Get(key)
	return value, found
}

func (c LocalCacheAdaptor) Set(key string, value interface{}) {
	c.Client.Set(key, value, c.DefaultExpiration)
}
