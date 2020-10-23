package go_lru_ttl_cache

import (
	"strconv"
	"testing"
	"time"
)

func BenchmarkGetSetSimple(b *testing.B) {
	config := Configuration().
		SetMaxSize(500000)
	cache := NewLRUCache(config)

	cacheKeys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}

	for i := 0; i < b.N; i++ {
		key := cacheKeys[i%10]
		cache.Set(key, i)
		cache.Get(key)
	}
}

func BenchmarkGetSetIncrementKeys(b *testing.B) {
	config := Configuration().
		SetMaxSize(500000)
	cache := NewLRUCache(config)

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cache.Set(key, i)
		cache.Get(key)
	}
}

func BenchmarkFrequentCleanup(b *testing.B) {
	config := Configuration().
		SetCleanupInterval(10 * time.Millisecond).
		SetDefaultTTL(20 * time.Millisecond).
		SetMaxSize(500000)
	cache := NewLRUCache(config)

	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cache.Set(key, i)
		cache.Get(key)
	}
}
