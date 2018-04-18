package cache

import (
	"errors"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/simplelru"

	tbntime "github.com/turbinelabs/nonstdlib/time"
)

type entry struct {
	deadline time.Time
	value    interface{}
}

// NewTTL create a new cache with a maximum size and a TTL for cache entries. When
// the cache is full and a new key is added, a linear search is undertaken to find an
// expired cache entry for eviction before evicting the least recently used cache
// entry. Invocations of ForEach do not modify the LRU eviction list but expired
// items are never returned from ForEach.
func NewTTL(size int, ttl time.Duration) (Cache, error) {
	underlying, err := lru.NewLRU(size, nil)
	if err != nil {
		return nil, err
	}

	if ttl <= 0 {
		return nil, errors.New("Must provide a positive TTL")
	}

	return &ttlLruCache{
		lru:        underlying,
		size:       size,
		ttl:        ttl,
		timeSource: tbntime.NewSource(),
	}, nil
}

type ttlLruCache struct {
	lru        *lru.LRU
	lock       sync.Mutex
	size       int
	ttl        time.Duration
	timeSource tbntime.Source
}

func (c *ttlLruCache) expired(e *entry) bool {
	return !c.timeSource.Now().Before(e.deadline)
}

func (c *ttlLruCache) Add(key, value interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.getEntry(key, false)
	if c.lru.Len() >= c.size && !exists {
		// Look for expired entries to evict to avoid
		// potentially evicting a live entry.
		keys := c.lru.Keys()
		for _, key := range keys {
			if v, ok := c.lru.Peek(key); ok {
				entry := v.(*entry)
				if c.expired(entry) {
					c.lru.Remove(key)
					break
				}
			}
		}
	}

	c.lru.Add(key, &entry{c.timeSource.Now().Add(c.ttl), value})
	return exists
}

func (c *ttlLruCache) Remove(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.get(key); ok {
		return c.lru.Remove(key)
	}

	return false
}

func (c *ttlLruCache) Get(key interface{}) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.get(key)
}

// ForEach iterates over the non-expired key-value pairs in the Cache from least to
// most recently used.
func (c *ttlLruCache) ForEach(f func(key, value interface{})) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, key := range c.lru.Keys() {
		if entry, ok := c.getEntry(key, true); ok {
			f(key, entry.value)
		}
	}
}

func (c *ttlLruCache) get(key interface{}) (interface{}, bool) {
	if entry, ok := c.getEntry(key, false); ok {
		return entry.value, true
	}

	return nil, false
}

func (c *ttlLruCache) getEntry(key interface{}, peek bool) (*entry, bool) {
	var (
		v  interface{}
		ok bool
	)

	if peek {
		v, ok = c.lru.Peek(key)
	} else {
		v, ok = c.lru.Get(key)
	}

	if !ok {
		return nil, false
	}

	entry := v.(*entry)
	if c.expired(entry) {
		c.lru.Remove(key)
		return nil, false
	}

	return entry, true
}

func (c *ttlLruCache) Len() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.lru.Len()
}

func (c *ttlLruCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.lru.Purge()
}
