package go_lru_ttl_cache

import (
	"container/list"
	"time"
)

type cacheValue struct {
	data interface{}
	link *list.Element
}

type lruQueueItem struct {
	key string
	ttl time.Duration
}
