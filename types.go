package go_lru_ttl_cache

import "time"

type cacheValue struct {
	data interface{}
	link *lruQueueItem
}

type lruQueueItem struct {
	key string
	ttl time.Duration

	prev *lruQueueItem
	next *lruQueueItem
}
