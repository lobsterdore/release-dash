package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type LocalCacheService struct {
	Client            *cache.Cache
	DefaultExpiration time.Duration
}

func NewLocalCacheService(defaultExpirationMinutes int, cleanupIntervalMinutes int) CacheAdaptor {
	defaultExpiration := time.Duration(defaultExpirationMinutes) * time.Minute
	c := cache.New(defaultExpiration, time.Duration(cleanupIntervalMinutes)*time.Minute)

	service := LocalCacheService{
		Client:            c,
		DefaultExpiration: defaultExpiration,
	}

	return &service
}

func (c LocalCacheService) Get(key string) (interface{}, bool) {
	value, found := c.Client.Get(key)
	return value, found
}

func (c LocalCacheService) Set(key string, value interface{}) {
	c.Client.Set(key, value, c.DefaultExpiration)
}
