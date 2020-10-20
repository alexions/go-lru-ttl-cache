package go_lru_ttl_cache

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	values map[string]*cacheValue
	queue  *list.List

	maxSize    int
	defaultTTL time.Duration

	sync.RWMutex
}

func NewLRUCache(config *ConfigBuilder) *LRUCache {
	cache := &LRUCache{
		values: make(map[string]*cacheValue),
		queue:  list.New(),

		// Configuration
		maxSize:    config.maxSize,
		defaultTTL: config.defaultTTL,
	}

	// Package agreement: there is no way to stop this goroutine. Reuse cache if possible.
	// Avoid multiple create/delete cycles
	if config.defaultTTL >= 0 {
		go func() {
			for {
				<-time.After(config.cleanInterval)
				cache.cleanInterval()
			}
		}()
	}

	return cache
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.RLock()
	value, found := c.values[key]
	c.RUnlock()
	if !found {
		return nil, false
	}

	c.Lock()
	c.queue.MoveToFront(value.link)
	c.Unlock()

	return value.data, true
}

func (c *LRUCache) Set(key string, value interface{}) {
	_, found := c.Get(key)
	if found {
		c.Lock()
		c.values[key].data = value
		c.Unlock()
	} else {
		LRUitem := &lruQueueItem{
			key: key,
			ttl: time.Duration(time.Now().Unix()) + c.defaultTTL,
		}

		c.Lock()
		queueItem := c.queue.PushFront(LRUitem)
		c.values[key] = &cacheValue{
			data: value,
			link: queueItem,
		}
		c.Unlock()

		if c.queue.Len() > c.maxSize {
			c.RLock()
			item := c.queue.Back().Value.(*lruQueueItem)
			c.RUnlock()
			c.Delete(item.key)
		}
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
	c.queue = list.New()
	c.Unlock()
}

// Cleans up the expired items. Do not set the clean interval too low to avoid CPU load
func (c *LRUCache) cleanInterval() {
	c.Lock()
	now := time.Duration(time.Now().Unix())
	for key, value := range c.values {
		item := value.link.Value.(*lruQueueItem)
		if item.ttl < now {
			c.unsafeDelete(key, value)
		}
	}
	c.Unlock()
}

func (c *LRUCache) unsafeDelete(key string, value *cacheValue) {
	c.queue.Remove(value.link)
	delete(c.values, key)
}
