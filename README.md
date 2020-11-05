# LRU Cache with TTL support for Golang

LRUCache is an LRU cache written in Go. The cache supports TTL background 
cleaning.

**The LRUCache is thread-safe**

# Installation

Use a `go-get` tool (or any go-compatible package manager) to download the cache:

```bash
go get github.com/alexions/go-lru-ttl-cache
```

Import the cache in your go file:

```
import (
    lrucache "github.com/alexions/go-lru-ttl-cache"
)
```

It is highly recommended to use tags to lock the cache version for your project.


# Configuration

Import and create a Cache instance:

```
var cache = lrucache.NewLRUCache(lrucache.Configuration())
```

Configure allows you to configure cache using chainable API:

```
var cache = lrucache.NewLRUCache(
    lrucache.Configuration().
    SetDefaultTTL(5 * time.Minute).
    SetMaxSize(100)
)
```

Possible configuration with default values:

* `SetDefaultTTL(ttl time.Duration)` - default expiration TTL. Item will be removed after the specified time (default: -1, no expiration)
* `SetMaxSize(size int)` - cache storage limitation. After exceeding the limit the LRU item will be removed (default: math.MaxInt32 - 1)
* `SetCleanupInterval(interval time.Duration)` - the cleaning interval. All the items with expired TTL will be removed from the cache.
    Works only with DefaultTTL greater than 0 (default: run every 1 min)
* `SetDeleteCallback(callback func(count int64))` - callback function to get amount of removed items. Runs every `Cleanup interval`.


# Usage

## Set/Get

```
config := Configuration()
cache := NewLRUCache(config)
cache.Set("hello", "world")

value := cache.Get("hello")
fmt.Println(value) // "hello"
```

Any data can be stored

```
type User struct {
    login string
    pass  string
}

config := Configuration()
cache := NewLRUCache(config)

cache.Set(1, &User{login: "alexions", pass: "password"})

user := cache.Get(1)
fmt.Println(user.(*User).login)
```

Checks if item is set:
```
if _, found := cache.Get("my_item"); !found {
    // Item not found
}
```

## Size

```
config := Configuration()
cache := NewLRUCache(config)
cache.Set(1, "hello")
cache.Set(2, "world")

fmt.Println(cache.Size()) // 2
```

## Delete

```
config := Configuration()
cache := NewLRUCache(config)
cache.Set(1, "hello")

cache.Size() // 1
cache.Delete(1)
cache.Size() // 0
```

## Clean

```
config := Configuration()
cache := NewLRUCache(config)
cache.Set(1, "hello")
cache.Set(2, "world")

cache.Size() // 2
cache.Clean()
cache.Size() // 0
```

