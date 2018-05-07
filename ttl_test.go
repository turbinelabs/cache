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
	"testing"
	"time"

	tbntime "github.com/turbinelabs/nonstdlib/time"
	"github.com/turbinelabs/test/assert"
)

func TestNewTTL(t *testing.T) {
	c, err := NewTTL(0, 1*time.Second)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "positive size")

	c, err = NewTTL(10, 0)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "positive TTL")

	c, err = NewTTL(10, 10*time.Second)
	assert.Nil(t, err)
	assert.NonNil(t, c)
	assert.Equal(t, c.(*ttlLruCache).size, 10)
	assert.Equal(t, c.(*ttlLruCache).ttl, 10*time.Second)
	assert.NonNil(t, c.(*ttlLruCache).lru)
	assert.NonNil(t, c.(*ttlLruCache).timeSource)
}

func TestTTLCacheBasicOperations(t *testing.T) {
	c, err := NewTTL(10, 5*time.Minute)
	assert.Nil(t, err)

	assert.False(t, c.Add("k1", "v1"))
	assert.False(t, c.Add("k2", "v2"))
	assert.False(t, c.Add("k3", "v3"))
	assert.True(t, c.Add("k1", "v1-again"))
	assert.Equal(t, c.Len(), 3)

	assert.True(t, c.Remove("k3"))
	assert.Equal(t, c.Len(), 2)

	assert.False(t, c.Remove("never-added"))
	assert.Equal(t, c.Len(), 2)

	v1, ok1 := c.Get("k1")
	assert.True(t, ok1)
	assert.Equal(t, v1, "v1-again")

	v2, ok2 := c.Get("k2")
	assert.True(t, ok2)
	assert.Equal(t, v2, "v2")

	c.Clear()
	assert.Equal(t, c.Len(), 0)
}

func TestTTLCacheLRUBehavior(t *testing.T) {
	c, err := NewTTL(3, 5*time.Minute)
	assert.Nil(t, err)

	for i := 1; i <= 6; i++ {
		assert.False(t, c.Add(i, i+100))

		if i >= 4 {
			v, ok := c.Get(i - 3)
			assert.Nil(t, v)
			assert.False(t, ok)
		}
	}

	for i := 4; i <= 6; i++ {
		v, ok := c.Get(i)
		assert.Equal(t, v, i+100)
		assert.True(t, ok)
	}
}

func TestTTLCacheForEachLRUBehavior(t *testing.T) {
	c, err := NewTTL(5, 5*time.Minute)
	assert.Nil(t, err)

	for i := 1; i <= 5; i++ {
		assert.False(t, c.Add(i, i+100))
	}

	keys := []int{}
	c.ForEach(func(k, v interface{}) {
		i := k.(int)
		assert.Equal(t, v.(int), i+100)
		keys = append(keys, i)
	})
	assert.ArrayEqual(t, keys, []int{1, 2, 3, 4, 5})

	// Re-order the LRU eviction list.
	for idx := len(keys) - 1; idx >= 0; idx-- {
		c.Get(keys[idx])
	}

	keys2 := []int{}
	c.ForEach(func(k, _ interface{}) {
		keys2 = append(keys2, k.(int))
	})
	assert.ArrayEqual(t, keys2, []int{5, 4, 3, 2, 1})

	c.Add(100, 200)
	keys3 := []int{}
	c.ForEach(func(k, _ interface{}) {
		keys3 = append(keys3, k.(int))
	})
	assert.ArrayEqual(t, keys3, []int{4, 3, 2, 1, 100})
}

func TestTTLCacheExpiry(t *testing.T) {
	c, err := NewTTL(10, 10*time.Second)
	assert.Nil(t, err)

	tbntime.WithCurrentTimeFrozen(func(ts tbntime.ControlledSource) {
		c.(*ttlLruCache).timeSource = ts

		for i := 1; i <= 3; i++ {
			c.Add(i, i+100)
			ts.Advance(1 * time.Second)
		}

		ts.Advance(6 * time.Second)

		for i := 1; i <= 3; i++ {
			v, ok := c.Get(i)
			assert.Equal(t, v, i+100)
			assert.True(t, ok)

			ts.Advance(1 * time.Second)

			v, ok = c.Get(i)
			assert.Nil(t, v)
			assert.False(t, ok)
		}

		assert.Equal(t, c.Len(), 0)
		c.Clear()

		tNow := ts.Now()
		for i := 1; i <= 10; i++ {
			c.Add(i, i+100)
			ts.Advance(10 * time.Millisecond)
		}
		for i := 10; i >= 1; i-- {
			c.Get(i)
		}

		ts.Set(tNow.Add(10*time.Second + 1*time.Nanosecond))

		// trigger eviction, but it should find the expired key 1 not the LRU key 10
		c.Add(100, 200)

		v, ok := c.Get(1)
		assert.Nil(t, v)
		assert.False(t, ok)

		for i := 2; i <= 10; i++ {
			v, ok := c.Get(i)
			assert.Equal(t, v, i+100)
			assert.True(t, ok)
		}

		v, ok = c.Get(100)
		assert.Equal(t, v, 200)
		assert.True(t, ok)
	})
}

func TestTTLCacheExpiryInForEach(t *testing.T) {
	c, err := NewTTL(10, 10*time.Second)
	assert.Nil(t, err)

	tbntime.WithCurrentTimeFrozen(func(ts tbntime.ControlledSource) {
		c.(*ttlLruCache).timeSource = ts

		for i := 1; i <= 3; i++ {
			c.Add(i, i+100)
			ts.Advance(1 * time.Second)
		}

		// Reverse the LRU ordering
		c.Get(3)
		c.Get(1)
		c.Get(2)

		ts.Advance(7 * time.Second)

		keys := []int{}
		c.ForEach(func(k, v interface{}) {
			i := k.(int)
			assert.Equal(t, v.(int), i+100)
			keys = append(keys, i)
		})
		assert.ArrayEqual(t, keys, []int{3, 2})
	})
}
