package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cache "github.com/lobsterdore/release-dash/cache"
)

func TestCacheNotSet(t *testing.T) {
	localCache := cache.NewLocalCacheAdapter(60, 60)

	_, found := localCache.Get("test-key")

	assert.False(t, found)
}

func TestCacheSet(t *testing.T) {
	localCache := cache.NewLocalCacheAdapter(60, 60)

	localCache.Set("test-key", "test-value", "")

	testData, found := localCache.Get("test-key")

	assert.True(t, found)
	assert.Equal(t, testData.(string), "test-value")

}
