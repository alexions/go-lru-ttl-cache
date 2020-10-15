package go_lru_ttl_cache

import (
	"sync"
	"time"
)

type LRUCache struct {
	values map[string]*cacheValue
	head   *lruQueueItem
	tail   *lruQueueItem
	size   uint32

	maxSize    uint32
	defaultTTL time.Duration

	sync.RWMutex
}

func NewLRUCache(config *ConfigBuilder) *LRUCache {
	cache := &LRUCache{
		values: make(map[string]*cacheValue),

		// Configuration
		maxSize:    config.maxSize,
		defaultTTL: config.defaultTTL,
	}

	// Package agreement: there is no way to stop this goroutine. Reuse cache if possible.
	// Avoid multiple create/delete cycles
	go func() {
		for {
			<-time.After(config.cleanInterval)
			cache.cleanInterval()
		}
	}()

	return cache
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.RLock()
	value, found := c.values[key]
	c.RUnlock()
	if !found {
		return nil, false
	}

	c.moveToTop(value.link)

	return value.data, true
}

func (c *LRUCache) Set(key string, value interface{}) {
	_, found := c.Get(key)
	if found {
		c.Lock()
		c.values[key].data = value
		c.Unlock()
	} else {
		item := &lruQueueItem{
			key: key,
			ttl: time.Duration(time.Now().Unix()) + c.defaultTTL,

			next: c.head,
		}

		c.Lock()
		c.head = item
		c.values[key] = &cacheValue{
			data: value,
			link: item,
		}
		c.Unlock()
	}
}

func (c *LRUCache) Delete(key string) {
	c.RLock()
	value, found := c.values[key]
	c.RUnlock()

	if found {
		c.Lock()
		c.unsafeDelete(key, value)
		c.Unlock()
	}

}

func (c *LRUCache) Clean() {
	// TODO: Check if GOCG cleans the dropped values and do not do a memory leaking
	c.Lock()
	c.values = make(map[string]*cacheValue)
	c.head = nil
	c.tail = nil
	c.Unlock()
}

// Cleans up the expired items. Do not set the clean interval too low to avoid CPU load
func (c *LRUCache) cleanInterval() {
	c.Lock()
	now := time.Duration(time.Now().Unix())
	for key, value := range c.values {
		if value.link.ttl < now {
			c.unsafeDelete(key, value)
		}
	}
	c.Unlock()
}

func (c *LRUCache) unsafeDelete(key string, value *cacheValue) {
	if value.link.prev != nil {
		value.link.prev = value.link.next
	}
	if value.link.next != nil {
		value.link.next = value.link.prev
	}

	if c.head == value.link {
		c.head = value.link.next
	}
	if c.tail == value.link {
		c.tail = value.link.prev
	}

	value.link = nil
	delete(c.values, key)
}

func (c *LRUCache) moveToTop(item *lruQueueItem) {
	c.Lock()
	if item.prev != nil {
		item.prev.next = item.next
	}
	if item.next != nil {
		item.next.prev = item.prev
	}

	item.next = c.head
	c.head = item

	item.ttl = time.Duration(time.Now().Unix()) + c.defaultTTL
	c.Unlock()
}
