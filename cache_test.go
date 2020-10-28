package go_lru_ttl_cache

import (
	"testing"
	"time"
)

func TestLRUCache_Set(t *testing.T) {
	config := Configuration()
	cache := NewLRUCache(config)
	cache.Set("hello", "world")

	if value, found := cache.Get("hello"); found {
		if value.(string) != "world" {
			t.Fatal("wrong cached value")
		}
	} else {
		t.Fatal("unable to get value")
	}
}

func TestLRUCache_TTL(t *testing.T) {
	config := Configuration().
		SetCleanupInterval(10 * time.Millisecond).
		SetDefaultTTL(25)
	cache := NewLRUCache(config)

	cache.Set("remove_by_ttl", 1)
	cache.Set("keep_by_ttl", 2)

	if cache.Size() != 2 {
		t.Fatal("cache size is wrong")
	}

	<-time.After(15 * time.Millisecond)
	if _, ok := cache.Get("remove_by_ttl"); ok {
		t.Fatal("element has not been removed after TTL")
	}
	if _, ok := cache.Get("keep_by_ttl"); ok {
		t.Fatal("element has been removed by non-expired TTL")
	}

	<-time.After(15 * time.Millisecond)
	if cache.Size() != 0 {
		t.Fatal("cache size is wrong. All elements must be removed by TTL")
	}
}

func TestLRUCache_Overflow(t *testing.T) {
	config := Configuration().
		SetMaxSize(2)

	cache := NewLRUCache(config)
	cache.Set("1", 1)
	cache.Set("2", 2)
	if cache.Size() != 2 {
		t.Fatal("cache size is wrong. Added 2 elements to 2-len queue, but got", cache.Size())
	}

	cache.Set("3", int(3))
	if cache.Size() != 2 {
		t.Fatal("cache size is wrong. Over the limit")
	}

	if _, ok := cache.Get("1"); ok {
		t.Fatal("item must be removed by queue size")
	}

	if val, ok := cache.Get("3"); !ok || val.(int) != 3 {
		t.Fatal("item must not be removed but was or item has a wrong value")
	}
}

func TestLRUCache_LRUMove(t *testing.T) {
	config := Configuration().
		SetMaxSize(2)
	cache := NewLRUCache(config)
	cache.Set("1", int(1)) // queue: 1
	cache.Set("2", int(2)) // queue: 2, 1

	cache.Get("1")         // LRU move: queue: 1, 2
	cache.Set("3", int(3)) // queue: 3, 1

	if cache.Size() != 2 {
		t.Fatal("cache size is wrong. Over the limit")
	}

	if _, ok := cache.Get("2"); ok {
		t.Fatal("item must be removed but was not")
	}

	if val, ok := cache.Get("1"); !ok || val.(int) != 1 {
		t.Fatal("item must not be removed but was or item has a wrong value")
	}

	if val, ok := cache.Get("3"); !ok || val.(int) != 3 {
		t.Fatal("item must not be removed but was or item has a wrong value")
	}
}

func TestLRUCache_LRUMoveSet(t *testing.T) {
	config := Configuration().
		SetMaxSize(2)
	cache := NewLRUCache(config)
	cache.Set("1", int(1)) // queue: 1
	cache.Set("2", int(2)) // queue: 2, 1
	cache.Set("1", 5)      // LRU move via Set: queue: 1, 2
	cache.Set("3", int(3)) // queue: 3, 1

	if cache.Size() != 2 {
		t.Fatal("cache size is wrong. Over the limit")
	}

	if _, ok := cache.Get("2"); ok {
		t.Fatal("item must be removed but was not")
	}

	if val, ok := cache.Get("1"); !ok || val.(int) != 5 {
		t.Fatal("item must not be removed but was or item has a wrong value")
	}

	if val, ok := cache.Get("3"); !ok || val.(int) != 3 {
		t.Fatal("item must not be removed but was or item has a wrong value")
	}
}

func TestLRUCache_DeleteCallbackL(t *testing.T) {
	expectation := []int64{1, 0, 2}
	callback := func(count int64) {
		if count != expectation[0] {
			t.Fatal("wrong deleted count")
		}
		expectation = expectation[1:]

	}
	config := Configuration().
		SetCleanupInterval(10 * time.Millisecond).
		SetDefaultTTL(15 * time.Millisecond).
		SetDeleteCallback(callback)
	cache := NewLRUCache(config)

	cache.Set("hello", 1)
	cache.Delete("hello")

	cache.Set("to_be_removed", 1)
	cache.Set("to_be_removed2", 2)

	<-time.After(25 * time.Millisecond)

	if len(expectation) != 0 {
		t.Fatal("not all callbacks have been called")
	}
}
