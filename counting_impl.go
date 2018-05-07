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
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// NewCountingCache creates a CountingCache that tracks integer counts for at most
// size unique string keys. When the number of keys in the cache reaches size, the
// next new key added will randomly replace a key among the set of keys with the
// smallest count, even if that count is larger than the newly added key's
// value. Size must be at least 2.
func NewCountingCache(size int) (CountingCache, error) {
	if size < 2 {
		return nil, errors.New("minimum counting cache size is 2")
	}

	return &counting{
		size:   size,
		lookup: make(map[string]*count, size),
		counts: make(counts, 0, size),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

type counting struct {
	size   int
	lookup map[string]*count
	counts counts
	rng    *rand.Rand
	lock   sync.RWMutex
}

type count struct {
	n   int
	key *string
}

type counts []*count

func (c counts) Len() int           { return len(c) }
func (c counts) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c counts) Less(i, j int) bool { return c[i].n > c[j].n }

func (c counts) dropStart() int {
	min := c[len(c)-1].n
	return c.find(min)
}

func (c counts) find(n int) int {
	return sort.Search(len(c), func(idx int) bool { return c[idx].n <= n })
}

func (cc *counting) Get(key string) int {
	cc.lock.RLock()
	defer cc.lock.RUnlock()

	if count, ok := cc.lookup[key]; ok {
		return count.n
	}

	return 0
}

func (cc *counting) ForEach(f func(key string, count int)) {
	cc.lock.RLock()
	defer cc.lock.RUnlock()

	for k, count := range cc.lookup {
		f(k, count.n)
	}
}

func (cc *counting) Add(key string, n int) int {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if count, ok := cc.lookup[key]; ok {
		// Update existing entry.
		count.n += n
		sort.Sort(cc.counts)

		if count.n == 0 {
			cc.remove(cc.counts.find(0))
			return 0
		}

		return count.n
	}

	// Only add new entries if their count is non-zero.
	if n == 0 {
		return n
	}

	// Insure we stay under the maximum size.
	for len(cc.counts) >= cc.size {
		// Find the index of the first counts entry with the minimum count.
		idx := cc.counts.dropStart()

		// Pick one of the minimum count entries at random
		pick := idx + cc.rng.Intn(len(cc.counts)-idx)
		cc.remove(pick)
	}

	keyCopy := key
	count := &count{n: n, key: &keyCopy}
	cc.lookup[keyCopy] = count

	cc.counts = append(cc.counts, count)
	sort.Sort(cc.counts)

	return n
}

func (cc *counting) Inc(key string) int {
	return cc.Add(key, 1)
}

func (cc *counting) Dec(key string) int {
	return cc.Add(key, -1)
}

func (cc *counting) Remove(key string) int {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	if count, ok := cc.lookup[key]; ok {
		prev := count.n
		count.n = 0
		sort.Sort(cc.counts)
		cc.remove(cc.counts.find(0))
		return prev
	}

	return 0
}

func (cc *counting) Clear() {
	cc.lock.Lock()
	defer cc.lock.Unlock()

	cc.lookup = map[string]*count{}
	cc.counts = counts{}
}

func (cc *counting) Len() int {
	cc.lock.RLock()
	defer cc.lock.RUnlock()

	return len(cc.counts)
}

func (cc *counting) remove(idx int) {
	if idx < 0 || idx >= len(cc.counts) {
		return
	}

	c := cc.counts[idx]

	// delete it
	delete(cc.lookup, *(c.key))

	copy(cc.counts[idx:], cc.counts[idx+1:])
	cc.counts = cc.counts[0 : len(cc.counts)-1]
}
