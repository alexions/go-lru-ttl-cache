package go_lru_ttl_cache

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	values map[interface{}]*cacheValue
	queue  *list.List

	maxSize        int
	defaultTTL     time.Duration
	deleteCallback func(count int64)

	lock sync.RWMutex
}

func NewLRUCache(config *ConfigBuilder) *LRUCache {
	cache := &LRUCache{
		values: make(map[interface{}]*cacheValue),
		queue:  list.New(),

		// Configuration
		maxSize:        config.maxSize,
		defaultTTL:     config.defaultTTL,
		deleteCallback: config.deleteCallback,
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

func (c *LRUCache) Get(key interface{}) (interface{}, bool) {
	c.lock.RLock()
	value, found := c.values[key]
	c.lock.RUnlock()
	if !found {
		return nil, false
	}

	c.lock.Lock()
	c.queue.MoveToFront(value.link)
	c.lock.Unlock()

	return value.data, true
}

func (c *LRUCache) Set(key interface{}, value interface{}) {
	c.lock.Lock()
	_, found := c.values[key]
	if found {
		c.values[key].data = value
		c.queue.MoveToFront(c.values[key].link)
		c.lock.Unlock()
		return
	}
	c.lock.Unlock()

	LRUitem := &lruQueueItem{
		key: key,
		ttl: time.Now().Add(c.defaultTTL),
	}
	c.lock.Lock()
	queueItem := c.queue.PushFront(LRUitem)
	c.values[key] = &cacheValue{
		data: value,
		link: queueItem,
	}
	c.lock.Unlock()

	if c.queue.Len() > c.maxSize {
		c.lock.Lock()
		item := c.queue.Back().Value.(*lruQueueItem)
		cachedItem := c.values[item.key]
		c.unsafeDelete(item.key, cachedItem)
		c.lock.Unlock()
	}
}

func (c *LRUCache) Delete(key interface{}) {
	c.lock.RLock()
	value, found := c.values[key]
	c.lock.RUnlock()

	if found {
		c.lock.Lock()
		c.unsafeDelete(key, value)
		c.lock.Unlock()
		if c.deleteCallback != nil {
			c.deleteCallback(1)
		}
	}
}

func (c *LRUCache) Clean() {
	// TODO: Check if GOCG cleans the dropped values and do not do a memory leaking
	c.lock.Lock()
	c.values = make(map[interface{}]*cacheValue)
	c.queue = list.New()
	c.lock.Unlock()
}

func (c *LRUCache) Size() int {
	return c.queue.Len()
}

// Cleans up the expired items. Do not set the clean interval too low to avoid CPU load
func (c *LRUCache) cleanInterval() {
	var deleted int64
	c.lock.Lock()
	for key, value := range c.values {
		item := value.link.Value.(*lruQueueItem)
		if item.ttl.Sub(time.Now()) < 0 {
			c.unsafeDelete(key, value)
			deleted++
		}
	}
	c.lock.Unlock()

	if c.deleteCallback != nil {
		c.deleteCallback(deleted)
	}
}

func (c *LRUCache) unsafeDelete(key interface{}, value *cacheValue) {
	c.queue.Remove(value.link)
	delete(c.values, key)
	value = nil
}
