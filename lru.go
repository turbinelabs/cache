/*
Copyright 2017 Turbine Labs, Inc.

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
