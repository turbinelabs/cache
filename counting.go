package cache

// CountingCache represents a Cache specialized for counting strings.
type CountingCache interface {
	// Get retrieves the count for the given key.
	Get(key string) int

	// ForEach invokes f for each key/couunt in the cache. Callers should not depend
	// on deterministic ordering.
	ForEach(f func(key string, count int))

	// Inc increases the count for the given key by n, which may be negative. Return
	// its new value.
	Add(key string, n int) int

	// Inc increments the count for the given key by 1 and returns its new values.
	Inc(key string) int

	// Dec decrements the count for the given key by 1 and returns its new value.
	Dec(key string) int

	// Remove removes the key and returns its previous value.
	Remove(key string) int

	// Clear removes all keys.
	Clear()

	// Len returns the number of entries in the cache.
	Len() int
}

// NewNoopCountingCache returns a CountingCache implementation that counts nothing.
func NewNoopCountingCache() CountingCache {
	return &noopCountingCache{}
}

type noopCountingCache struct{}

func (*noopCountingCache) Get(_ string) int                { return 0 }
func (*noopCountingCache) ForEach(_ func(_ string, _ int)) {}
func (*noopCountingCache) Add(_ string, n int) int         { return n }
func (*noopCountingCache) Inc(_ string) int                { return 1 }
func (*noopCountingCache) Dec(_ string) int                { return 0 }
func (*noopCountingCache) Remove(_ string) int             { return 0 }
func (*noopCountingCache) Clear()                          {}
func (*noopCountingCache) Len() int                        { return 0 }
