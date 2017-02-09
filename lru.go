package cache

import lru "github.com/hashicorp/golang-lru"

// NewLRU create a new, thread-safe LRU cache with a maximum size
func NewLRU(size int) (Cache, error) {
	underlying, err := lru.New(size)
	if err != nil {
		return nil, err
	}

	return &lruCache{underlying}, nil
}

type lruCache struct {
	*lru.Cache
}

func (c *lruCache) Clear() {
	c.Cache.Purge()
}

func (c *lruCache) Remove(key interface{}) bool {
	c.Cache.Remove(key)
	return true
}
