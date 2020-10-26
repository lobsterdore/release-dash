package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=cache.go --destination=../mocks/cache/cache.go
type CacheProvider interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type CacheService struct {
	Client            *cache.Cache
	DefaultExpiration time.Duration
}

func NewCacheService(defaultExpirationMinutes int, cleanupIntervalMinutes int) CacheProvider {
	defaultExpiration := time.Duration(defaultExpirationMinutes) * time.Minute
	c := cache.New(defaultExpiration, time.Duration(cleanupIntervalMinutes)*time.Minute)

	service := CacheService{
		Client:            c,
		DefaultExpiration: defaultExpiration,
	}

	return &service
}

func (c CacheService) Get(key string) (interface{}, bool) {
	value, found := c.Client.Get(key)
	return value, found
}

func (c CacheService) Set(key string, value interface{}) {
	c.Client.Set(key, value, c.DefaultExpiration)
}
