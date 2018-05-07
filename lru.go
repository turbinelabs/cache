/*
Copyright 2018 Turbine Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
