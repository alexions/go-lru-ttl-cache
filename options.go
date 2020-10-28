package go_lru_ttl_cache

import (
	"math"
	"time"
)

type ConfigBuilder struct {
	defaultTTL     time.Duration
	maxSize        int
	cleanInterval  time.Duration
	deleteCallback func(count int64)
}

func Configuration() *ConfigBuilder {
	return &ConfigBuilder{
		defaultTTL:     -1,
		maxSize:        math.MaxInt32 - 1,
		cleanInterval:  time.Minute,
		deleteCallback: nil,
	}
}

func (b *ConfigBuilder) SetDefaultTTL(ttl time.Duration) *ConfigBuilder {
	b.defaultTTL = ttl
	return b
}

func (b *ConfigBuilder) SetMaxSize(size int) *ConfigBuilder {
	b.maxSize = size
	return b
}

func (b *ConfigBuilder) SetCleanupInterval(interval time.Duration) *ConfigBuilder {
	b.cleanInterval = interval
	return b
}

func (b *ConfigBuilder) SetDeleteCallback(callback func(count int64)) *ConfigBuilder {
	b.deleteCallback = callback
	return b
}
