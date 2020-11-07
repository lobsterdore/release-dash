package cache

//go:generate go run -mod=mod github.com/golang/mock/mockgen --build_flags=-mod=mod --source=cache.go --destination=../mocks/cache/cache.go
type CacheAdaptor interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}
