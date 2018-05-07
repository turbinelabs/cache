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
	"fmt"
	"sort"
	"testing"

	"github.com/turbinelabs/nonstdlib/arrays/dedupe"
	"github.com/turbinelabs/nonstdlib/arrays/indexof"
	"github.com/turbinelabs/test/assert"
)

func contains(t *testing.T, c CountingCache, elems []string) {
	values := []string{}

	c.ForEach(func(key string, n int) {
		values = append(values, fmt.Sprintf("%s:%d", key, n))
	})

	assert.HasSameElements(t, values, elems)
}

func TestNewCountingCache(t *testing.T) {
	c, err := NewCountingCache(1)
	assert.ErrorContains(t, err, "minimum counting cache size")

	c, err = NewCountingCache(2)
	assert.Nil(t, err)
	assert.NonNil(t, c)
	impl := c.(*counting)
	assert.Equal(t, impl.size, 2)
	assert.NonNil(t, impl.lookup)
	assert.NonNil(t, impl.counts)
	assert.NonNil(t, impl.rng)
}

func TestCountingCacheAdd(t *testing.T) {
	c, _ := NewCountingCache(5)

	assert.Equal(t, c.Add("a", 1), 1)
	assert.Equal(t, c.Add("b", 1), 1)
	assert.Equal(t, c.Add("c", 1), 1)
	assert.Equal(t, c.Add("d", 1), 1)
	assert.Equal(t, c.Add("e", 1), 1)
	assert.Equal(t, c.Add("a", 1), 2)
	assert.Equal(t, c.Add("c", 2), 3)
	assert.Equal(t, c.Add("e", 3), 4)
	assert.Equal(t, c.Len(), 5)

	contains(t, c, []string{
		"a:2",
		"b:1",
		"c:3",
		"d:1",
		"e:4",
	})
}

func TestCountingCacheGet(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Add("a", 5)
	c.Add("b", 1)
	c.Add("c", 4)
	c.Add("d", 2)
	c.Add("e", 3)

	assert.Equal(t, c.Get("a"), 5)
	assert.Equal(t, c.Get("c"), 4)
	assert.Equal(t, c.Get("e"), 3)
	assert.Equal(t, c.Get("d"), 2)
	assert.Equal(t, c.Get("b"), 1)
	assert.Equal(t, c.Get("X"), 0)
}

func TestCountingCacheAddEvictsRandom(t *testing.T) {
	evicted := []int{}
	for repeat := 0; repeat < 100; repeat++ {
		c, _ := NewCountingCache(5)

		c.Add("a", 1)
		c.Add("b", 1)
		c.Add("c", 1)
		c.Add("d", 1)
		c.Add("e", 1)
		assert.Equal(t, c.Len(), 5)

		c.Add("f", 1)
		assert.Equal(t, c.Len(), 5)

		values := []string{}
		c.ForEach(func(key string, n int) {
			values = append(values, fmt.Sprintf("%s:%d", key, n))
		})
		sort.Strings(values)

		possible := []string{
			"a:1",
			"b:1",
			"c:1",
			"d:1",
			"e:1",
			"f:1",
		}

		found := 0
		for pi, p := range possible {
			if n := indexof.String(values, p); n == indexof.NotFound {
				evicted = append(evicted, pi)
			} else {
				found++
			}
		}

		assert.Equal(t, found, 5)
	}

	assert.Equal(t, len(evicted), 100)
	assert.GreaterThan(t, len(dedupe.Ints(evicted)), 1)
}

func TestCountingCacheAddEvictsMinCount(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Add("a", 5)
	c.Add("b", 4)
	c.Add("c", 3)
	c.Add("d", 2)
	c.Add("e", 1)
	assert.Equal(t, c.Len(), 5)

	c.Add("f", 1)
	assert.Equal(t, c.Len(), 5)

	contains(t, c, []string{
		"a:5",
		"b:4",
		"c:3",
		"d:2",
		"f:1",
	})
}

func TestCountingCacheAddDoesNotEvictOnUpdate(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Add("a", 5)
	c.Add("b", 4)
	c.Add("c", 3)
	c.Add("d", 2)
	c.Add("e", 1)
	assert.Equal(t, c.Len(), 5)

	c.Add("a", 1)
	assert.Equal(t, c.Len(), 5)

	contains(t, c, []string{
		"a:6",
		"b:4",
		"c:3",
		"d:2",
		"e:1",
	})
}

func TestCountingCacheAddDoesNotEvictOnZeroAdd(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Add("a", 5)
	c.Add("b", 4)
	c.Add("c", 3)
	c.Add("d", 2)
	c.Add("e", 1)
	assert.Equal(t, c.Len(), 5)

	c.Add("f", 0)
	assert.Equal(t, c.Len(), 5)

	contains(t, c, []string{
		"a:5",
		"b:4",
		"c:3",
		"d:2",
		"e:1",
	})
}

func TestCountingCacheAddRemovesOnZeroCount(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Add("a", 5)
	c.Add("b", 4)
	c.Add("c", 3)
	c.Add("d", 2)
	c.Add("e", 1)
	assert.Equal(t, c.Len(), 5)

	c.Add("e", -1)
	assert.Equal(t, c.Len(), 4)

	contains(t, c, []string{
		"a:5",
		"b:4",
		"c:3",
		"d:2",
	})

	c.Add("a", -5)
	assert.Equal(t, c.Len(), 3)

	contains(t, c, []string{
		"b:4",
		"c:3",
		"d:2",
	})
}

func TestCountingCacheIncDec(t *testing.T) {
	c, _ := NewCountingCache(5)

	assert.Equal(t, c.Inc("a"), 1)
	assert.Equal(t, c.Inc("b"), 1)
	assert.Equal(t, c.Inc("c"), 1)
	assert.Equal(t, c.Inc("d"), 1)
	assert.Equal(t, c.Inc("e"), 1)

	assert.Equal(t, c.Inc("a"), 2)
	assert.Equal(t, c.Inc("b"), 2)
	assert.Equal(t, c.Inc("c"), 2)
	assert.Equal(t, c.Inc("d"), 2)

	assert.Equal(t, c.Inc("f"), 1)

	contains(t, c, []string{
		"a:2",
		"b:2",
		"c:2",
		"d:2",
		"f:1",
	})

	assert.Equal(t, c.Dec("a"), 1)
	assert.Equal(t, c.Dec("b"), 1)
	assert.Equal(t, c.Dec("c"), 1)
	assert.Equal(t, c.Dec("d"), 1)

	contains(t, c, []string{
		"a:1",
		"b:1",
		"c:1",
		"d:1",
		"f:1",
	})

	assert.Equal(t, c.Dec("a"), 0)
	assert.Equal(t, c.Dec("b"), 0)
	assert.Equal(t, c.Dec("c"), 0)
	assert.Equal(t, c.Dec("d"), 0)
	assert.Equal(t, c.Dec("e"), -1)

	contains(t, c, []string{
		"e:-1",
		"f:1",
	})
}

func TestCountingCacheRemove(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Inc("a")
	c.Inc("b")
	c.Inc("c")
	c.Inc("d")
	c.Inc("e")

	assert.Equal(t, c.Remove("b"), 1)
	assert.Equal(t, c.Remove("d"), 1)

	contains(t, c, []string{
		"a:1",
		"c:1",
		"e:1",
	})

	assert.Equal(t, c.Remove("e"), 1)
	assert.Equal(t, c.Remove("c"), 1)
	assert.Equal(t, c.Remove("a"), 1)
	assert.Equal(t, c.Len(), 0)

	assert.Equal(t, c.Remove("a"), 0)
	assert.Equal(t, c.Len(), 0)
}

func TestCountingCacheClear(t *testing.T) {
	c, _ := NewCountingCache(5)

	c.Add("a", 1)
	c.Add("b", 2)
	c.Add("c", 3)
	c.Add("d", 4)
	c.Add("e", 5)

	contains(t, c, []string{
		"a:1",
		"b:2",
		"c:3",
		"d:4",
		"e:5",
	})

	c.Clear()
	c.Add("x", 1)
	contains(t, c, []string{
		"x:1",
	})
}
