package cache

import (
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
)

// NewLRU creates a new, thread-safe LRU cache with a maximum size. When adding a key
// would exceed the maximum size, the least recently used key is evicted to make
// space. Invocations of Get or Add modify eviction ordering by marking the given key
// as the most recently used key. Invocations of ForEach do not modify eviction
// ordering.
func NewLRU(size int) (Cache, error) {
	underlying, err := simplelru.NewLRU(size, nil)
	if err != nil {
		return nil, err
	}

	return &lruCache{lru: underlying}, nil
}

type lruCache struct {
	lru  *simplelru.LRU
	lock sync.RWMutex
}

func (c *lruCache) Get(key interface{}) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.lru.Get(key)
}

// ForEach iterates over the key-value pairs in the Cache from least to most recently
// used.
func (c *lruCache) ForEach(f func(key, value interface{})) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, key := range c.lru.Keys() {
		value, _ := c.lru.Peek(key)
		f(key, value)
	}
}

func (c *lruCache) Add(key, value interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	existed := c.lru.Contains(key)
	c.lru.Add(key, value)
	return existed
}

func (c *lruCache) Remove(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.lru.Remove(key)
}

func (c *lruCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.lru.Purge()
}

func (c *lruCache) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.lru.Len()
}
