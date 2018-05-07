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

// Package cache provides a simple Cache interface, with several
// concrete implementations.
package cache

//go:generate mockgen -source $GOFILE -destination mock_$GOFILE -package $GOPACKAGE --write_package_comment=false

// Cache represents a generic Cache. Specific Cache implementations
// may provide LRU or expiration semantics.
type Cache interface {
	// Retrieve an item from the cache. The second parameter
	// indicates whether an entry was found, allowing callers to
	// distinguish a nil value from "not found."
	Get(key interface{}) (interface{}, bool)

	// ForEach invokes f for each key/value in the cache. Callers should not depend
	// on deterministic ordering.
	ForEach(f func(key, value interface{}))

	// Add an item to the cache. Returns true if it replaced an
	// existing item. The value may be nil.
	Add(key, value interface{}) bool

	// Removes an item from the cache. Returns true if an item was
	// removed.
	Remove(key interface{}) bool

	// Remove all items from the cache.
	Clear()

	// Returns the number of items in the cache.
	Len() int
}

// NewNoopCache returns a Cache implementation that caches nothing.
func NewNoopCache() Cache {
	return &noopCache{}
}

type noopCache struct{}

func (*noopCache) Get(_ interface{}) (interface{}, bool) { return nil, false }
func (*noopCache) ForEach(_ func(_, _ interface{}))      {}
func (*noopCache) Add(_, _ interface{}) bool             { return false }
func (*noopCache) Remove(_ interface{}) bool             { return false }
func (*noopCache) Clear()                                {}
func (*noopCache) Len() int                              { return 0 }
